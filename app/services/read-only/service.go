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

package read_only

import (
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirrorNodeTransaction "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	mirrorNode         client.MirrorNode
	transferRepository repository.Transfer
	pollingInterval    time.Duration
	logger             *log.Entry
}

const CryptoTransfer = "CRYPTOTRANSFER"
const CryptoApproveAllowance = "CRYPTOAPPROVEALLOWANCE"

func New(
	mirrorNode client.MirrorNode,
	transferRepository repository.Transfer,
	pollingInterval time.Duration) *Service {
	return &Service{
		mirrorNode:         mirrorNode,
		transferRepository: transferRepository,
		pollingInterval:    pollingInterval,
		logger:             config.GetLoggerFor("Read-only Transfer Fetcher"),
	}
}

func (s Service) FindAssetTransfer(
	transferID string,
	asset string,
	expectedTransfers []model.Hedera,
	fetch func() (*mirrorNodeTransaction.Response, error),
	save func(transactionID, scheduleID, status string) error) {
	for {
		response, err := fetch()
		if err != nil {
			s.logger.Errorf("[%s] - Failed to get transactions after timestamp. Error: [%s]", transferID, err)
			continue
		}

		finished := false
		for _, transaction := range response.Transactions {
			isFound := false
			scheduledTx, err := s.mirrorNode.GetScheduledTransaction(transaction.TransactionID)
			if err != nil {
				s.logger.Errorf("[%s] - Failed to retrieve scheduled transaction [%s]. Error: [%s]", transferID, transaction.TransactionID, err)
				continue
			}
			for _, tx := range scheduledTx.Transactions {
				if tx.Result == hedera.StatusSuccess.String() {
					scheduleID, err := s.mirrorNode.GetSchedule(tx.EntityId)
					if err != nil {
						s.logger.Errorf("[%s] - Failed to get scheduled entity [%s]. Error: [%s]", transferID, tx.EntityId, err)
						break
					}
					if scheduleID.Memo == transferID {
						isFound = true
					}
				}
				if isFound && transfersAreFound(expectedTransfers, asset, transaction) {
					s.logger.Infof("[%s] - Found a corresponding transaction [%s], ScheduleID [%s].", transferID, transaction.TransactionID, tx.EntityId)
					finished = true
					isSuccessful := transaction.Result == hedera.StatusSuccess.String()
					txStatus := status.Completed
					if !isSuccessful {
						txStatus = status.Failed
					}

					err := save(transaction.TransactionID, tx.EntityId, txStatus)
					if err != nil {
						s.logger.Errorf("[%s] - Failed to save entity [%s]. Error: [%s]", transferID, tx.EntityId, err)
						break
					}

					if isSuccessful {
						err = s.transferRepository.UpdateStatusCompleted(transferID)
					} else {
						err = s.transferRepository.UpdateStatusFailed(transferID)
					}
					if err != nil {
						s.logger.Errorf("[%s] - Failed to update status. Error: [%s]", transferID, err)
						break
					}
					break
				}
			}
		}
		if finished {
			break
		}
		s.logger.Tracef("[%s] - No asset transfers found.", transferID)

		time.Sleep(s.pollingInterval * time.Second)
	}
}

func (s Service) FindScheduledNftAllowanceApprove(
	t *payload.Transfer,
	sender hedera.AccountID,
	save func(transactionID, scheduleID, status string) error) {
	transferID := t.TransactionId
	for {
		txs, err := s.mirrorNode.GetTransactionsAfterTimestamp(sender, t.Timestamp.UnixNano(), CryptoApproveAllowance)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to get transactions after timestamp. Error: [%s]", transferID, err)
			continue
		}

		finished := false
		for _, tx := range txs {
			scheduledTx, err := s.mirrorNode.GetScheduledTransaction(tx.TransactionID)
			if err != nil {
				s.logger.Errorf("[%s] - Failed to retrieve scheduled transaction [%s]. Error: [%s]", transferID, tx.TransactionID, err)
				continue
			}

			found := false
			for _, tx := range scheduledTx.Transactions {
				if tx.Result == hedera.StatusSuccess.String() {
					schedule, err := s.mirrorNode.GetSchedule(tx.EntityId)
					if err != nil {
						s.logger.Errorf("[%s] - Failed to get scheduled entity [%s]. Error: [%s]", transferID, schedule, err)
						break
					}
					if schedule.Memo == transferID {
						found = true
					}
				}

				if found {
					s.logger.Infof("[%s] - Found a corresponding transaction [%s], ScheduleID [%s].", transferID, tx.TransactionID, tx.EntityId)
					finished = true
					txStatus := status.Completed
					success := tx.Result == hedera.StatusSuccess.String()
					if !success {
						txStatus = status.Failed
					}

					err := save(tx.TransactionID, tx.EntityId, txStatus)
					if err != nil {
						s.logger.Errorf("[%s] - Failed to save entity [%s]. Error: [%s]", transferID, tx.EntityId, err)
						break
					}
					break
				}
			}
		}
		if finished {
			break
		}

		s.logger.Tracef("[%s] - No transfers found.", transferID)
		time.Sleep(s.pollingInterval * time.Second)
	}
}

func (s Service) FindNftTransfer(
	transferID string, tokenID string, serialNum int64, sender string, receiver string,
	save func(transactionID, scheduleID, status string) error) {
	for {
		response, err := s.mirrorNode.GetNftTransactions(tokenID, serialNum)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to get nft transactions after timestamp. Error: [%s]", transferID, err)
			continue
		}

		finished := false
		for _, transaction := range response.Transactions {
			if transaction.Type == CryptoTransfer &&
				transaction.ReceiverAccountID == receiver &&
				transaction.SenderAccountID == sender {

				scheduledTx, err := s.mirrorNode.GetScheduledTransaction(transaction.TransactionID)
				if err != nil {
					s.logger.Errorf("[%s] - Failed to retrieve scheduled transaction [%s]. Error: [%s]", transferID, transaction.TransactionID, err)
					continue
				}
				for _, tx := range scheduledTx.Transactions {
					if tx.Result == hedera.StatusSuccess.String() {
						scheduleID, err := s.mirrorNode.GetSchedule(tx.EntityId)
						if err != nil {
							s.logger.Errorf("[%s] - Failed to get scheduled entity [%s]. Error: [%s]", transferID, scheduleID, err)
							break
						}
						if scheduleID.Memo == transferID {
							s.logger.Infof("[%s] - Found a corresponding transaction [%s], ScheduleID [%s].", transferID, transaction.TransactionID, tx.EntityId)
							finished = true
							txStatus := status.Completed

							err := save(transaction.TransactionID, tx.EntityId, txStatus)
							if err != nil {
								s.logger.Errorf("[%s] - Failed to save entity [%s]. Error: [%s]", transferID, tx.EntityId, err)
								break
							}

							break
						}
					}
				}
			}
		}
		if finished {
			break
		}

		time.Sleep(s.pollingInterval * time.Second)
	}
}

func (s Service) FindTransfer(
	transferID string,
	fetch func() (*mirrorNodeTransaction.Response, error),
	save func(transactionID, scheduleID, status string) error) {
	for {
		response, err := fetch()
		if err != nil {
			s.logger.Errorf("[%s] - Failed to get transactions after timestamp. Error: [%s]", transferID, err)
			continue
		}

		finished := false
		for _, transaction := range response.Transactions {
			isFound := false
			scheduledTx, err := s.mirrorNode.GetScheduledTransaction(transaction.TransactionID)
			if err != nil {
				s.logger.Errorf("[%s] - Failed to retrieve scheduled transaction [%s]. Error: [%s]", transferID, transaction.TransactionID, err)
				continue
			}
			for _, tx := range scheduledTx.Transactions {
				if tx.Result == hedera.StatusSuccess.String() {
					scheduleID, err := s.mirrorNode.GetSchedule(tx.EntityId)
					if err != nil {
						s.logger.Errorf("[%s] - Failed to get scheduled entity [%s]. Error: [%s]", transferID, scheduleID, err)
						break
					}
					if scheduleID.Memo == transferID {
						isFound = true
					}
				}
				if isFound {
					s.logger.Infof("[%s] - Found a corresponding transaction [%s], ScheduleID [%s].", transferID, transaction.TransactionID, tx.EntityId)
					finished = true
					isSuccessful := transaction.Result == hedera.StatusSuccess.String()
					txStatus := status.Completed
					if !isSuccessful {
						txStatus = status.Failed
					}

					err := save(transaction.TransactionID, tx.EntityId, txStatus)
					if err != nil {
						s.logger.Errorf("[%s] - Failed to save entity [%s]. Error: [%s]", transferID, tx.EntityId, err)
						break
					}

					if isSuccessful {
						err = s.transferRepository.UpdateStatusCompleted(transferID)
					} else {
						err = s.transferRepository.UpdateStatusFailed(transferID)
					}
					if err != nil {
						s.logger.Errorf("[%s] - Failed to update status. Error: [%s]", transferID, err)
						break
					}
					break
				}
			}
		}
		if finished {
			break
		}

		s.logger.Tracef("[%s] - No transfers found.", transferID)
		time.Sleep(s.pollingInterval * time.Second)
	}
}

func transfersAreFound(expectedTransfers []model.Hedera, asset string, transaction mirrorNodeTransaction.Transaction) bool {
	for _, expectedTransfer := range expectedTransfers {
		found := false
		if asset == constants.Hbar {
			for _, transfer := range transaction.Transfers {
				if expectedTransfer.AccountID.String() == transfer.Account &&
					expectedTransfer.Amount == transfer.Amount {
					found = true
					break
				}
			}
		} else {
			for _, transfer := range transaction.TokenTransfers {
				if expectedTransfer.AccountID.String() == transfer.Account &&
					expectedTransfer.Amount == transfer.Amount &&
					asset == transfer.Token {
					found = true
					break
				}
			}
		}

		if !found {
			return false
		}
	}

	return true
}
