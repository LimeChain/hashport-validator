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
	"github.com/hashgraph/hedera-sdk-go/v2"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
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
	distributor        service.Distributor
	feeService         service.Fee
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
	distributor service.Distributor,
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
		logger:             config.GetLoggerFor("Hedera Transfer and Topic Submission Read-only Handler"),
		transfersService:   transfersService,
		distributor:        distributor,
		feeService:         feeService,
		readOnlyService:    readOnlyService,
		prometheusService:  prometheusServices,
	}
}

func (fmh Handler) Handle(payload interface{}) {
	transferMsg, ok := payload.(*model.Transfer)
	if !ok {
		fmh.logger.Errorf("Could not cast payload [%s]", payload)
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

	calculatedFee, _ := fmh.feeService.CalculateFee(transferMsg.SourceAsset, intAmount)
	validFee := fmh.distributor.ValidAmount(calculatedFee)

	err = fmh.transferRepository.UpdateFee(transferMsg.TransactionId, strconv.FormatInt(validFee, 10))
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to update fee [%d]. Error: [%s]", transferMsg.TransactionId, validFee, err)
		return
	}

	transfers, err := fmh.distributor.CalculateMemberDistribution(validFee)

	splitTransfers := distributor.SplitAccountAmounts(transfers,
		model.Hedera{
			AccountID: fmh.bridgeAccount,
			Amount:    -validFee,
		})

	var (
		feeOutParams *hederaHelper.FeeOutParams
	)

	if fmh.prometheusService.GetIsMonitoringEnabled() {
		feeOutParams = hederaHelper.NewFeeOutParams(len(splitTransfers))
	}

	for _, splitTransfer := range splitTransfers {
		feeAmount := -splitTransfer[len(splitTransfer)-1].Amount

		fmh.readOnlyService.FindAssetTransfer(transferMsg.TransactionId, transferMsg.NativeAsset, splitTransfer,
			func() (*mirror_node.Response, error) {
				return fmh.mirrorNode.GetAccountDebitTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.Timestamp)
			},

			func(transactionID, scheduleID, status string) error {
				if fmh.prometheusService.GetIsMonitoringEnabled() {
					// For Awaiting result to set Metrics
					result := false
					if status == entityStatus.Completed {
						result = true
					}
					feeOutParams.HandleResultForAwaitedTransfer(&result, false, splitTransfer)
				}

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
			})
	}

	if fmh.prometheusService.GetIsMonitoringEnabled() {
		go hederaHelper.AwaitMultipleScheduledTransactions(
			feeOutParams.OutParams,
			transferMsg.SourceChainId,
			transferMsg.TargetChainId,
			transferMsg.SourceAsset,
			transferMsg.TransactionId,
			fmh.onMinedFeeTransactionsSetMetrics,
		)
	}
}

func (fmh *Handler) onMinedFeeTransactionsSetMetrics(sourceChainId, targetChainId uint64, nativeAsset string, transferID string, isTransferSuccessful bool) {
	if sourceChainId != constants.HederaNetworkId || isTransferSuccessful == false || !fmh.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	metrics.SetFeeTransferred(sourceChainId, targetChainId, nativeAsset, transferID, fmh.prometheusService, fmh.logger)
}
