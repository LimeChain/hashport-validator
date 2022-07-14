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

package hedera

import (
	"database/sql"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	syncHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/sync"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	log "github.com/sirupsen/logrus"
	"sync"
)

func ScheduledNftTxExecutionCallbacks(
	transferRepository repository.Transfer,
	scheduleRepository repository.Schedule,
	logger *log.Entry,
	id string,
	hasReceiver bool,
	statusResult *string,
	wg *sync.WaitGroup,
) (onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		logger.Debugf("[%s] - Updating db status to Submitted with TransactionID [%s].",
			id,
			transactionID)
		err := scheduleRepository.Create(&entity.Schedule{
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
			defer wg.Done()
			*statusResult = syncHelper.FAIL
			logger.Errorf(
				"[%s] - Failed to update submitted status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				id, transactionID, scheduleID, err)

			return
		}
	}

	onExecutionFail = func(transactionID string) {
		defer wg.Done()
		*statusResult = syncHelper.FAIL
		err := scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			Status:        status.Failed,
			HasReceiver:   hasReceiver,
			TransferID: sql.NullString{
				String: id,
				Valid:  true,
			},
		})
		if err != nil {
			logger.Errorf("[%s] - Failed to update status failed. Error [%s].", id, err)
			return
		}

		err = transferRepository.UpdateStatusFailed(id)
		if err != nil {
			logger.Errorf("[%s] - Failed to update status failed. Error [%s].", id, err)
			return
		}
	}

	return onExecutionSuccess, onExecutionFail
}

func ScheduledNftTxMinedCallbacks(
	transferRepository repository.Transfer,
	scheduleRepository repository.Schedule,
	logger *log.Entry,
	id string,
	status *string,
	wg *sync.WaitGroup,
) (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {
		defer wg.Done()
		logger.Debugf("[%s] - Scheduled TX execution successful.", id)
		err := transferRepository.UpdateStatusCompleted(id)
		if err != nil {
			*status = syncHelper.FAIL
			logger.Errorf("[%s] - Failed to update status completed. Error [%s].", id, err)
			return
		}
		err = scheduleRepository.UpdateStatusCompleted(transactionID)
		if err != nil {
			*status = syncHelper.FAIL
			logger.Errorf("[%s] - Failed to update status completed. Error [%s].", transactionID, err)
			return
		}
		*status = syncHelper.DONE
	}

	onFail = func(transactionID string) {
		defer wg.Done()
		*status = syncHelper.FAIL
		logger.Debugf("[%s] - Scheduled TX execution has failed.", id)
		err := scheduleRepository.UpdateStatusFailed(transactionID)
		if err != nil {
			logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", id, err)
			return
		}

		err = transferRepository.UpdateStatusFailed(id)
		if err != nil {
			logger.Errorf("[%s] - Failed to update status failed. Error [%s].", transactionID, err)
			return
		}
	}

	return onSuccess, onFail
}
