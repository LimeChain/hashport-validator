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
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	log "github.com/sirupsen/logrus"
	"sync"
	"testing"
)

var (
	logger                   *log.Entry
	transactionId            = "0.0.1234"
	scheduleId               = "0.0.5678"
	statusResult             *string
	wg                       *sync.WaitGroup
	createdScheduleOnSuccess = &entity.Schedule{
		TransactionID: transactionId,
		ScheduleID:    scheduleId,
		HasReceiver:   true,
		Operation:     schedule.TRANSFER,
		Status:        status.Submitted,
		TransferID: sql.NullString{
			String: transactionId,
			Valid:  true,
		},
	}
	createdScheduleOnError = *createdScheduleOnSuccess
	error                  = errors.New("some-error")
)

func Test_ScheduledNftTxExecutionCallbacks(t *testing.T) {
	setupNftTest(true)

	onSuccess, onFail := ScheduledNftTxExecutionCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, true, statusResult, wg)

	onSuccess(transactionId, scheduleId)
	onFail(transactionId)
}

func Test_ScheduledNftTxExecutionCallbacks_ErrScheduleCreateOnSuccess(t *testing.T) {
	setupNftTest(false)

	mocks.MScheduleRepository.On("Create", createdScheduleOnSuccess).Return(error)

	onSuccess, _ := ScheduledNftTxExecutionCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, true, statusResult, wg)

	onSuccess(transactionId, scheduleId)
}

func Test_ScheduledNftTxExecutionCallbacks_ErrScheduleCreateOnFail(t *testing.T) {
	setupNftTest(false)
	updateFieldsForCreatedScheduleOnError()
	mocks.MScheduleRepository.On("Create", &createdScheduleOnError).Return(error)

	_, onFail := ScheduledNftTxExecutionCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, true, statusResult, wg)

	onFail(transactionId)
}

func Test_ScheduledNftTxExecutionCallbacks_ErrUpdateStatusFailedOnFail(t *testing.T) {
	setupNftTest(false)
	updateFieldsForCreatedScheduleOnError()
	mocks.MScheduleRepository.On("Create", &createdScheduleOnError).Return(nil)
	mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(error)

	_, onFail := ScheduledNftTxExecutionCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, true, statusResult, wg)

	onFail(transactionId)
}

func Test_ScheduledNftTxMinedCallbacks(t *testing.T) {
	setupNftTest(true)
	mocks.MTransferRepository.On("UpdateStatusCompleted", transactionId).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", transactionId).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusFailed", transactionId).Return(nil)
	wg.Add(1)

	onSuccess, onFail := ScheduledNftTxMinedCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, statusResult, wg)

	onSuccess(transactionId)
	onFail(transactionId)
}

func Test_ScheduledNftTxMinedCallbacks_ErrTransferUpdateStatusCompletedOnSuccess(t *testing.T) {
	setupNftTest(true)
	mocks.MTransferRepository.On("UpdateStatusCompleted", transactionId).Return(error)
	wg.Add(1)

	onSuccess, _ := ScheduledNftTxMinedCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, statusResult, wg)

	onSuccess(transactionId)
}

func Test_ScheduledNftTxMinedCallbacks_ErrScheduleUpdateStatusCompletedOnSuccess(t *testing.T) {
	setupNftTest(true)
	mocks.MTransferRepository.On("UpdateStatusCompleted", transactionId).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", transactionId).Return(error)
	wg.Add(1)

	onSuccess, _ := ScheduledNftTxMinedCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, statusResult, wg)

	onSuccess(transactionId)
}

func Test_ScheduledNftTxMinedCallbacks_ErrScheduleUpdateStatusCompletedOnFail(t *testing.T) {
	setupNftTest(true)
	mocks.MScheduleRepository.On("UpdateStatusFailed", transactionId).Return(error)
	wg.Add(1)

	_, onFail := ScheduledNftTxMinedCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, statusResult, wg)

	onFail(transactionId)
}

func Test_ScheduledNftTxMinedCallbacks_ErrTransferUpdateStatusCompletedOnFail(t *testing.T) {
	setupNftTest(false)
	mocks.MScheduleRepository.On("UpdateStatusFailed", transactionId).Return(nil)
	mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(error)
	wg.Add(1)

	_, onFail := ScheduledNftTxMinedCallbacks(mocks.MTransferRepository, mocks.MScheduleRepository, logger, transactionId, statusResult, wg)

	onFail(transactionId)
}

func setupNftTest(withMocks bool) {
	mocks.Setup()
	wg = new(sync.WaitGroup)
	wg.Add(1)
	statusResult = new(string)
	logger = log.WithField("context", "Test")

	if withMocks {
		updateFieldsForCreatedScheduleOnError()
		mocks.MScheduleRepository.On("Create", createdScheduleOnSuccess).Return(nil).Once()
		mocks.MScheduleRepository.On("Create", &createdScheduleOnError).Return(nil).Once()
		mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(nil)
	}
}

func updateFieldsForCreatedScheduleOnError() {
	createdScheduleOnError = *createdScheduleOnSuccess
	createdScheduleOnError.ScheduleID = ""
	createdScheduleOnError.Operation = ""
	createdScheduleOnError.Status = status.Failed
}
