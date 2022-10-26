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

package fee

import (
	"database/sql"
	"strconv"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	transferRepository repository.Transfer
	feeRepository      repository.Fee
	scheduleRepository repository.Schedule
	mirrorNode         client.MirrorNode
	bridgeAccount      hedera.AccountID
	distributor        service.Distributor
	transfersService   service.Transfers
	readOnlyService    service.ReadOnly
	logger             *log.Entry
}

func NewHandler(
	transferRepository repository.Transfer,
	feeRepository repository.Fee,
	scheduleRepository repository.Schedule,
	mirrorNode client.MirrorNode,
	bridgeAccount string,
	distributor service.Distributor,
	transfersService service.Transfers,
	readOnlyService service.ReadOnly) *Handler {
	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}

	instance := &Handler{
		transferRepository: transferRepository,
		feeRepository:      feeRepository,
		scheduleRepository: scheduleRepository,
		mirrorNode:         mirrorNode,
		bridgeAccount:      bridgeAcc,
		logger:             config.GetLoggerFor("Hedera Transfer and Topic Submission Read-only Handler"),
		transfersService:   transfersService,
		distributor:        distributor,
		readOnlyService:    readOnlyService,
	}

	return instance
}

func (fmh Handler) Handle(p interface{}) {
	transferMsg, ok := p.(*payload.Transfer)
	if !ok {
		fmh.logger.Errorf("Could not cast payload [%s]", p)
		return
	}

	transactionRecord, err := fmh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		fmh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != status.Initial {
		fmh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	fmh.readOnlyService.FindNftTransfer(
		transferMsg.TransactionId,
		transferMsg.SourceAsset,
		transferMsg.SerialNum,
		transferMsg.Originator,
		fmh.bridgeAccount.String(),
		func(transactionID, scheduleID, status string) error {
			err := fmh.scheduleRepository.Create(&entity.Schedule{
				TransactionID: transferMsg.TransactionId,
				ScheduleID:    scheduleID,
				Operation:     schedule.TRANSFER,
				Status:        status,
				HasReceiver:   true,
				TransferID: sql.NullString{
					String: transferMsg.TransactionId,
					Valid:  true,
				},
			})
			if err != nil {
				fmh.logger.Errorf("[%s] - Failed to create scheduled entity [%s]. Error: [%s]", transactionID, scheduleID, err)
				return err
			}

			return fmh.transferRepository.UpdateStatusCompleted(transactionID)
		},
	)

	validFee := fmh.distributor.ValidAmount(transferMsg.Fee)
	err = fmh.transferRepository.UpdateFee(transferMsg.TransactionId, strconv.FormatInt(validFee, 10))
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to update fee [%d]. Error: [%s]", transferMsg.TransactionId, validFee, err)
		return
	}

	transfers, _ := fmh.distributor.CalculateMemberDistribution(validFee)

	splitTransfers := distributor.SplitAccountAmounts(transfers,
		model.Hedera{
			AccountID: fmh.bridgeAccount,
			Amount:    -validFee,
		})

	for _, splitTransfer := range splitTransfers {
		feeAmount := -splitTransfer[len(splitTransfer)-1].Amount
		fmh.readOnlyService.FindAssetTransfer(transferMsg.TransactionId, constants.Hbar, splitTransfer,
			func() (*mirror_node.Response, error) {
				return fmh.feeTransfersFetch(transferMsg)
			},
			func(transactionID, scheduleID, status string) error {
				return fmh.feeTransfersSave(transactionID, scheduleID, status, transferMsg, feeAmount)
			})
	}
}

func (fmh Handler) feeTransfersFetch(transferMsg *payload.Transfer) (*mirror_node.Response, error) {
	return fmh.mirrorNode.GetAccountDebitTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.NetworkTimestamp)
}

func (fmh Handler) feeTransfersSave(transactionID string, scheduleID string, status string, transferMsg *payload.Transfer, feeAmount int64) error {
	err := fmh.scheduleRepository.Create(&entity.Schedule{
		TransactionID: transactionID,
		ScheduleID:    scheduleID,
		Operation:     schedule.TRANSFER,
		Status:        status,
		TransferID: sql.NullString{
			String: transferMsg.TransactionId,
			Valid:  true,
		},
	})
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to create scheduled entity [%s]. Error: [%s]", transferMsg.TransactionId, scheduleID, err)
		return err
	}
	err = fmh.feeRepository.Create(&entity.Fee{
		TransactionID: transactionID,
		ScheduleID:    scheduleID,
		Amount:        strconv.FormatInt(feeAmount, 10),
		Status:        status,
		TransferID: sql.NullString{
			String: transferMsg.TransactionId,
			Valid:  true,
		},
	})
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to create fee  entity [%s]. Error: [%s]", transferMsg.TransactionId, scheduleID, err)
	}
	return err
}
