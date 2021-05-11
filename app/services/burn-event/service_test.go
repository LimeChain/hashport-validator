package burn_event

import (
	"database/sql"
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	feeRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/fee"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	hederaAccount = hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 222222,
	}
	burnEvent = burn_event.BurnEvent{
		Amount: 111,
		Recipient: hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 222222,
		},
		Id:           "0.0.444444",
		NativeAsset:  "0.0.222222",
		WrappedAsset: "0.0.000000",
	}
	s               = &Service{}
	mockBurnEventId = "some-burnevent-id"
	id              = "0.0.123123"
	txId            = "0.0.123123@123123-321321"
	scheduleId      = "0.0.666666"
	feeAmount       = "10000"
)

func Test_ProcessEvent(t *testing.T) {
	setup()

	mockFee := int64(12)
	mockRemainder := int64(1)
	mockValidFee := int64(11)
	mockTransfersAfterPreparation := []transfer.Hedera{
		{
			AccountID: burnEvent.Recipient,
			Amount:    mockRemainder + (mockFee - mockValidFee),
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -burnEvent.Amount,
		},
	}

	mocks.MBurnEventRepository.On("Create", burnEvent.Id, burnEvent.Amount, burnEvent.Recipient.String()).Return(nil)
	mocks.MFeeService.On("CalculateFee", burnEvent.Amount).Return(mockFee, mockRemainder)
	mocks.MDistributorService.On("ValidAmount", mockFee).Return(mockValidFee)
	mocks.MDistributorService.On("CalculateMemberDistribution", mockValidFee).Return([]transfer.Hedera{}, nil)
	mocks.MScheduledService.On("Execute", burnEvent.Id, burnEvent.NativeAsset, mockTransfersAfterPreparation).Return()

	s.ProcessEvent(burnEvent)
}

func Test_ProcessEventCreateFail(t *testing.T) {
	setup()

	mockFee := int64(11)
	mockRemainder := int64(1)
	mockValidFee := int64(11)
	mockTransfersAfterPreparation := []transfer.Hedera{
		{
			AccountID: burnEvent.Recipient,
			Amount:    mockRemainder,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -burnEvent.Amount,
		},
	}

	mocks.MBurnEventRepository.On("Create", burnEvent.Id, burnEvent.Amount, burnEvent.Recipient.String()).Return(errors.New("invalid-result"))
	mocks.MFeeService.AssertNotCalled(t, "CalculateFee", burnEvent.Amount)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", mockFee)
	mocks.MDistributorService.AssertNotCalled(t, "CalculateMemberDistribution", mockValidFee)
	mocks.MScheduledService.AssertNotCalled(t, "Execute", burnEvent.Id, burnEvent.NativeAsset, mockTransfersAfterPreparation)

	s.ProcessEvent(burnEvent)
}

func Test_ProcessEventCalculateMemberDistributionFails(t *testing.T) {
	setup()

	mockFee := int64(11)
	mockRemainder := int64(1)
	mockValidFee := int64(11)
	mockTransfersAfterPreparation := []transfer.Hedera{
		{
			AccountID: burnEvent.Recipient,
			Amount:    mockRemainder,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -burnEvent.Amount,
		},
	}

	mocks.MBurnEventRepository.On("Create", burnEvent.Id, burnEvent.Amount, burnEvent.Recipient.String()).Return(nil)
	mocks.MFeeService.On("CalculateFee", burnEvent.Amount).Return(mockFee, mockRemainder)
	mocks.MDistributorService.On("ValidAmount", mockFee).Return(mockValidFee)
	mocks.MDistributorService.On("CalculateMemberDistribution", mockValidFee).Return(nil, errors.New("invalid-result"))
	mocks.MScheduledService.AssertNotCalled(t, "Execute", burnEvent.Id, burnEvent.NativeAsset, mockTransfersAfterPreparation)

	s.ProcessEvent(burnEvent)
}

func Test_New(t *testing.T) {
	setup()
	actualService := NewService(hederaAccount.String(), mocks.MBurnEventRepository, mocks.MFeeRepository, mocks.MDistributorService, mocks.MScheduledService, mocks.MFeeService)
	assert.Equal(t, s, actualService)
}

func Test_TransactionID(t *testing.T) {
	setup()

	expectedTransactionId := "0.0.123123-123412.123412"
	mockBurnEventRecord := &entity.BurnEvent{
		TransactionId: sql.NullString{String: expectedTransactionId, Valid: true},
	}

	mocks.MBurnEventRepository.On("Get", mockBurnEventId).Return(mockBurnEventRecord, nil)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Nil(t, err)
	assert.Equal(t, expectedTransactionId, actualTransactionId)
}

func Test_TransactionIDRepositoryError(t *testing.T) {
	setup()

	expectedError := errors.New("connection-refused")

	mocks.MBurnEventRepository.On("Get", mockBurnEventId).Return(nil, expectedError)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Error(t, expectedError, err)
	assert.Empty(t, actualTransactionId)
}

func Test_TransactionIDNotFound(t *testing.T) {
	setup()

	expectedError := errors.New("not found")
	mocks.MBurnEventRepository.On("Get", mockBurnEventId).Return(nil, nil)

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
		BurnEventID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MBurnEventRepository.On("UpdateStatusSubmitted", id, scheduleId, txId).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(nil)

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
		BurnEventID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MBurnEventRepository.On("UpdateStatusSubmitted", id, scheduleId, txId).Return(errors.New("update-status-failed"))
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
		BurnEventID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MBurnEventRepository.On("UpdateStatusSubmitted", id, scheduleId, txId).Return(nil)
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
		BurnEventID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MBurnEventRepository.On("UpdateStatusFailed", id).Return(nil)
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
		BurnEventID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MBurnEventRepository.On("UpdateStatusFailed", id).Return(errors.New("update-status-failed"))
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
		BurnEventID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	mocks.MBurnEventRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("Create", mockEntityFee).Return(errors.New("create-failed"))

	_, onError := s.scheduledTxExecutionCallbacks(id, feeAmount)
	onError(txId)
}

func Test_ScheduledTxMinedExecutionSuccessCallback(t *testing.T) {
	setup()

	mocks.MBurnEventRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(nil)

	onSuccess, _ := s.scheduledTxMinedCallbacks(id)
	onSuccess(txId)
}

func Test_ScheduledTxMinedExecutionSuccessUpdateStatusFails(t *testing.T) {
	setup()

	mocks.MBurnEventRepository.On("UpdateStatusCompleted", id).Return(errors.New("update-status-fail"))
	mocks.MFeeRepository.AssertNotCalled(t, "UpdateStatusCompleted", txId)

	onSuccess, _ := s.scheduledTxMinedCallbacks(id)
	onSuccess(txId)
}

func Test_ScheduledTxMinedExecutionFailCallback(t *testing.T) {
	setup()

	mocks.MBurnEventRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(nil)

	_, onFail := s.scheduledTxMinedCallbacks(id)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionFailUpdateStatusFailedFails(t *testing.T) {
	setup()

	mocks.MBurnEventRepository.On("UpdateStatusFailed", id).Return(errors.New("update-status-fail"))
	mocks.MFeeRepository.AssertNotCalled(t, "UpdateStatusFailed", txId)

	_, onFail := s.scheduledTxMinedCallbacks(id)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionFailFeeUpdateFails(t *testing.T) {
	setup()

	mocks.MBurnEventRepository.On("UpdateStatusFailed", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusFailed", txId).Return(errors.New("update-fail"))

	_, onFail := s.scheduledTxMinedCallbacks(id)
	onFail(txId)
}

func Test_ScheduledTxMinedExecutionSuccessFeeUpdateFails(t *testing.T) {
	setup()

	mocks.MBurnEventRepository.On("UpdateStatusCompleted", id).Return(nil)
	mocks.MFeeRepository.On("UpdateStatusCompleted", txId).Return(errors.New("update-fail"))

	onSuccess, _ := s.scheduledTxMinedCallbacks(id)
	onSuccess(txId)
}

func setup() {
	mocks.Setup()
	s = &Service{
		bridgeAccount:      hederaAccount,
		feeRepository:      mocks.MFeeRepository,
		repository:         mocks.MBurnEventRepository,
		distributorService: mocks.MDistributorService,
		feeService:         mocks.MFeeService,
		scheduledService:   mocks.MScheduledService,
		logger:             config.GetLoggerFor("Burn Event Service"),
	}
}
