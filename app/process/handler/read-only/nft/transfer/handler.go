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

package transfer

import (
	"database/sql"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	transferRepository repository.Transfer
	scheduleRepository repository.Schedule
	bridgeAccount      hedera.AccountID
	transfersService   service.Transfers
	readOnlyService    service.ReadOnly
	logger             *log.Entry
}

func NewHandler(
	bridgeAccount string,
	transferRepository repository.Transfer,
	scheduleRepository repository.Schedule,
	readOnlyService service.ReadOnly,
	transfersService service.Transfers) *Handler {
	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}
	return &Handler{
		bridgeAccount:      bridgeAcc,
		transferRepository: transferRepository,
		scheduleRepository: scheduleRepository,
		readOnlyService:    readOnlyService,
		transfersService:   transfersService,
		logger:             config.GetLoggerFor("Read-only Hedera NFT Transfer"),
	}
}

func (rnth Handler) Handle(p interface{}) {
	transfer, ok := p.(*payload.Transfer)
	if !ok {
		rnth.logger.Errorf("Could not cast payload [%s]", p)
		return
	}

	transactionRecord, err := rnth.transfersService.InitiateNewTransfer(*transfer)
	if err != nil {
		rnth.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transfer.TransactionId, err)
		return
	}

	if transactionRecord.Status != status.Initial {
		rnth.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	rnth.readOnlyService.FindScheduledNftAllowanceApprove(
		transfer,
		rnth.bridgeAccount,
		func(transactionID, scheduleID, status string) error {
			err := rnth.scheduleRepository.Create(&entity.Schedule{
				TransactionID: transactionID,
				ScheduleID:    scheduleID,
				Operation:     schedule.APPROVE,
				Status:        status,
				HasReceiver:   true,
				TransferID: sql.NullString{
					String: transfer.TransactionId,
					Valid:  true,
				},
			})
			if err != nil {
				rnth.logger.Errorf("[%s] - Error to create scheduled entity. Error: [%s]", transactionID, err)
				return err
			}
			return rnth.transferRepository.UpdateStatusCompleted(transactionID)
		},
	)
}
