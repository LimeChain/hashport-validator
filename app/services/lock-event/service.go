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

package lock_event

import (
	"database/sql"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	syncHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/sync"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Service struct {
	bridgeAccount      hedera.AccountID
	repository         repository.Transfer
	scheduleRepository repository.Schedule
	transferService    service.Transfers
	scheduledService   service.Scheduled
	prometheusService  service.Prometheus
	logger             *log.Entry
}

func NewService(
	bridgeAccount string,
	repository repository.Transfer,
	scheduleRepository repository.Schedule,
	scheduled service.Scheduled,
	transferService service.Transfers,
	prometheusService service.Prometheus) *Service {

	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid bridge account: [%s].", bridgeAccount)
	}

	return &Service{
		bridgeAccount:      bridgeAcc,
		repository:         repository,
		scheduleRepository: scheduleRepository,
		scheduledService:   scheduled,
		transferService:    transferService,
		prometheusService:  prometheusService,
		logger:             config.GetLoggerFor("Lock Event Service"),
	}
}

func (s *Service) ProcessEvent(event transfer.Transfer) {
	amount, err := strconv.ParseInt(event.Amount, 10, 64)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to parse event amount [%s]. Error [%s].", event.TransactionId, event.Amount, err)
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

	status := make(chan string)

	onTokenMintSuccess, onTokenMintFail := s.scheduledTxMinedCallbacks(event.TransactionId, &status, event, schedule.MINT)
	onExecutionMintSuccess, onExecutionMintFail := s.scheduledTxExecutionCallbacks(event.TransactionId, schedule.MINT, &status, false)

	s.scheduledService.ExecuteScheduledMintTransaction(
		event.TransactionId,
		event.TargetAsset,
		amount,
		&status,
		onExecutionMintSuccess,
		onExecutionMintFail,
		onTokenMintSuccess,
		onTokenMintFail,
	)

	// TODO: Figure out Unit Testing on this one
	s.logger.Debugf("[%s] - Waiting for Mint Transaction Execution.", event.TransactionId)
statusBlocker:
	for {
		switch <-status {
		case syncHelper.DONE:
			s.logger.Debugf("[%s] - Proceeding to submit the Scheduled Transfer Transaction.", event.TransactionId)
			break statusBlocker
		case syncHelper.FAIL:
			s.logger.Errorf("[%s] - Failed to await the execution of Scheduled Mint Transaction.", event.TransactionId)
			return
		}
	}
	accountID, err := hedera.AccountIDFromString(event.Receiver)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to parse receiver [%s]. Error: [%s].", event.TransactionId, event.Receiver, err)
		return
	}

	transfers := []transfer.Hedera{
		{
			AccountID: accountID,
			Amount:    amount,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -amount,
		},
	}

	onExecutionTransferSuccess, onExecutionTransferFail := s.scheduledTxExecutionCallbacks(event.TransactionId, schedule.TRANSFER, &status, true)
	onTransferSuccess, onTransferFail := s.scheduledTxMinedCallbacks(event.TransactionId, &status, event, schedule.TRANSFER)

	s.scheduledService.ExecuteScheduledTransferTransaction(
		event.TransactionId,
		event.TargetAsset,
		transfers,
		onExecutionTransferSuccess,
		onExecutionTransferFail,
		onTransferSuccess,
		onTransferFail,
	)
}

func (s *Service) scheduledTxExecutionCallbacks(id, operation string, blocker *chan string, hasReceiver bool) (onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		s.logger.Debugf("[%s] - Updating db status Submitted with TransactionID [%s].",
			id,
			transactionID)
		err := s.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Operation:     operation,
			HasReceiver:   hasReceiver,
			Status:        status.Submitted,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			*blocker <- syncHelper.FAIL
			s.logger.Errorf(
				"[%s] - Failed to update submitted scheduled status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				id, transactionID, scheduleID, err)
			return
		}
	}

	onExecutionFail = func(transactionID string) {
		*blocker <- syncHelper.FAIL
		err := s.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			Status:        status.Failed,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", id, err)
			return
		}
	}

	return onExecutionSuccess, onExecutionFail
}

func (s *Service) scheduledTxMinedCallbacks(id string, status *chan string, event transfer.Transfer, scheduleType string) (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {

		if scheduleType == schedule.TRANSFER && s.prometheusService.GetIsMonitoringEnabled() {
			s.setUserGetHisTokensMetric(event.SourceChainId, event.TargetChainId, event.SourceAsset, event.TransactionId, true)
		}

		s.logger.Debugf("[%s] - Scheduled [%s] TX execution successful.", id, transactionID)

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

		if err != nil {
			*status <- syncHelper.FAIL
			s.logger.Errorf("[%s] - Failed to update scheduled [%s] status completed. Error [%s].", id, transactionID, err)
			return
		}
		*status <- syncHelper.DONE
	}

	onFail = func(transactionID string) {

		*status <- syncHelper.FAIL
		s.logger.Debugf("[%s] - Scheduled TX execution has failed.", id)
		err := s.scheduleRepository.UpdateStatusFailed(id)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update schedule status failed. Error [%s].", id, err)
			return
		}

		err = s.repository.UpdateStatusFailed(transactionID)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update transfer status failed. Error [%s].", transactionID, err)
			return
		}
	}

	return onSuccess, onFail
}

func (s *Service) setUserGetHisTokensMetric(sourceChainId int64, targetChainId int64, sourceAsset string, transferID string, isTransferSuccessful bool) {
	gauge, err := s.prometheusService.CreateAndRegisterGaugeMetricForSuccessRate(
		transferID,
		sourceChainId,
		targetChainId,
		sourceAsset,
		constants.UserGetHisTokensNameSuffix,
		constants.UserGetHisTokensHelp)

	if err != nil {
		s.logger.Errorf("[%s] - Failed to create gauge metric for [%s]. Error: %s", transferID, constants.UserGetHisTokensNameSuffix, err)
	}

	if isTransferSuccessful {
		s.logger.Infof("[%s] - Setting value to 1.0 for metric [%v]", transferID, constants.UserGetHisTokensNameSuffix)
		gauge.Set(1.0)
	}

}
