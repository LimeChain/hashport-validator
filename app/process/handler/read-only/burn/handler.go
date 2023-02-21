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

package burn

import (
	"database/sql"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirrorNodeTransaction "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	bridgeAccount      hedera.AccountID
	transfersService   service.Transfers
	scheduleRepository repository.Schedule
	transferRepository repository.Transfer
	mirrorNode         client.MirrorNode
	readOnlyService    service.ReadOnly
	logger             *log.Entry
}

func NewHandler(
	bridgeAccount string,
	mirrorNode client.MirrorNode,
	scheduleRepository repository.Schedule,
	transferRepository repository.Transfer,
	transferService service.Transfers,
	readOnlyService service.ReadOnly) *Handler {
	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}
	return &Handler{
		bridgeAccount:      bridgeAcc,
		mirrorNode:         mirrorNode,
		transfersService:   transferService,
		scheduleRepository: scheduleRepository,
		transferRepository: transferRepository,
		readOnlyService:    readOnlyService,
		logger:             config.GetLoggerFor("Hedera Burn and Topic Message Read-only Handler"),
	}
}

func (mhh Handler) Handle(p interface{}) {
	transferMsg, ok := p.(*payload.Transfer)
	if !ok {
		mhh.logger.Errorf("Could not cast payload [%s]", p)
		return
	}

	transactionRecord, err := mhh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		mhh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != status.Initial {
		mhh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	mhh.readOnlyService.FindTransfer(transferMsg.TransactionId,
		func() (*mirrorNodeTransaction.Response, error) {
			return mhh.mirrorNode.GetAccountTokenBurnTransactionsAfterTimestampString(mhh.bridgeAccount, transferMsg.NetworkTimestamp)
		},
		func(transactionID, scheduleID, s string) error {

			if s == status.Completed {
				err = mhh.transferRepository.UpdateStatusCompleted(transferMsg.TransactionId)
			} else {
				err = mhh.transferRepository.UpdateStatusFailed(transferMsg.TransactionId)
			}

			if err != nil {
				mhh.logger.Errorf("[%s] - Failed to update status. Error: [%s]", transferMsg.TransactionId, err)
			}

			return mhh.scheduleRepository.Create(&entity.Schedule{
				TransactionID: transactionID,
				ScheduleID:    scheduleID,
				Operation:     schedule.BURN,
				Status:        s,
				TransferID: sql.NullString{
					String: transferMsg.TransactionId,
					Valid:  true,
				},
			})
		})
}
