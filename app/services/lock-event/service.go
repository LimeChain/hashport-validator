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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

const (
	SCHEDULED_MINT_TYPE     = "mint"
	SCHEDULED_TRANSFER_TYPE = "transfer"
)

type Service struct {
	bridgeAccount    hedera.AccountID
	repository       repository.LockEvent
	scheduledService service.Scheduled
	logger           *log.Entry
}

func NewService(
	bridgeAccount string,
	repository repository.LockEvent,
	scheduled service.Scheduled) *Service {

	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid bridge account: [%s].", bridgeAccount)
	}

	return &Service{
		bridgeAccount:    bridgeAcc,
		repository:       repository,
		scheduledService: scheduled,
		logger:           config.GetLoggerFor("Lock Event Service"),
	}
}

func (s Service) ProcessEvent(event lock_event.LockEvent) {
	err := s.repository.Create(event.Id, event.Amount, event.Recipient.String(), event.NativeAsset, event.WrappedAsset)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create a lock event record. Error [%s].", event.Id, err)
		return
	}

	onExecutionMintSuccess, onExecutionMintFail := s.scheduledTxExecutionCallbacks(event.Id, SCHEDULED_MINT_TYPE)
	onTokenMintSuccess, onTokenMintFail := s.scheduledTxMinedCallbacks(event.Id, SCHEDULED_MINT_TYPE)

	err = s.scheduledService.ExecuteScheduledMintTransaction(
		event.Id,
		event.WrappedAsset,
		event.Amount,
		onExecutionMintSuccess,
		onExecutionMintFail,
		onTokenMintSuccess,
		onTokenMintFail,
	)
	if err != nil {
		s.logger.Errorf("[%s] - Execution of Scheduled Mint Transaction failed. Error [%s].", event.Id, err)
		return
	}

	transfers := []transfer.Hedera{
		{
			AccountID: event.Recipient,
			Amount:    event.Amount,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -event.Amount,
		},
	}

	onExecutionTransferSuccess, onExecutionTransferFail := s.scheduledTxExecutionCallbacks(event.Id, SCHEDULED_TRANSFER_TYPE)
	onTransferSuccess, onTransferFail := s.scheduledTxMinedCallbacks(event.Id, SCHEDULED_TRANSFER_TYPE)

	s.scheduledService.ExecuteScheduledTransferTransaction(
		event.Id,
		event.WrappedAsset,
		transfers,
		onExecutionTransferSuccess,
		onExecutionTransferFail,
		onTransferSuccess,
		onTransferFail,
	)
}

func (s *Service) scheduledTxExecutionCallbacks(id, txType string) (onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		var err error
		s.logger.Debugf("[%s] - Updating db status to Submitted with TransactionID [%s].",
			id,
			transactionID)
		switch txType {
		case SCHEDULED_MINT_TYPE:
			err = s.repository.UpdateStatusScheduledTokenMintSubmitted(id, scheduleID, transactionID)
		case SCHEDULED_TRANSFER_TYPE:
			err = s.repository.UpdateStatusScheduledTokenTransferSubmitted(id, scheduleID, transactionID)
		}
		if err != nil {
			s.logger.Errorf(
				"[%s] - Failed to update submitted scheduled %s status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				id, txType, transactionID, scheduleID, err)
			return
		}
	}

	onExecutionFail = func(transactionID string) {
		err := s.repository.UpdateStatusFailed(id)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", id, err)
			return
		}
	}

	return onExecutionSuccess, onExecutionFail
}

func (s *Service) scheduledTxMinedCallbacks(id, txType string) (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {
		s.logger.Debugf("[%s] - Scheduled %s TX execution successful.", id, txType)

		var err error
		switch txType {
		case SCHEDULED_MINT_TYPE:
			err = s.repository.UpdateStatusScheduledTokenMintCompleted(id)
		case SCHEDULED_TRANSFER_TYPE:
			err = s.repository.UpdateStatusCompleted(id)
		}
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update scheduled %s status completed. Error [%s].", id, txType, err)
			return
		}
	}

	onFail = func(transactionID string) {
		s.logger.Debugf("[%s] - Scheduled TX execution has failed.", id)
		err := s.repository.UpdateStatusFailed(id)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", id, err)
			return
		}
	}

	return onSuccess, onFail
}
