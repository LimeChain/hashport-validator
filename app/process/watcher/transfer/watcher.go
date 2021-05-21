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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Watcher struct {
	transfers        service.Transfers
	client           client.MirrorNode
	accountID        hedera.AccountID
	pollingInterval  time.Duration
	statusRepository repository.Status
	startTimestamp   int64
	logger           *log.Entry
	contractService  service.Contracts
}

func NewWatcher(
	transfers service.Transfers,
	client client.MirrorNode,
	accountID string,
	pollingInterval time.Duration,
	repository repository.Status,
	startTimestamp int64,
	contractService service.Contracts,
) *Watcher {
	id, err := hedera.AccountIDFromString(accountID)
	if err != nil {
		log.Fatalf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", accountID, err)
	}

	return &Watcher{
		transfers:        transfers,
		client:           client,
		accountID:        id,
		pollingInterval:  pollingInterval,
		statusRepository: repository,
		startTimestamp:   startTimestamp,
		logger:           config.GetLoggerFor(fmt.Sprintf("[%s] Transfer Watcher", accountID)),
		contractService:  contractService,
	}
}

func (ctw Watcher) Watch(q *pair.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		ctw.logger.Errorf("Could not start monitoring account [%s] - Account not found.", ctw.accountID.String())
		return
	}

	account := ctw.accountID.String()
	_, err := ctw.statusRepository.GetLastFetchedTimestamp(account)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := ctw.statusRepository.CreateTimestamp(account, ctw.startTimestamp)
			if err != nil {
				ctw.logger.Fatalf("Failed to create Transfer Watcher timestamp. Error [%s]", err)
			}
			ctw.logger.Tracef("Created new Transfer Watcher timestamp [%s]", timestamp.ToHumanReadable(ctw.startTimestamp))
		} else {
			ctw.logger.Fatalf("Failed to fetch last Transfer Watcher timestamp. Error: [%s]", err)
		}
	} else {
		ctw.updateStatusTimestamp(ctw.startTimestamp)
	}

	go ctw.beginWatching(q)
	ctw.logger.Infof("Watching for Transfers after Timestamp [%s]", timestamp.ToHumanReadable(ctw.startTimestamp))
}

func (ctw Watcher) updateStatusTimestamp(ts int64) {
	err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID.String(), ts)
	if err != nil {
		ctw.logger.Fatalf("Failed to update Transfer Watcher Status timestamp. Error [%s]", err)
	}
	ctw.logger.Tracef("Updated Transfer Watcher timestamp to [%s]", timestamp.ToHumanReadable(ts))
}

func (ctw Watcher) beginWatching(q *pair.Queue) {
	milestoneTimestamp, err := ctw.statusRepository.GetLastFetchedTimestamp(ctw.accountID.String())
	if err != nil {
		ctw.logger.Fatalf("Failed to retrieve Transfer Watcher Status timestamp. Error [%s]", err)
	}

	for {
		transactions, e := ctw.client.GetAccountCreditTransactionsAfterTimestamp(ctw.accountID, milestoneTimestamp)
		if e != nil {
			ctw.logger.Errorf("Suddenly stopped monitoring account - [%s]", e)
			go ctw.beginWatching(q)
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
				ctw.logger.Errorf("Unable to parse latest transfer timestamp. Error - [%s].", err)
				continue
			}

			ctw.updateStatusTimestamp(milestoneTimestamp)
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}

func (ctw Watcher) processTransaction(tx mirror_node.Transaction, q *pair.Queue) {
	ctw.logger.Infof("New Transaction with ID: [%s]", tx.TransactionID)
	amount, nativeAsset, err := tx.GetIncomingTransfer(ctw.accountID.String())
	if err != nil {
		ctw.logger.Errorf("[%s] - Could not extract incoming transfer. Error: [%s]", tx.TransactionID, err)
		return
	}

	wrappedAsset, err := ctw.contractService.ToWrapped(nativeAsset)
	if err != nil {
		ctw.logger.Errorf("[%s] - Could not parse native asset [%s] - Error: [%s]", tx.TransactionID, nativeAsset, err)
		return
	}

	ethAddress, err := ctw.transfers.SanityCheckTransfer(tx)
	if err != nil {
		ctw.logger.Errorf("[%s] - Sanity check failed. Error: [%s]", tx.TransactionID, err)
		return
	}

	transferMessage := transfer.New(tx.TransactionID, ethAddress, nativeAsset, wrappedAsset, amount, ctw.contractService.Address().String())
	q.Push(&pair.Message{Payload: transferMessage})
}

//func (ctw *Watcher) restart(q *pair.Queue) {
//	if ctw.maxRetries > 0 {
//		ctw.maxRetries--
//		ctw.logger.Infof("Watcher is trying to reconnect")
//		time.Sleep(5 * time.Second)
//		go ctw.beginWatching(q)
//		return
//	}
//	ctw.logger.Errorf("Watcher failed: [Too many retries]")
//}
