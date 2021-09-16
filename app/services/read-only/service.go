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

package read_only

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	mirrorNode         client.MirrorNode
	transferRepository repository.Transfer
	logger             *log.Entry
}

func New(
	mirrorNode client.MirrorNode,
	transferRepository repository.Transfer) *Service {
	return &Service{
		mirrorNode:         mirrorNode,
		transferRepository: transferRepository,
		logger:             config.GetLoggerFor("Read-only Transfer Fetcher"),
	}
}

func (s Service) FindTransfer(
	transferID string,
	fetch func() (*mirror_node.Response, error),
	save func(transactionID, scheduleID, status string) error) {
	for {
		response, err := fetch()
		if err != nil {
			s.logger.Errorf("[%s] - Failed to get token burn transactions after timestamp. Error: [%s]", transferID, err)
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
	}
}
