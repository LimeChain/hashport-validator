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
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/burn-event"
	feeRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var (
	hederaAccount = hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 222222,
	}
	tr = transfer.Transfer{
		TransactionId: "0x1aFA123",
		SourceChainId: 80001,
		TargetChainId: 0,
		NativeChainId: 0,
		SourceAsset:   "0x1Af32C",
		TargetAsset:   "0.0.22222",
		NativeAsset:   "0.0.22222",
		Receiver:      "0.0.1337",
		Amount:        "100",
	}
	s                    = &Service{}
	mockBurnEventId      = "some-burnevent-id"
	id                   = "0.0.123123"
	txId                 = "0.0.123123@123123-321321"
	scheduleId           = "0.0.666666"
	feeAmount            = "10000"
	burnEventReceiver, _ = hedera.AccountIDFromString(tr.Receiver)
	burnEventAmount, _   = strconv.ParseInt(tr.Amount, 10, 64)
	entityTransfer       = &entity.Transfer{
		TransactionID: tr.TransactionId,
		SourceChainID: tr.SourceChainId,
		TargetChainID: tr.TargetChainId,
		NativeChainID: tr.NativeChainId,
		SourceAsset:   tr.SourceAsset,
		TargetAsset:   tr.TargetAsset,
		NativeAsset:   tr.NativeAsset,
		Receiver:      tr.Receiver,
		Amount:        tr.Amount,
		Status:        burn_event.StatusInitial,
		Messages:      nil,
		Fee:           entity.Fee{},
		Schedules:     nil,
	}
)

func Test_ProcessEvent(t *testing.T) {
	setup()

	mockFee := int64(12)
	mockRemainder := int64(1)
	mockValidFee := int64(11)
	mockTransfersAfterPreparation := []transfer.Hedera{
		{
			AccountID: burnEventReceiver,
			Amount:    mockRemainder + (mockFee - mockValidFee),
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -burnEventAmount,
		},
	}

	mocks.MTransferRepository.On("Create", &tr).Return(entityTransfer, nil)
	mocks.MFeeService.On("CalculateFee", tr.NativeAsset, burnEventAmount).Return(mockFee, mockRemainder)
	mocks.MDistributorService.On("ValidAmount", mockFee).Return(mockValidFee)
	mocks.MDistributorService.On("CalculateMemberDistribution", mockValidFee).Return([]transfer.Hedera{}, nil)
	mocks.MScheduledService.On("ExecuteScheduledTransferTransaction", tr.TransactionId, tr.NativeAsset, mockTransfersAfterPreparation).Return()

	s.ProcessEvent(tr)
}

func Test_ProcessEventCreateFail(t *testing.T) {
	setup()

	mockFee := int64(11)
	mockRemainder := int64(1)
	mockValidFee := int64(11)
	mockTransfersAfterPreparation := []transfer.Hedera{
		{
			AccountID: burnEventReceiver,
			Amount:    mockRemainder,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -burnEventAmount,
		},
	}

	mocks.MTransferRepository.On("Create", &tr).Return(nil, errors.New("invalid-result"))
	mocks.MFeeService.AssertNotCalled(t, "CalculateFee", tr.NativeAsset, burnEventAmount)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", mockFee)
	mocks.MDistributorService.AssertNotCalled(t, "CalculateMemberDistribution", mockValidFee)
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledTransferTransaction", tr.TransactionId, tr.NativeAsset, mockTransfersAfterPreparation)

	s.ProcessEvent(tr)
}

func Test_ProcessEventCalculateMemberDistributionFails(t *testing.T) {
	setup()

	mockFee := int64(11)
	mockRemainder := int64(1)
	mockValidFee := int64(11)
	mockTransfersAfterPreparation := []transfer.Hedera{
		{
			AccountID: burnEventReceiver,
			Amount:    mockRemainder,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -burnEventAmount,
		},
	}

	mocks.MTransferRepository.On("Create", &tr).Return(entityTransfer, nil)
	mocks.MFeeService.On("CalculateFee", tr.NativeAsset, burnEventAmount).Return(mockFee, mockRemainder)
	mocks.MDistributorService.On("ValidAmount", mockFee).Return(mockValidFee)
	mocks.MDistributorService.On("CalculateMemberDistribution", mockValidFee).Return(nil, errors.New("invalid-result"))
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledTransferTransaction", tr.TransactionId, tr.NativeAsset, mockTransfersAfterPreparation)

	s.ProcessEvent(tr)
}

func Test_New(t *testing.T) {
	setup()
	actualService := NewService(hederaAccount.String(),
		mocks.MTransferRepository,
		mocks.MScheduleRepository,
		mocks.MFeeRepository,
		mocks.MDistributorService,
		mocks.MScheduledService,
		mocks.MFeeService)
	assert.Equal(t, s, actualService)
}

func Test_TransactionID(t *testing.T) {
	setup()

	expectedTransactionId := "0.0.123123-123412.123412"
	mockBurnEventRecord := &entity.Schedule{
		TransactionID: expectedTransactionId,
		TransferID:    sql.NullString{String: expectedTransactionId, Valid: true},
	}

	mocks.MScheduleRepository.On("GetTransferByTransactionID", mockBurnEventId).Return(mockBurnEventRecord, nil)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Nil(t, err)
	assert.Equal(t, expectedTransactionId, actualTransactionId)
}

func Test_TransactionIDRepositoryError(t *testing.T) {
	setup()

	expectedError := errors.New("connection-refused")

	mocks.MScheduleRepository.On("GetTransferByTransactionID", mockBurnEventId).Return(nil, expectedError)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Error(t, expectedError, err)
	assert.Empty(t, actualTransactionId)
}

func Test_TransactionIDNotFound(t *testing.T) {
	setup()

	expectedError := errors.New("not-found")
	mocks.MScheduleRepository.On("GetTransferByTransactionID", mockBurnEventId).Return(nil, expectedError)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Error(t, expectedError, err)
	assert.Empty(t, actualTransactionId)
}

func Test_ScheduledExecutionSuccessCallback(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        feeRepo.StatusSubmitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Operation:     schedule.TRANSFER,
		Status:        schedule.StatusSubmitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil, nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(nil, nil)

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionUpdateStatusFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        feeRepo.StatusSubmitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Operation:     schedule.TRANSFER,
		Status:        schedule.StatusSubmitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(errors.New("update-status-failed"))
	mocks.MFeeRepository.AssertNotCalled(t, "Create", mockEntityFee)

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionCreateFeeFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        feeRepo.StatusSubmitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Operation:     schedule.TRANSFER,
		Status:        schedule.StatusSubmitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(errors.New("create-failed"))

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionFailCallback(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		Amount:        feeAmount,
		Status:        feeRepo.StatusFailed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		Status:        schedule.StatusFailed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(nil)

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onError(txId)
}

func Test_ScheduledExecutionFailedUpdateStatusFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        feeRepo.StatusFailed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		Status:        schedule.StatusFailed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(errors.New("update-status-failed"))
	mocks.MFeeRepository.AssertNotCalled(t, "Create", mockEntityFee)

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onError(txId)
}

func Test_ScheduledExecutionFailedCreateFeeFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		Amount:        feeAmount,
		Status:        feeRepo.StatusFailed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		Status:        schedule.StatusFailed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(errors.New("create-failed"))

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onError(txId)
}

func Test_ScheduledTxMinedExecutionSuccessCallback(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(nil)

	onSuccess, _ := s.scheduledTxMinedCallbacks(id)
	onSuccess(txId)
}

func Test_ScheduledTxMinedExecutionSuccessUpdateStatusFails(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(errors.New("update-status-fail"))
	mocks.MFeeRepository.AssertNotCalled(t, "UpdateStatusCompleted", txId)

	onSuccess, _ := s.scheduledTxMinedCallbacks(id)
	onSuccess(txId)
}

func Test_ScheduledTxMinedExecutionFailCallback(t *testing.T) {
	setup()

	mocks.MScheduleRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(nil)

	_, onFail := s.scheduledTxMinedCallbacks(id)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionFailUpdateStatusFailedFails(t *testing.T) {
	setup()

	mocks.MScheduleRepository.On("UpdateStatusFailed", id).Return(errors.New("update-status-fail"))
	mocks.MFeeRepository.AssertNotCalled(t, "UpdateStatusFailed", txId)

	_, onFail := s.scheduledTxMinedCallbacks(id)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionFailFeeUpdateFails(t *testing.T) {
	setup()

	mocks.MScheduleRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(errors.New("update-fail"))

	_, onFail := s.scheduledTxMinedCallbacks(id)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionSuccessFeeUpdateFails(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(errors.New("update-fail"))

	onSuccess, _ := s.scheduledTxMinedCallbacks(id)
	onSuccess(txId)
}

func setup() {
	mocks.Setup()
	s = &Service{
		bridgeAccount:      hederaAccount,
		feeRepository:      mocks.MFeeRepository,
		repository:         mocks.MTransferRepository,
		scheduleRepository: mocks.MScheduleRepository,
		distributorService: mocks.MDistributorService,
		feeService:         mocks.MFeeService,
		scheduledService:   mocks.MScheduledService,
		logger:             config.GetLoggerFor("Burn Event Service"),
	}
}
