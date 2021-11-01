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

package fee_transfer

import (
	"database/sql"
	"github.com/hashgraph/hedera-sdk-go/v2"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	util "github.com/limechain/hedera-eth-bridge-validator/app/helper/fee"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"strconv"
)

// Handler is transfers event handler
type Handler struct {
	transferRepository repository.Transfer
	feeRepository      repository.Fee
	scheduleRepository repository.Schedule
	mirrorNode         client.MirrorNode
	bridgeAccount      hedera.AccountID
	feeService         service.Fee
	distributorService service.Distributor
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
	distributorService service.Distributor,
	feeService service.Fee,
	transfersService service.Transfers,
	readOnlyService service.ReadOnly) *Handler {
	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}
	return &Handler{
		transferRepository: transferRepository,
		feeRepository:      feeRepository,
		scheduleRepository: scheduleRepository,
		mirrorNode:         mirrorNode,
		bridgeAccount:      bridgeAcc,
		distributorService: distributorService,
		feeService:         feeService,
		logger:             config.GetLoggerFor("Hedera Fee and Schedule Transfer Read-only Handler"),
		transfersService:   transfersService,
		readOnlyService:    readOnlyService,
	}
}

func (fmh Handler) Handle(payload interface{}) {
	transferMsg, ok := payload.(*model.Transfer)
	if !ok {
		fmh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	receiver, err := hedera.AccountIDFromString(transferMsg.Receiver)
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to parse event account [%s]. Error [%s].", transferMsg.TransactionId, transferMsg.Receiver, err)
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

	intAmount, err := strconv.ParseInt(transferMsg.Amount, 10, 64)
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to parse amount. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	calculatedFee, remainder := fmh.feeService.CalculateFee(transferMsg.TargetAsset, intAmount)

	validFee := fmh.distributorService.ValidAmount(calculatedFee)
	if validFee != calculatedFee {
		remainder += calculatedFee - validFee
	}

	err = fmh.transferRepository.UpdateFee(transferMsg.TransactionId, strconv.FormatInt(validFee, 10))
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to update fee [%d]. Error: [%s]", transferMsg.TransactionId, validFee, err)
		return
	}

	transfers, err := fmh.distributorService.CalculateMemberDistribution(validFee)
	transfers = append(transfers,
		model.Hedera{
			AccountID: receiver,
			Amount:    remainder,
		})

	splitTransfers := distributor.SplitAccountAmounts(transfers,
		model.Hedera{
			AccountID: fmh.bridgeAccount,
			Amount:    -intAmount,
		})

	for _, splitTransfer := range splitTransfers {
		feeAmount, hasReceiver := util.GetTotalFeeFromTransfers(splitTransfer, receiver)

		fmh.readOnlyService.FindAssetTransfer(transferMsg.TransactionId, transferMsg.TargetAsset, splitTransfer, func() (*mirror_node.Response, error) {
			return fmh.mirrorNode.GetAccountDebitTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.Timestamp)
		}, func(transactionID, scheduleID, status string) error {
			err := fmh.scheduleRepository.Create(&entity.Schedule{
				TransactionID: transactionID,
				ScheduleID:    scheduleID,
				Operation:     schedule.TRANSFER,
				HasReceiver:   hasReceiver,
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
				Amount:        feeAmount,
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
		})
	}
}
