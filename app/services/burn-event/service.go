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

package burn_event

import (
	"database/sql"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	util "github.com/limechain/hedera-eth-bridge-validator/app/helper/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	"github.com/limechain/hedera-eth-bridge-validator/config"
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
	logger             *log.Entry
}

func NewService(
	bridgeAccount string,
	repository repository.Transfer,
	scheduleRepository repository.Schedule,
	feeRepository repository.Fee,
	distributor service.Distributor,
	scheduled service.Scheduled,
	feeService service.Fee) *Service {

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
		logger:             config.GetLoggerFor("Burn Event Service"),
	}
}

func (s Service) ProcessEvent(event transfer.Transfer) {
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

	_, err = s.repository.Create(&event)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create a burn event record. Error [%s].", event.TransactionId, err)
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

	for _, splitTransfer := range splitTransfers {
		feeAmount := util.GetTotalFeeFromTransfers(splitTransfer, receiver)
		onExecutionSuccess, onExecutionFail := s.scheduledTxExecutionCallbacks(event.TransactionId, feeAmount)
		onSuccess, onFail := s.scheduledTxMinedCallbacks(event.TransactionId)

		s.scheduledService.ExecuteScheduledTransferTransaction(event.TransactionId, event.NativeAsset, splitTransfer, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	}
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
	event, err := s.scheduleRepository.GetTransferByTransactionID(id)
	if err != nil {
		s.logger.Errorf("[%s] - failed to get event.", id)
		return "", err
	}

	if event == nil {
		return "", service.ErrNotFound
	}

	return event.TransactionID, nil
}

func (s *Service) scheduledTxExecutionCallbacks(id string, feeAmount string) (onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		s.logger.Debugf("[%s] - Updating db status to Submitted with TransactionID [%s].",
			id,
			transactionID)
		err := s.scheduleRepository.Create(&entity.Schedule{
			ScheduleID:    scheduleID,
			Operation:     schedule.TRANSFER,
			TransactionID: transactionID,
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

func (s *Service) scheduledTxMinedCallbacks(id string) (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {
		s.logger.Debugf("[%s] - Scheduled TX execution successful.", id)
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
