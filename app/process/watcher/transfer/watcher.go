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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
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
	transfers        service.Transfers
	client           client.MirrorNode
	accountID        hedera.AccountID
	typeMessage      string
	pollingInterval  time.Duration
	statusRepository repository.Status
	maxRetries       int
	startTimestamp   int64
	logger           *log.Entry
}

func NewWatcher(
	transfers service.Transfers,
	client client.MirrorNode,
	accountID hedera.AccountID,
	pollingInterval time.Duration,
	repository repository.Status,
	maxRetries int,
	startTimestamp int64,
) *Watcher {
	return &Watcher{
		transfers:        transfers,
		client:           client,
		accountID:        accountID,
		typeMessage:      process.CryptoTransferMessageType,
		pollingInterval:  pollingInterval,
		statusRepository: repository,
		maxRetries:       maxRetries,
		startTimestamp:   startTimestamp,
		logger:           config.GetLoggerFor(fmt.Sprintf("[%s] Transfer Watcher", accountID.String())),
	}
}

func (ctw Watcher) Watch(q *queue.Queue) {
	accountAddress := ctw.accountID.String()
	_, err := ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := ctw.statusRepository.CreateTimestamp(accountAddress, ctw.startTimestamp)
			if err != nil {
				ctw.logger.Fatalf("Failed to create Transfer Watcher Status timestamp. Error %s", err)
			}
			ctw.logger.Tracef("Cteated new Transfer Watcher status timestamp [%s]", timestamp.ToHumanReadable(ctw.startTimestamp))
		} else {
			ctw.logger.Fatalf("Failed to fetch last Transfer Watcher timestamp. Err: %s", err)
		}
	} else {
		ctw.updateStatusTimestamp(ctw.startTimestamp)
	}

	if !ctw.client.AccountExists(ctw.accountID) {
		ctw.logger.Errorf("Error incoming: Could not start monitoring account - Account not found.")
		return
	}

	go ctw.beginWatching(q)
	ctw.logger.Infof("Watching for Transfers after Timestamp [%s]", timestamp.ToHumanReadable(ctw.startTimestamp))
}

func (ctw Watcher) updateStatusTimestamp(ts int64) {
	err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID.String(), ts)
	if err != nil {
		ctw.logger.Fatalf("Failed to update Transfer Watcher Status timestamp. Error %s", err)
	}
	ctw.logger.Tracef("Updated Transfer Watcher timestamp to [%s]", timestamp.ToHumanReadable(ts))
}

func (ctw Watcher) beginWatching(q *queue.Queue) {
	milestoneTimestamp := ctw.startTimestamp
	for {
		transactions, e := ctw.client.GetAccountCreditTransactionsAfterTimestamp(ctw.accountID, milestoneTimestamp)
		if e != nil {
			ctw.logger.Errorf("Error incoming: Suddenly stopped monitoring account - [%s]", e)
			ctw.restart(q)
			return
		}

		ctw.logger.Tracef("Polling found [%d] Transactions", len(transactions.Transactions))
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

			ctw.updateStatusTimestamp(milestoneTimestamp)
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}

func (ctw Watcher) processTransaction(tx mirror_node.Transaction, q *queue.Queue) {
	ctw.logger.Infof("New Transaction with ID: [%s]", tx.TransactionID)
	amount, err := tx.GetIncomingAmountFor(ctw.accountID.String())
	if err != nil {
		ctw.logger.Errorf("Could not extract incoming amount for TX [%s]. Error: [%s]", tx.TransactionID, err)
		return
	}

	m, err := ctw.transfers.SanityCheckTransfer(tx)
	if err != nil {
		ctw.logger.Errorf("Sanity check for TX [%s] failed. Error: [%s]", tx.TransactionID, err)
		return
	}

	shouldExecuteEthTransaction := true
	//Check memo for the format {eth_address-0-0}
	if m.GasPriceGwei == "0" && m.TxReimbursementFee == "0" {
		shouldExecuteEthTransaction = false
	}

	transferMessage := encoding.NewTransferMessage(tx.TransactionID, m.EthereumAddress, amount, m.TxReimbursementFee, m.GasPriceGwei, shouldExecuteEthTransaction)
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
