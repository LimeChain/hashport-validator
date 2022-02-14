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

package recovery

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	r Recovery
)

func Test_New(t *testing.T) {
	setup()
	assert.Equal(t, &r, New(mocks.MFeeRepository, mocks.MScheduleRepository, mocks.MHederaMirrorClient))
}

func Test_CheckSubmittedFees(t *testing.T) {
	setup()
	mocks.MFeeRepository.On("GetAllSubmittedIds").Return([]*entity.Fee{{
		TransactionID: "some-tx-id",
		ScheduleID:    "some-schedule-id",
		Amount:        "100",
		Status:        "some-status",
	}}, nil)
	mocks.MHederaMirrorClient.On("WaitForScheduledTransaction", "some-tx-id", mock.Anything, mock.Anything)
	r.checkSubmittedFees()
	mocks.MHederaMirrorClient.AssertCalled(t, "WaitForScheduledTransaction", "some-tx-id", mock.Anything, mock.Anything)
}

func Test_CheckSubmittedFees_GetAllSubmitedIds_Fails(t *testing.T) {
	setup()
	mocks.MFeeRepository.On("GetAllSubmittedIds").Return(nil, errors.New("some-error"))
	r.checkSubmittedFees()
	mocks.MHederaMirrorClient.AssertNotCalled(t, "WaitForScheduledTransaction", mock.Anything, mock.Anything, mock.Anything)
}

func Test_CheckSubmittedSchedules(t *testing.T) {
	setup()
	mocks.MScheduleRepository.On("GetAllSubmittedIds").Return([]*entity.Schedule{{
		TransactionID: "some-tx-id",
		ScheduleID:    "some-schedule-id",
		Operation:     "some-operation",
		Status:        "some-status",
	}}, nil)
	mocks.MHederaMirrorClient.On("WaitForScheduledTransaction", "some-tx-id", mock.Anything, mock.Anything)
	r.checkSubmittedSchedules()
	mocks.MHederaMirrorClient.AssertCalled(t, "WaitForScheduledTransaction", "some-tx-id", mock.Anything, mock.Anything)
}

func Test_CheckSubmittedSchedules_GetAllSubmitedIds_Fails(t *testing.T) {
	setup()
	mocks.MScheduleRepository.On("GetAllSubmittedIds").Return(nil, errors.New("some-error"))
	r.checkSubmittedSchedules()
	mocks.MHederaMirrorClient.AssertNotCalled(t, "WaitForScheduledTransaction", mock.Anything, mock.Anything, mock.Anything)
}

func Test_CallBacks_IsFee(t *testing.T) {
	setup()
	txId := "some-id"
	onSuccess, onRevert := r.callbacks(txId, true)

	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(nil)
	onSuccess()
	mocks.MFeeRepository.AssertCalled(t, "UpdateStatusCompleted", txId)

	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(nil)
	onRevert()
	mocks.MFeeRepository.AssertCalled(t, "UpdateStatusFailed", txId)
}

func Test_CallBacks_IsFee_Fails(t *testing.T) {
	setup()
	txId := "some-id"
	onSuccess, onRevert := r.callbacks(txId, true)

	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(errors.New("some-error"))
	onSuccess()
	mocks.MFeeRepository.AssertCalled(t, "UpdateStatusCompleted", txId)

	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(errors.New("some-error"))
	onRevert()
	mocks.MFeeRepository.AssertCalled(t, "UpdateStatusFailed", txId)
}

func Test_CallBacks_IsNotFee(t *testing.T) {
	setup()
	txId := "some-id"
	onSuccess, onRevert := r.callbacks(txId, false)

	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(nil)
	onSuccess()
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusCompleted", txId)

	mocks.MScheduleRepository.On("UpdateStatusFailed", txId).Return(nil)
	onRevert()
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusFailed", txId)
}

func Test_CallBacks_IsNotFee_Fails(t *testing.T) {
	setup()
	txId := "some-id"
	onSuccess, onRevert := r.callbacks(txId, false)

	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(errors.New("some-error"))
	onSuccess()
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusCompleted", txId)

	mocks.MScheduleRepository.On("UpdateStatusFailed", txId).Return(errors.New("some-error"))
	onRevert()
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusFailed", txId)
}

func setup() {
	mocks.Setup()
	r = Recovery{
		feeRepository:      mocks.MFeeRepository,
		scheduleRepository: mocks.MScheduleRepository,
		mirrorClient:       mocks.MHederaMirrorClient,
		logger:             config.GetLoggerFor("Recovery"),
	}
}
