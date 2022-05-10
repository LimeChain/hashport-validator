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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	validatorCfg "github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

func main() {
	cfg := new(config)
	err := validatorCfg.GetConfig(cfg, "config.yml")
	if err != nil {
		log.Fatal(err)
	}

	// Setup hedera, evm clients and nodes to be migrated
	hederaClient := setupHederaClient(cfg.Env)
	nodes := make([]*node, len(cfg.DBs))
	for i, db := range cfg.DBs {
		pgDb := persistence.NewPgConnector(db).Connect()

		nodes[i].db, err = pgDb.DB()
		nodes[i].repo = transfer.NewRepository(pgDb)

		if err != nil {
			log.Fatal(err)
		}
	}
	evmClients := make(map[uint64]client.Core)
	for k, v := range cfg.Evm {
		evmClients[k] = evm.NewClient(validatorCfg.Evm{
			NodeUrl: v,
		}, k)
	}

	migrator := newMigrator(nodes, hederaClient, evmClients)

	log.Infof("finna update %d nodes using %d evm clients", len(migrator.nodes), len(migrator.evmClients))
	err = migrator.migrate()
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	Env hedera.NetworkName      `yaml:"env"`
	DBs []validatorCfg.Database `yaml:"dbs"`
	Evm map[uint64]string       `yaml:"evm"`
}

type node struct {
	db   *sql.DB
	repo repository.Transfer
}

type migrator struct {
	nodes        []*node
	hederaClient *hedera.Client
	evmClients   map[uint64]client.Core
}

func newMigrator(nodes []*node, hederaClient *hedera.Client, evmClients map[uint64]client.Core) *migrator {
	return &migrator{nodes: nodes, hederaClient: hederaClient, evmClients: evmClients}
}

func (m *migrator) migrate() error {
	for i, node := range m.nodes {
		err := m.migrateNode(node, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) migrateNode(node *node, i int) error {
	c, err := node.db.Query(`select count (*) from transfers`)
	if err != nil {
		return err
	}

	var count int64
	if err := c.Scan(&count); err != nil {
		return err
	}

	tx, err := node.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	page := 1
	const perPage = 100
	totalUpdated := int64(0)
	for {
		if totalUpdated == count {
			if err := tx.Commit(); err != nil {
				return err
			}
			log.Infof("finished migrating %d transfers", totalUpdated)
			break
		}

		transfers, err := node.repo.PagedOffset(page, perPage)
		if err != nil {
			return err
		}

		for _, t := range transfers {
			err := m.fillTimestampAndOriginator(t)
			if err != nil {
				return err
			}
		}

		q := m.prepareQuery(transfers)
		res, err := tx.Exec(q)
		if err != nil {
			return err
		}
		r, err := res.RowsAffected()
		if err != nil {
			return err
		}

		totalUpdated += r
		log.Infof("updated %d rows", r)
	}

	return nil
}

func (m *migrator) hederaFields(transfer *entity.Transfer) error {
	txId, err := hedera.TransactionIdFromString(transfer.TransactionID)
	tx, err := hedera.NewTransactionRecordQuery().
		SetTransactionID(txId).
		Execute(m.hederaClient)
	if err != nil {
		return err
	}

	transfer.Timestamp = tx.ConsensusTimestamp
	transfer.Originator = tx.Receipt.AccountID.String()

	return nil
}

func (m *migrator) evmFields(transfer *entity.Transfer) error {
	parts := strings.SplitN(transfer.TransactionID, "-", 2)
	if len(parts) != 2 {
		return errors.New("transfer is not an evm transfer")
	}

	txHash := parts[0]
	c, err := m.evmClientFor(transfer.SourceChainID)
	if err != nil {
		return err
	}

	t1, cancel1 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel1()

	rx, err := c.TransactionReceipt(t1, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	t2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel2()

	block, err := c.BlockByNumber(t2, rx.BlockNumber)
	if err != nil {
		return err
	}

	uT := time.Unix(0, int64(block.Time()))
	transfer.Timestamp = uT

	t3, cancel3 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel3()

	tx, pending, err := c.TransactionByHash(t3, common.HexToHash(txHash))
	if err != nil {
		return err
	}
	if pending {
		return errors.New("pending transaction")
	}

	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()), nil)
	if err != nil {
		return err
	}

	transfer.Originator = msg.From().String()

	return nil
}

func (m *migrator) fillTimestampAndOriginator(transfer *entity.Transfer) error {
	if strings.Contains("0x", transfer.TransactionID) {
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

func (m *migrator) prepareQuery(transfers []*entity.Transfer) string {
	var q string
	for _, t := range transfers {
		q += fmt.Sprintf(
			`update transfers set timestamp = '%s', originator = '%s' where transaction_id = '%s';`,
			t.Timestamp.Format(time.RFC3339), t.Originator, t.TransactionID)
	}

	return q
}

func (m *migrator) name() {

}

func setupHederaClient(env hedera.NetworkName) *hedera.Client {
	var res *hedera.Client
	switch env {
	case hedera.NetworkNameTestnet:
		res = hedera.ClientForTestnet()
	case hedera.NetworkNameMainnet:
		res = hedera.ClientForMainnet()
	}
	return res
}
