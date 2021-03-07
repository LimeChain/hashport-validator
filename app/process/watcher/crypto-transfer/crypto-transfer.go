/*
 * Copyright 2021 LimeChain Ltd.
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

package cryptotransfer

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hederaAPIModel "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Watcher struct {
	bridgeService    services.Bridge
	client           clients.MirrorNode
	accountID        hedera.AccountID
	typeMessage      string
	pollingInterval  time.Duration
	statusRepository repositories.Status
	maxRetries       int
	startTimestamp   int64
	started          bool
	logger           *log.Entry
}

func NewWatcher(
	bridgeService services.Bridge,
	client clients.MirrorNode,
	accountID hedera.AccountID,
	pollingInterval time.Duration,
	repository repositories.Status,
	maxRetries int,
	startTimestamp int64,
) *Watcher {
	return &Watcher{
		bridgeService:    bridgeService,
		client:           client,
		accountID:        accountID,
		typeMessage:      process.CryptoTransferMessageType,
		pollingInterval:  pollingInterval,
		statusRepository: repository,
		maxRetries:       maxRetries,
		startTimestamp:   startTimestamp,
		started:          false,
		logger:           config.GetLoggerFor(fmt.Sprintf("Account [%s] Transfer Watcher", accountID.String())),
	}
}

func (ctw Watcher) Watch(q *queue.Queue) {
	go ctw.beginWatching(q)
}

func (ctw Watcher) getTimestamp(q *queue.Queue) int64 {
	accountAddress := ctw.accountID.String()
	milestoneTimestamp := ctw.startTimestamp
	var err error

	if !ctw.started {
		if milestoneTimestamp > 0 {
			return milestoneTimestamp
		}

		ctw.logger.Warnf("[%s] Starting Timestamp was empty, proceeding to get [timestamp] from database.", accountAddress)
		milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
		if err == nil && milestoneTimestamp > 0 {
			return milestoneTimestamp
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctw.logger.Fatal(err)
		}

		ctw.logger.Warnf("[%s] Database Timestamp was empty, proceeding with [timestamp] from current moment.", accountAddress)
		milestoneTimestamp = time.Now().UnixNano()
		e := ctw.statusRepository.CreateTimestamp(accountAddress, milestoneTimestamp)
		if e != nil {
			ctw.logger.Fatal(e)
		}
		return milestoneTimestamp
	}

	milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
	if err != nil {
		ctw.logger.Warnf("[%s] Database Timestamp was empty. Restarting. Error - [%s]", accountAddress, err)
		ctw.started = false
		ctw.restart(q)
	}

	return milestoneTimestamp
}

func (ctw Watcher) beginWatching(q *queue.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		ctw.logger.Errorf("Error incoming: Could not start monitoring account - Account not found.")
		return
	}

	// TODO start from `now` (from recovery)
	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == 0 {
		ctw.logger.Fatalf("Could not start watcher - Could not generate a milestone timestamp.")
	}

	for {
		transactions, e := ctw.client.GetAccountCreditTransactionsAfterTimestamp(ctw.accountID, milestoneTimestamp)
		if e != nil {
			ctw.logger.Errorf("Error incoming: Suddenly stopped monitoring account - [%s]", e)
			ctw.restart(q)
			return
		}

		if len(transactions.Transactions) > 0 {
			for _, tx := range transactions.Transactions {
				go ctw.processTransaction(tx, q)
			}
			var err error
			milestoneTimestamp, err = timestamp.FromString(transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp)
			if err != nil {
				ctw.logger.Errorf("Watcher [%s] - Unable to parse latest transaction timestamp. Error - [%s].", ctw.accountID.String(), err)
				continue
			}
		}

		err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID.String(), milestoneTimestamp)
		if err != nil {
			ctw.logger.Errorf("Error incoming: Failed to update last fetched timestamp - [%s]", e)
			return
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}

func (ctw Watcher) processTransaction(tx hederaAPIModel.Transaction, q *queue.Queue) {
	ctw.logger.Infof("New Transaction with ID: [%s]", tx.TransactionID)
	amount, err := tx.GetIncomingAmountFor(ctw.accountID.String())
	if err != nil {
		ctw.logger.Errorf("Could not extract incoming amount for TX [%s]. Error: [%s]", tx.TransactionID, err)
		return
	}

	m, err := ctw.bridgeService.SanityCheck(tx)
	if err != nil {
		ctw.logger.Errorf("Sanity check for TX [%s] failed. Error: [%s]", tx.TransactionID, err)
		return
	}

	transferMessage := encoding.NewTransferMessage(tx.TransactionID, m.EthereumAddress, amount, m.TxReimbursementFee, m.GasPriceGwei)
	publisher.Publish(transferMessage, ctw.typeMessage, ctw.accountID, q)
}

func (ctw Watcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		ctw.logger.Infof("Watcher is trying to reconnect")
		go ctw.Watch(q)
		return
	}
	ctw.logger.Errorf("Watcher failed: [Too many retries]")
}
