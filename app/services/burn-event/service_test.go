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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
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
		Status:        status.Initial,
		Messages:      nil,
		Fees:          []entity.Fee{},
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

	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(entityTransfer, nil)
	mocks.MFeeService.On("CalculateFee", tr.NativeAsset, burnEventAmount).Return(mockFee, mockRemainder)
	mocks.MDistributorService.On("ValidAmount", mockFee).Return(mockValidFee)
	mocks.MDistributorService.On("CalculateMemberDistribution", mockValidFee).Return([]transfer.Hedera{}, nil)
	mocks.MTransferRepository.On("UpdateFee", tr.TransactionId, strconv.FormatInt(mockValidFee, 10)).Return(nil)
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

	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(nil, errors.New("invalid-result"))
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

	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(entityTransfer, nil)
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
		mocks.MFeeService,
		mocks.MTransferService,
		mocks.MPrometheusService)
	assert.Equal(t, s, actualService)
}

func Test_TransactionID(t *testing.T) {
	setup()

	expectedTransactionId := "0.0.123123-123412.123412"
	mockBurnEventRecord := &entity.Schedule{
		TransactionID: expectedTransactionId,
		TransferID:    sql.NullString{String: expectedTransactionId, Valid: true},
	}

	mocks.MScheduleRepository.On("GetReceiverTransferByTransactionID", mockBurnEventId).Return(mockBurnEventRecord, nil)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Nil(t, err)
	assert.Equal(t, expectedTransactionId, actualTransactionId)
}

func Test_TransactionIDRepositoryError(t *testing.T) {
	setup()

	expectedError := errors.New("connection-refused")

	mocks.MScheduleRepository.On("GetReceiverTransferByTransactionID", mockBurnEventId).Return(nil, expectedError)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Error(t, expectedError, err)
	assert.Empty(t, actualTransactionId)
}

func Test_TransactionIDNotFound(t *testing.T) {
	setup()

	expectedError := errors.New("not-found")
	mocks.MScheduleRepository.On("GetReceiverTransferByTransactionID", mockBurnEventId).Return(nil, expectedError)

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
		Status:        status.Submitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Operation:     schedule.TRANSFER,
		HasReceiver:   true,
		Status:        status.Submitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil, nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(nil, nil)

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, feeAmount, true)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionUpdateStatusFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        status.Submitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Operation:     schedule.TRANSFER,
		Status:        status.Submitted,
		HasReceiver:   true,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(errors.New("update-status-failed"))
	mocks.MFeeRepository.AssertNotCalled(t, "Create", mockEntityFee)

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, feeAmount, true)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionCreateFeeFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        status.Submitted,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Operation:     schedule.TRANSFER,
		Status:        status.Submitted,
		HasReceiver:   true,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(errors.New("create-failed"))

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, feeAmount, true)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionFailCallback(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		Amount:        feeAmount,
		Status:        status.Failed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		Status:        status.Failed,
		HasReceiver:   true,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil)
	mocks.MTransferRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(nil)

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount, true)
	onError(txId)
}

func Test_ScheduledExecutionFailedUpdateStatusFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		Amount:        feeAmount,
		Status:        status.Failed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		Status:        status.Failed,
		HasReceiver:   true,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(errors.New("update-status-failed"))
	mocks.MFeeRepository.AssertNotCalled(t, "Create", mockEntityFee)

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount, true)
	onError(txId)
}

func Test_ScheduledExecutionFailedCreateFeeFails(t *testing.T) {
	setup()

	mockEntityFee := &entity.Fee{
		TransactionID: txId,
		Amount:        feeAmount,
		Status:        status.Failed,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mockEntitySchedule := &entity.Schedule{
		TransactionID: txId,
		Status:        status.Failed,
		HasReceiver:   true,
		TransferID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MScheduleRepository.On("Create", mockEntitySchedule).Return(nil)
	mocks.MTransferRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(errors.New("create-failed"))

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount, true)
	onError(txId)
}

func Test_ScheduledTxMinedExecutionSuccessCallback(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(nil)

	resultPerTransfer := true
	wg := new(sync.WaitGroup)
	wg.Add(1)
	onSuccess, _ := s.scheduledTxMinedCallbacks(id, &resultPerTransfer, wg)
	onSuccess(txId)
}

func Test_ScheduledTxMinedExecutionSuccessUpdateStatusFails(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(errors.New("update-status-fail"))
	mocks.MFeeRepository.AssertNotCalled(t, "UpdateStatusCompleted", txId)

	resultPerTransfer := true
	wg := new(sync.WaitGroup)
	wg.Add(1)
	onSuccess, _ := s.scheduledTxMinedCallbacks(id, &resultPerTransfer, wg)
	onSuccess(txId)
}

func Test_ScheduledTxMinedExecutionFailCallback(t *testing.T) {
	setup()

	mocks.MScheduleRepository.On("UpdateStatusFailed", txId).Return(nil)
	mocks.MTransferRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(nil)

	resultPerTransfer := true
	wg := new(sync.WaitGroup)
	wg.Add(1)
	_, onFail := s.scheduledTxMinedCallbacks(id, &resultPerTransfer, wg)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionFailUpdateStatusFailedFails(t *testing.T) {
	setup()

	mocks.MScheduleRepository.On("UpdateStatusFailed", txId).Return(errors.New("update-status-fail"))
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateStatusFailed", id)
	mocks.MFeeRepository.AssertNotCalled(t, "UpdateStatusFailed", txId)

	resultPerTransfer := true
	wg := new(sync.WaitGroup)
	wg.Add(1)
	_, onFail := s.scheduledTxMinedCallbacks(id, &resultPerTransfer, wg)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionFailFeeUpdateFails(t *testing.T) {
	setup()

	mocks.MScheduleRepository.On("UpdateStatusFailed", txId).Return(nil)
	mocks.MTransferRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(errors.New("update-fail"))

	resultPerTransfer := true
	wg := new(sync.WaitGroup)
	wg.Add(1)
	_, onFail := s.scheduledTxMinedCallbacks(id, &resultPerTransfer, wg)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionSuccessFeeUpdateFails(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", txId).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(errors.New("update-fail"))

	resultPerTransfer := true
	wg := new(sync.WaitGroup)
	wg.Add(1)
	onSuccess, _ := s.scheduledTxMinedCallbacks(id, &resultPerTransfer, wg)
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
		transferService:    mocks.MTransferService,
		prometheusService:  mocks.MPrometheusService,
		logger:             config.GetLoggerFor("Burn Event Service"),
	}
}
