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
	evmhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/evm"
	"strconv"
	"strings"
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"

	"github.com/limechain/hedera-eth-bridge-validator/config/parser"

	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"

	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"

	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"

	"github.com/ethereum/go-ethereum/common"
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
		pgDb := persistence.NewPgConnector(validatorCfg.Database(db)).Connect()
		db, err := pgDb.DB()
		if err != nil {
			log.Fatal(err)
		}

		nodes[i] = &node{
			db: db,
		}
	}
	evmClients := make(map[uint64]client.EVM)
	for k, v := range cfg.Evm {
		evmClients[k] = evm.NewClient(validatorCfg.Evm{
			NodeUrl:            v,
			BlockConfirmations: 5,
		}, k)
	}

	mnc := mirror_node.NewClient(validatorCfg.MirrorNode{
		ClientAddress:   cfg.Hedera.MirrorNode.ClientAddress,
		ApiAddress:      cfg.Hedera.MirrorNode.ApiAddress,
		PollingInterval: cfg.Hedera.MirrorNode.PollingInterval,
	})

	migrator := newMigrator(nodes, evmClients, mnc)

	log.Infof("updating %d nodes using %d evm clients", len(migrator.nodes), len(migrator.evmClients))
	err = migrator.migrate()
	if err != nil {
		log.Fatal(err)
	}
}

type tempTransfer struct {
	TransactionID string
	SourceChainID uint64
	Timestamp     entity.NanoTime
	Originator    sql.NullString
}

type hederaCfg struct {
	MirrorNode parser.MirrorNode `yaml:"mirror_node"`
}

type config struct {
	Hedera hederaCfg         `yaml:"hedera"`
	DBs    []parser.Database `yaml:"dbs"`
	Evm    map[uint64]string `yaml:"evm"`
}

type node struct {
	db *sql.DB
}

type migrator struct {
	nodes        []*node
	evmClients   map[uint64]client.EVM
	hederaClient client.MirrorNode
}

func newMigrator(nodes []*node, evmClients map[uint64]client.EVM, hederaClient client.MirrorNode) *migrator {
	return &migrator{nodes: nodes, evmClients: evmClients, hederaClient: hederaClient}
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
	const perPage = 50
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
			if t.Originator.Valid && !t.Timestamp.IsZero() {
				continue
			}

			err := m.fillTimestampAndOriginator(t)
			if err != nil {
				return err
			}
			log.Infof("[DB: %d]: updated transfer [%s] with timestamp [%d] and originator [%s]",
				i, t.TransactionID, t.Timestamp.Time.UnixNano(), t.Originator.String)
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
	tx, err := m.hederaClient.GetTransaction(tr.TransactionID)
	if err != nil {
		return err
	}

	tNano, err := tx.GetLatestTxnConsensusTime()
	if err != nil {
		return err
	}

	o := hederahelper.OriginatorFromTxId(tr.TransactionID)

	if err != nil {
		return err
	}

	tr.Timestamp = entity.NanoTime{Time: timestampHelper.FromNanos(tNano)}
	tr.Originator = sql.NullString{String: o, Valid: true}

	return nil
}

func (m *migrator) evmFields(transfer *tempTransfer) error {
	parts := strings.Split(transfer.TransactionID, "-")
	if len(parts) < 1 {
		return errors.New("transfer is not an evm transfer")
	}

	txHash := parts[0]
	c, err := m.evmClientFor(transfer.SourceChainID)
	if err != nil {
		return err
	}

	rx, err := c.WaitForTransactionReceipt(common.HexToHash(txHash))
	if err != nil {
		return err
	}

	t2, cancel2 := context.WithTimeout(context.Background(), evmCallTimeout)
	defer cancel2()

	block, err := c.(client.Core).BlockByNumber(t2, rx.BlockNumber)
	if err != nil {
		return err
	}

	uT := time.Unix(int64(block.Time()), 0)
	transfer.Timestamp = entity.NanoTime{Time: uT.UTC()}

	tx, err := c.RetryTransactionByHash(common.HexToHash(txHash))
	if err != nil {
		return err
	}

	originator, err := evmhelper.OriginatorFromTx(tx)
	if err != nil {
		return err
	}

	transfer.Originator = sql.NullString{String: originator}

	return nil
}

func (m *migrator) fillTimestampAndOriginator(transfer *tempTransfer) error {
	if strings.Contains(transfer.TransactionID, "0x") {
		return m.evmFields(transfer)
	}
	return m.hederaFields(transfer)
}

func (m *migrator) evmClientFor(networkId uint64) (client.EVM, error) {
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
			"update transfers set \"timestamp\" = %d, \"originator\" = '%s' where \"transaction_id\" = '%s';\n",
			t.Timestamp.Time.UTC().UnixNano(), t.Originator.String, t.TransactionID)
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
