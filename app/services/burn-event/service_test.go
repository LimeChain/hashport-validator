package burn_event

import (
	"database/sql"
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	s             = &Service{}
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
	mockBurnEventId := "some-burnevent-id"
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
	mockBurnEventId := "some-burnevent-id"

	mocks.MBurnEventRepository.On("Get", mockBurnEventId).Return(nil, expectedError)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Error(t, expectedError, err)
	assert.Empty(t, actualTransactionId)
}

func Test_TransactionIDNotFound(t *testing.T) {
	setup()

	expectedError := errors.New("not found")
	mockBurnEventId := "some-burnevent-id"

	mocks.MBurnEventRepository.On("Get", mockBurnEventId).Return(nil, nil)

	actualTransactionId, err := s.TransactionID(mockBurnEventId)
	assert.Error(t, expectedError, err)
	assert.Empty(t, actualTransactionId)
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
