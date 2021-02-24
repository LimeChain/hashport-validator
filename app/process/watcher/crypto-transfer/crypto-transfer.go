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
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CryptoTransferWatcher struct {
	client           *hederaClient.HederaMirrorClient
	accountID        hedera.AccountID
	typeMessage      string
	pollingInterval  time.Duration
	statusRepository repositories.StatusRepository
	maxRetries       int
	startTimestamp   int64
	started          bool
	logger           *log.Entry
}

func NewCryptoTransferWatcher(
	client *hederaClient.HederaMirrorClient,
	accountID hedera.AccountID,
	pollingInterval time.Duration,
	repository repositories.StatusRepository,
	maxRetries int,
	startTimestamp int64,
) *CryptoTransferWatcher {
	return &CryptoTransferWatcher{
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

func (ctw CryptoTransferWatcher) Watch(q *queue.Queue) {
	go ctw.beginWatching(q)
}

func (ctw CryptoTransferWatcher) getTimestamp(q *queue.Queue) int64 {
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

func (ctw CryptoTransferWatcher) beginWatching(q *queue.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		ctw.logger.Errorf("Error incoming: Could not start monitoring account - Account not found.")
		return
	}
	ctw.logger.Info("Starting watcher")

	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == 0 {
		ctw.logger.Fatalf("Could not start watcher - Could not generate a milestone timestamp.")
	}

	ctw.logger.Infof("Started watcher")
	for {
		transactions, e := ctw.client.GetSuccessfulAccountCreditTransactionsAfterDate(ctw.accountID, milestoneTimestamp)
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

func (ctw CryptoTransferWatcher) processTransaction(tx transaction.HederaTransaction, q *queue.Queue) {
	ctw.logger.Infof("New Transaction with ID: [%s]", tx.TransactionID)

	amount := ExtractAmount(tx, ctw.accountID)

	decodedMemo, e := DecodeMemo(tx.MemoBase64)
	if e != nil {
		ctw.logger.Errorf("Could not parse transaction memo for Transaction with ID [%s] - Error: [%s]", tx.TransactionID, e)
		return
	}

	_, e = ctw.client.GetStateProof(tx.TransactionID)
	if e != nil {
		ctw.logger.Errorf("Could not GET state proof, TransactionID [%s]. Error [%s]", tx.TransactionID, e)
		return
	}

	// TODO: Uncomment after support for V5 record and signature files has been added
	//verified, e := stateproof.Verify(tx.TransactionID, stateProof)
	//if e != nil {
	//	ctw.logger.Errorf("Error while trying to verify state proof for TransactionID [%s]. Error [%s]", tx.TransactionID, e)
	//	return
	//}
	//
	//if !verified {
	//	ctw.logger.Errorf("Failed to verify state proof for TransactionID [%s]", tx.TransactionID)
	//	return
	//}

	information := &protomsg.CryptoTransferMessage{
		TransactionId: tx.TransactionID,
		EthAddress:    decodedMemo.EthAddress,
		Amount:        strconv.Itoa(int(amount)),
		Fee:           decodedMemo.Fee,
		GasPriceGwei:  decodedMemo.GasPriceGwei,
	}
	publisher.Publish(information, ctw.typeMessage, ctw.accountID, q)
}

func (ctw CryptoTransferWatcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		ctw.logger.Infof("Watcher is trying to reconnect")
		go ctw.Watch(q)
		return
	}
	ctw.logger.Errorf("Watcher failed: [Too many retries]")
}

func DecodeMemo(memo string) (*MemoInfo, error) {
	wholeMemoCheck := regexp.MustCompile("^0x([A-Fa-f0-9]){40}-[1-9][0-9]*-[1-9][0-9]*$")

	decodedMemo, e := base64.StdEncoding.DecodeString(memo)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse transaction memo: [%s]", e))
	}

	if len(decodedMemo) < 46 || !wholeMemoCheck.MatchString(string(decodedMemo)) {
		return nil, errors.New(fmt.Sprintf("Transaction memo provides invalid or insufficient data - Memo: [%s]", string(decodedMemo)))
	}

	memoSplit := strings.Split(string(decodedMemo), "-")
	ethAddress := memoSplit[0]
	fee := memoSplit[1]
	gasPriceGwei := memoSplit[2]

	return &MemoInfo{EthAddress: ethAddress, Fee: fee, GasPriceGwei: gasPriceGwei}, nil
}

func ExtractAmount(tx transaction.HederaTransaction, accountID hedera.AccountID) int64 {
	var amount int64
	for _, tr := range tx.Transfers {
		if tr.Account == accountID.String() {
			amount = tr.Amount
			break
		}
	}
	return amount
}

type MemoInfo struct {
	EthAddress   string
	Fee          string
	GasPriceGwei string
}
