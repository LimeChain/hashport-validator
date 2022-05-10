/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	validatorCfg "github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

const evmCallTimeout = 5 * time.Second

func main() {
	cfg := new(config)
	err := validatorCfg.GetConfig(cfg, "scripts/migration/add_timestamp_and_originator_to_transfers/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	// Setup hedera, evm clients and nodes to be migrated
	nodes := make([]*node, len(cfg.DBs))
	for i, db := range cfg.DBs {
		pgDb := persistence.NewPgConnector(db).Connect()
		db, err := pgDb.DB()
		if err != nil {
			log.Fatal(err)
		}

		nodes[i] = &node{
			db: db,
		}
	}
	evmClients := make(map[uint64]client.Core)
	for k, v := range cfg.Evm {
		evmClients[k] = evm.NewClient(validatorCfg.Evm{
			NodeUrl:            v,
			BlockConfirmations: 5,
		}, k)
	}

	migrator := newMigrator(nodes, evmClients)

	log.Infof("finna update %d nodes using %d evm clients", len(migrator.nodes), len(migrator.evmClients))
	err = migrator.migrate()
	if err != nil {
		log.Fatal(err)
	}
}

type tempTransfer struct {
	TransactionID string
	SourceChainID uint64
	Timestamp     sql.NullTime
	Originator    sql.NullString
}

type hederaCfg struct {
	Env        hedera.NetworkName      `yaml:"env"`
	AccountId  string                  `yaml:"account_id"`
	PrivateKey string                  `yaml:"private_key"`
	MirrorNode validatorCfg.MirrorNode `yaml:"mirror_node"`
}

type config struct {
	Hedera hederaCfg               `yaml:"hedera"`
	DBs    []validatorCfg.Database `yaml:"dbs"`
	Evm    map[uint64]string       `yaml:"evm"`
}

type node struct {
	db *sql.DB
}

type migrator struct {
	nodes      []*node
	evmClients map[uint64]client.Core
}

func newMigrator(nodes []*node, evmClients map[uint64]client.Core) *migrator {
	return &migrator{nodes: nodes, evmClients: evmClients}
}

func (m *migrator) migrate() error {
	for i, _ := range m.nodes {
		err := m.migrateNode(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) migrateNode(i int) error {
	migratedNode := m.nodes[i]
	c, err := migratedNode.db.Query(`select count (*) from transfers`)
	if err != nil {
		return err
	}

	var count int
	if c.Next() {
		if err := c.Scan(&count); err != nil {
			return err
		}
	} else {
		return errors.New("no transfers count")
	}
	c.Close()

	tx, err := migratedNode.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	page := 1
	const perPage = 5
	totalUpdated := 0
	for {
		if totalUpdated >= count {
			if err := tx.Commit(); err != nil {
				return err
			}
			log.Infof("finished migrating %d transfers", totalUpdated)
			break
		}

		transfers, err := migratedNode.pagedTransfers(page, perPage)
		if err != nil {
			return err
		}

		for _, t := range transfers {
			if t.Originator.Valid && t.Timestamp.Valid {
				continue
			}

			err := m.fillTimestampAndOriginator(t)
			if err != nil {
				return err
			}
		}

		qs := m.prepareQuery(transfers)
		if err != nil {
			return err
		}

		for _, q := range qs {
			if _, err := tx.Exec(q); err != nil {
				return err
			}
		}
		totalUpdated += len(qs)
		log.Infof("updated %d/%d transfers", totalUpdated, count)

		page++
	}

	return nil
}

func (m *migrator) hederaFields(tr *tempTransfer) error {
	o := hederahelper.OriginatorFromTxId(tr.TransactionID)
	t, err := hederahelper.TimestampFromTxId(tr.TransactionID)
	if err != nil {
		return err
	}

	tr.Timestamp = sql.NullTime{Time: t.UTC(), Valid: true}
	tr.Originator = sql.NullString{String: o, Valid: true}

	return nil
}

func (m *migrator) evmFields(transfer *tempTransfer) error {
	parts := strings.SplitN(transfer.TransactionID, "-", 2)
	if len(parts) != 2 {
		return errors.New("transfer is not an evm transfer")
	}

	txHash := parts[0]
	c, err := m.evmClientFor(transfer.SourceChainID)
	if err != nil {
		return err
	}

	t1, cancel1 := context.WithTimeout(context.Background(), evmCallTimeout)
	defer cancel1()

	rx, err := c.TransactionReceipt(t1, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	t2, cancel2 := context.WithTimeout(context.Background(), evmCallTimeout)
	defer cancel2()

	block, err := c.BlockByNumber(t2, rx.BlockNumber)
	if err != nil {
		return err
	}

	uT := time.Unix(int64(block.Time()), 0)
	transfer.Timestamp = sql.NullTime{Time: uT.UTC()}

	t3, cancel3 := context.WithTimeout(context.Background(), evmCallTimeout)
	defer cancel3()

	tx, pending, err := c.TransactionByHash(t3, common.HexToHash(txHash))
	if err != nil {
		return err
	}
	if pending {
		return errors.New("pending transaction")
	}

	msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), nil)
	if err != nil {
		return err
	}

	transfer.Originator = sql.NullString{String: msg.From().String()}

	return nil
}

func (m *migrator) fillTimestampAndOriginator(transfer *tempTransfer) error {
	if strings.Contains(transfer.TransactionID, "0x") {
		return m.evmFields(transfer)
	}
	return m.hederaFields(transfer)
}

func (m *migrator) evmClientFor(networkId uint64) (client.Core, error) {
	c, ok := m.evmClients[networkId]
	if !ok {
		return nil, errors.New("no evm client for network id " + strconv.FormatUint(networkId, 10))
	}
	return c, nil
}

func (m *migrator) prepareQuery(transfers []*tempTransfer) []string {
	qs := make([]string, 0, len(transfers))
	for _, t := range transfers {
		q := fmt.Sprintf(
			"update transfers set \"timestamp\" = '%s', \"originator\" = '%s' where \"transaction_id\" = '%s';\n",
			t.Timestamp.Time.UTC().Format(time.RFC3339), t.Originator.String, t.TransactionID)
		qs = append(qs, q)
	}

	return qs
}

func (n *node) pagedTransfers(page, perPage int) ([]*tempTransfer, error) {
	offset := (page - 1) * perPage
	cur, err := n.db.Query(`
		select "transaction_id", "source_chain_id", "timestamp", "originator" from "transfers"
		offset $1 limit $2`,
		offset, perPage)
	if err != nil {
		return nil, err
	}

	res := make([]*tempTransfer, 0, perPage)
	for cur.Next() {
		var t tempTransfer
		if err := cur.Scan(&t.TransactionID, &t.SourceChainID, &t.Timestamp, &t.Originator); err != nil {
			return nil, err
		}
		res = append(res, &t)
	}

	return res, nil
}
