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
	hederaHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	entityStatus "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
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
	prometheusService  service.Prometheus
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
	readOnlyService service.ReadOnly,
	prometheusServices service.Prometheus) *Handler {
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
		prometheusService:  prometheusServices,
	}
}

func (fmh *Handler) Handle(payload interface{}) {
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

	if transactionRecord.Status != entityStatus.Initial {
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

	var (
		feeOutParams  *hederaHelper.FeeOutParams
		userOutParams *hederaHelper.UserOutParams
	)

	if fmh.prometheusService.GetIsMonitoringEnabled() {
		feeOutParams = hederaHelper.NewFeeOutParams(len(splitTransfers))
		userOutParams = hederaHelper.NewUserOutParams()
	}

	for _, splitTransfer := range splitTransfers {
		feeAmount, hasReceiver := util.GetTotalFeeFromTransfers(splitTransfer, receiver)

		fmh.readOnlyService.FindAssetTransfer(transferMsg.TransactionId, transferMsg.TargetAsset, splitTransfer, func() (*mirror_node.Response, error) {
			return fmh.mirrorNode.GetAccountDebitTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.Timestamp)
		}, func(transactionID, scheduleID, status string) error {
			result := false
			if status == entityStatus.Completed {
				result = true
			}

			if fmh.prometheusService.GetIsMonitoringEnabled() {
				feeOutParams.HandleResultForAwaitedTransfer(&result, hasReceiver, splitTransfer)
				userOutParams.HandleResultForAwaitedTransfer(&result, hasReceiver)
			}

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

	fmh.startAwaitingFunctionsForMetrics(userOutParams, transferMsg, feeOutParams)
}

func (fmh *Handler) startAwaitingFunctionsForMetrics(userOutParams *hederaHelper.UserOutParams, transferMsg *model.Transfer, feeOutParams *hederaHelper.FeeOutParams) {
	if !fmh.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	// Await results to set metrics for 'user_get_his_tokens'
	go hederaHelper.AwaitMultipleScheduledTransactions(
		userOutParams.OutParams,
		transferMsg.SourceChainId,
		transferMsg.TargetChainId,
		transferMsg.TargetAsset,
		transferMsg.TransactionId,
		fmh.onMinedUserTransactionSetMetrics,
	)

	// Await results to set metrics for 'fee_transferred'
	go hederaHelper.AwaitMultipleScheduledTransactions(
		feeOutParams.OutParams,
		transferMsg.SourceChainId,
		transferMsg.TargetChainId,
		transferMsg.TargetAsset,
		transferMsg.TransactionId,
		fmh.onMinedFeeTransactionsSetMetrics,
	)
}

func (fmh *Handler) onMinedFeeTransactionsSetMetrics(sourceChainId int64, targetChainId int64, nativeAsset string, transferID string, isTransferSuccessful bool) {
	if sourceChainId == constants.HederaNetworkId || isTransferSuccessful == false || !fmh.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	metrics.SetFeeTransferred(sourceChainId, targetChainId, nativeAsset, transferID, fmh.prometheusService, fmh.logger)
}

func (fmh *Handler) onMinedUserTransactionSetMetrics(sourceChainId int64, targetChainId int64, nativeAsset string, transferID string, isTransferSuccessful bool) {
	if sourceChainId == constants.HederaNetworkId || isTransferSuccessful == false || !fmh.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	metrics.SetUserGetHisTokens(sourceChainId, targetChainId, nativeAsset, transferID, fmh.prometheusService, fmh.logger)
}
