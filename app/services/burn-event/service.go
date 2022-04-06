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

package burn_event

import (
	"database/sql"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	util "github.com/limechain/hedera-eth-bridge-validator/app/helper/fee"
	hederaHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Service struct {
	bridgeAccount      hedera.AccountID
	feeRepository      repository.Fee
	repository         repository.Transfer
	scheduleRepository repository.Schedule
	distributorService service.Distributor
	feeService         service.Fee
	scheduledService   service.Scheduled
	transferService    service.Transfers
	logger             *log.Entry
	prometheusService  service.Prometheus
}

func NewService(
	bridgeAccount string,
	repository repository.Transfer,
	scheduleRepository repository.Schedule,
	feeRepository repository.Fee,
	distributor service.Distributor,
	scheduled service.Scheduled,
	feeService service.Fee,
	transferService service.Transfers,
	prometheusService service.Prometheus) *Service {

	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid bridge account: [%s].", bridgeAccount)
	}

	return &Service{
		bridgeAccount:      bridgeAcc,
		feeRepository:      feeRepository,
		repository:         repository,
		scheduleRepository: scheduleRepository,
		distributorService: distributor,
		feeService:         feeService,
		scheduledService:   scheduled,
		transferService:    transferService,
		prometheusService:  prometheusService,
		logger:             config.GetLoggerFor("Burn Event Service"),
	}
}

func (s Service) ProcessEvent(event transfer.Transfer) {
	s.initSuccessRatePrometheusMetrics(event.TransactionId, event.SourceChainId, event.TargetChainId, event.TargetAsset)

	amount, err := strconv.ParseInt(event.Amount, 10, 64)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to parse event amount [%s]. Error [%s].", event.TransactionId, event.Amount, err)
		return
	}

	receiver, err := hedera.AccountIDFromString(event.Receiver)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to parse event account [%s]. Error [%s].", event.TransactionId, event.Receiver, err)
		return
	}

	transactionRecord, err := s.transferService.InitiateNewTransfer(event)
	if err != nil {
		s.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", event.TransactionId, err)
		return
	}

	if transactionRecord.Status != status.Initial {
		s.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	fee, splitTransfers, err := s.prepareTransfers(event.NativeAsset, amount, receiver)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to prepare transfers. Error [%s].", event.TransactionId, err)
		return
	}

	err = s.repository.UpdateFee(event.TransactionId, strconv.FormatInt(fee, 10))
	if err != nil {
		s.logger.Errorf("[%s] - Failed to update fee [%d]. Error [%s].", event.TransactionId, fee, err)
		return
	}

	var (
		feeOutParams  *hederaHelper.FeeOutParams
		userOutParams *hederaHelper.UserOutParams
	)

	if s.prometheusService.GetIsMonitoringEnabled() {
		feeOutParams = hederaHelper.NewFeeOutParams(len(splitTransfers))
		userOutParams = hederaHelper.NewUserOutParams()
	}

	for _, splitTransfer := range splitTransfers {
		feeAmount, hasReceiver := util.TotalFeeFromTransfers(splitTransfer, receiver)
		onExecutionSuccess, onExecutionFail := s.scheduledTxExecutionCallbacks(event.TransactionId, feeAmount, hasReceiver)

		onSuccess, onFail := s.scheduledTxMinedCallbacks(
			event.TransactionId,
			hasReceiver,
			splitTransfer,
			feeOutParams,
			userOutParams,
		)

		s.scheduledService.ExecuteScheduledTransferTransaction(event.TransactionId, event.NativeAsset, splitTransfer, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	}

	s.startAwaitingFunctionsForMetrics(event, feeOutParams, userOutParams)
}

func (s Service) startAwaitingFunctionsForMetrics(event transfer.Transfer, feeOutParams *hederaHelper.FeeOutParams, userOutParams *hederaHelper.UserOutParams) {
	if !s.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	// Await Fee Transfer To Set Metrics
	go hederaHelper.AwaitMultipleScheduledTransactions(
		feeOutParams.OutParams,
		event.SourceChainId,
		event.TargetChainId,
		event.NativeAsset,
		event.TransactionId,
		s.onMinedFeeTransactionsSetMetrics,
	)

	// Await User Transfer To Set Metrics
	go hederaHelper.AwaitMultipleScheduledTransactions(
		userOutParams.OutParams,
		event.SourceChainId,
		event.TargetChainId,
		event.NativeAsset,
		event.TransactionId,
		s.onMinedUserTransactionSetMetrics,
	)
}

func (s Service) initSuccessRatePrometheusMetrics(transactionId string, sourceChainId, targetChainId uint64, asset string) {
	if !s.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	if sourceChainId != constants.HederaNetworkId {
		if targetChainId == constants.HederaNetworkId {
			metrics.CreateFeeTransferredIfNotExists(sourceChainId, targetChainId, asset, transactionId, s.prometheusService, s.logger)
		}
		metrics.CreateUserGetHisTokensIfNotExists(sourceChainId, targetChainId, asset, transactionId, s.prometheusService, s.logger)
	}
}

func (s *Service) onMinedFeeTransactionsSetMetrics(sourceChainId, targetChainId uint64, nativeAsset string, transactionId string, isTransferSuccessful bool) {

	if !s.prometheusService.GetIsMonitoringEnabled() || targetChainId != constants.HederaNetworkId || !isTransferSuccessful {
		return
	}

	metrics.SetFeeTransferred(sourceChainId, targetChainId, nativeAsset, transactionId, s.prometheusService, s.logger)
}

func (s *Service) onMinedUserTransactionSetMetrics(sourceChainId, targetChainId uint64, nativeAsset string, transactionId string, isTransferSuccessful bool) {

	if !s.prometheusService.GetIsMonitoringEnabled() || sourceChainId == constants.HederaNetworkId || !isTransferSuccessful {
		return
	}

	metrics.SetUserGetHisTokens(sourceChainId, targetChainId, nativeAsset, transactionId, s.prometheusService, s.logger)
}

func (s *Service) prepareTransfers(token string, amount int64, receiver hedera.AccountID) (fee int64, splitTransfers [][]transfer.Hedera, err error) {
	fee, remainder := s.feeService.CalculateFee(token, amount)

	validFee := s.distributorService.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	transfers, err := s.distributorService.CalculateMemberDistribution(validFee)
	if err != nil {
		return 0, nil, err
	}

	transfers = append(transfers,
		transfer.Hedera{
			AccountID: receiver,
			Amount:    remainder,
		})

	splitTransfers = distributor.SplitAccountAmounts(transfers,
		transfer.Hedera{
			AccountID: s.bridgeAccount,
			Amount:    -amount,
		})

	return validFee, splitTransfers, nil
}

// TransactionID returns the corresponding Scheduled Transaction paying out the
// fees to validators and the amount being bridged to the receiver address
func (s *Service) TransactionID(id string) (string, error) {
	event, err := s.scheduleRepository.GetReceiverTransferByTransactionID(id)
	if err != nil {
		s.logger.Errorf("[%s] - failed to get event.", id)
		return "", err
	}

	if event == nil {
		return "", service.ErrNotFound
	}

	return event.TransactionID, nil
}

func (s *Service) scheduledTxExecutionCallbacks(id string, feeAmount string, hasReceiver bool) (onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		s.logger.Debugf("[%s] - Updating db status to Submitted with TransactionID [%s].",
			id,
			transactionID)
		err := s.scheduleRepository.Create(&entity.Schedule{
			ScheduleID:    scheduleID,
			Operation:     schedule.TRANSFER,
			TransactionID: transactionID,
			HasReceiver:   hasReceiver,
			Status:        status.Submitted,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			s.logger.Errorf(
				"[%s] - Failed to update submitted status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				id, transactionID, scheduleID, err)
			return
		}
		err = s.feeRepository.Create(&entity.Fee{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Amount:        feeAmount,
			Status:        status.Submitted,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			s.logger.Errorf(
				"[%s] - Failed to create Fee Record [%s]. Error [%s].",
				transactionID, id, err)
			return
		}
	}

	onExecutionFail = func(transactionID string) {
		err := s.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			Status:        status.Failed,
			HasReceiver:   hasReceiver,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", id, err)
			return
		}

		err = s.repository.UpdateStatusFailed(id)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", id, err)
			return
		}

		err = s.feeRepository.Create(&entity.Fee{
			TransactionID: transactionID,
			Amount:        feeAmount,
			Status:        status.Failed,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			s.logger.Errorf("[%s] Fee - Failed to create failed record. Error [%s].", transactionID, err)
			return
		}
	}

	return onExecutionSuccess, onExecutionFail
}

func (s *Service) scheduledTxMinedCallbacks(id string, hasReceiver bool, splitTransfer []transfer.Hedera, feeOutParams *hederaHelper.FeeOutParams, userOutParams *hederaHelper.UserOutParams) (onSuccess, onFail func(transactionID string)) {

	onSuccess = func(transactionID string) {

		s.logger.Debugf("[%s] - Scheduled TX execution successful.", id)
		if s.prometheusService.GetIsMonitoringEnabled() {
			result := true
			feeOutParams.HandleResultForAwaitedTransfer(&result, hasReceiver, splitTransfer)
			userOutParams.HandleResultForAwaitedTransfer(&result, hasReceiver)
		}

		err := s.repository.UpdateStatusCompleted(id)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status completed. Error [%s].", id, err)
			return
		}
		err = s.scheduleRepository.UpdateStatusCompleted(transactionID)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status completed. Error [%s].", transactionID, err)
			return
		}

		err = s.feeRepository.UpdateStatusCompleted(transactionID)
		if err != nil {
			s.logger.Errorf("[%s] Fee - Failed to update status completed. Error [%s].", transactionID, err)
			return
		}
	}

	onFail = func(transactionID string) {
		s.logger.Debugf("[%s] - Scheduled TX execution has failed.", id)
		if s.prometheusService.GetIsMonitoringEnabled() {
			result := false
			feeOutParams.HandleResultForAwaitedTransfer(&result, hasReceiver, splitTransfer)
			userOutParams.HandleResultForAwaitedTransfer(&result, hasReceiver)
		}

		err := s.scheduleRepository.UpdateStatusFailed(transactionID)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", id, err)
			return
		}

		err = s.repository.UpdateStatusFailed(id)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", transactionID, err)
			return
		}

		err = s.feeRepository.UpdateStatusFailed(transactionID)
		if err != nil {
			s.logger.Errorf("[%s] Fee - Failed to update status failed. Error [%s].", transactionID, err)
			return
		}
	}

	return onSuccess, onFail
}
