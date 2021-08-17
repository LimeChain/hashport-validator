package lock_event

import (
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var (
	hederaAccount = hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 222222,
	}
	lockEvent = lock_event.LockEvent{
		Id:     "0.0.444444",
		Amount: 111,
		Recipient: hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 222222,
		},
		NativeAsset:   "0.0.222222",
		WrappedAsset:  "0.0.000000",
		SourceChainId: big.NewInt(3),
		TargetChainId: big.NewInt(0),
	}
	s               = &Service{}
	mockLockEventId = "some-lock-event-id"
	id              = "0.0.123123"
	txId            = "0.0.123123@123123-321321"
	scheduleId      = "0.0.666666"
	feeAmount       = "10000"
)

func Test_New(t *testing.T) {
	setup()
	actualService := NewService(
		hederaAccount.String(),
		mocks.MLockEventRepository,
		mocks.MScheduledService)
	assert.Equal(t, s, actualService)
}

func Test_ProcessEventFailsOnCreate(t *testing.T) {
	setup()
	actualService := NewService(
		hederaAccount.String(),
		mocks.MLockEventRepository,
		mocks.MScheduledService)

	mocks.MLockEventRepository.On("Create", lockEvent.Id, lockEvent.Amount, lockEvent.Recipient.String(), lockEvent.NativeAsset, lockEvent.WrappedAsset, lockEvent.SourceChainId.Int64(), lockEvent.TargetChainId.Int64()).Return(errors.New("e"))
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledMintTransaction")
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledTransferTransaction")

	actualService.ProcessEvent(lockEvent)
}

func Test_ProcessEventFailsOnScheduleMint(t *testing.T) {
	setup()
	actualService := NewService(
		hederaAccount.String(),
		mocks.MLockEventRepository,
		mocks.MScheduledService)

	mocks.MLockEventRepository.On("Create", lockEvent.Id, lockEvent.Amount, lockEvent.Recipient.String(), lockEvent.NativeAsset, lockEvent.WrappedAsset, lockEvent.SourceChainId.Int64(), lockEvent.TargetChainId.Int64()).Return(nil)
	mocks.MScheduledService.On("ExecuteScheduledMintTransaction", lockEvent.Id, lockEvent.WrappedAsset, lockEvent.Amount).Return(errors.New("e"))
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledTransferTransaction")

	actualService.ProcessEvent(lockEvent)
}

func Test_ProcessEventFailsOnScheduleTransfer(t *testing.T) {
	setup()
	actualService := NewService(
		hederaAccount.String(),
		mocks.MLockEventRepository,
		mocks.MScheduledService)

	mockTransfers := []transfer.Hedera{
		{
			AccountID: lockEvent.Recipient,
			Amount:    lockEvent.Amount,
		},
		{
			AccountID: s.bridgeAccount,
			Amount:    -lockEvent.Amount,
		},
	}

	mocks.MLockEventRepository.On("Create", lockEvent.Id, lockEvent.Amount, lockEvent.Recipient.String(), lockEvent.NativeAsset, lockEvent.WrappedAsset, lockEvent.SourceChainId.Int64(), lockEvent.TargetChainId.Int64()).Return(nil)
	mocks.MScheduledService.On("ExecuteScheduledMintTransaction", lockEvent.Id, lockEvent.WrappedAsset, lockEvent.Amount).Return(nil)
	mocks.MScheduledService.On("ExecuteScheduledTransferTransaction", lockEvent.Id, lockEvent.WrappedAsset, mockTransfers).Return()

	actualService.ProcessEvent(lockEvent)
}

func Test_ScheduledTokenMintExecutionSuccessCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusScheduledTokenMintSubmitted", id, scheduleId, txId).Return(nil)

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, SCHEDULED_MINT_TYPE)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledTokenMintExecutionThrowsErrorCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusScheduledTokenMintSubmitted", id, scheduleId, txId).Return(errors.New("e"))

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, SCHEDULED_MINT_TYPE)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledTransferExecutionSuccessCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusScheduledTokenTransferSubmitted", id, scheduleId, txId).Return(nil)

	onSuccess, _ := s.scheduledTxExecutionCallbacks(id, SCHEDULED_TRANSFER_TYPE)
	onSuccess(txId, scheduleId)
}

func Test_ScheduledExecutionFailCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusFailed", id).Return(nil)

	_, onFail := s.scheduledTxExecutionCallbacks(id, SCHEDULED_TRANSFER_TYPE)
	onFail(txId)
}

func Test_ScheduledExecutionFailThrowsErrorCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusFailed", id).Return(errors.New("e"))

	_, onFail := s.scheduledTxExecutionCallbacks(id, SCHEDULED_TRANSFER_TYPE)
	onFail(txId)
}

func Test_ScheduledTokenMintMineSuccessCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusScheduledTokenMintCompleted", id).Return(nil)

	onSuccess, _ := s.scheduledTxMinedCallbacks(id, SCHEDULED_MINT_TYPE)
	onSuccess(txId)
}

func Test_ScheduledTokenMintMineThrowsErrorCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusScheduledTokenMintCompleted", id).Return(errors.New("e"))

	onSuccess, _ := s.scheduledTxMinedCallbacks(id, SCHEDULED_MINT_TYPE)
	onSuccess(txId)
}

func Test_ScheduledTransferMineSuccessCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusCompleted", id).Return(nil)

	onSuccess, _ := s.scheduledTxMinedCallbacks(id, SCHEDULED_TRANSFER_TYPE)
	onSuccess(txId)
}

func Test_ScheduledMinedExecutionFailCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusFailed", id).Return(nil)

	_, onFail := s.scheduledTxMinedCallbacks(id, SCHEDULED_TRANSFER_TYPE)
	onFail(txId)
}

func Test_ScheduledMinedFailThrowsErrorCallback(t *testing.T) {
	setup()

	mocks.MLockEventRepository.On("UpdateStatusFailed", id).Return(errors.New("e"))

	_, onFail := s.scheduledTxMinedCallbacks(id, SCHEDULED_TRANSFER_TYPE)
	onFail(txId)
}

func setup() {
	mocks.Setup()
	s = &Service{
		bridgeAccount:    hederaAccount,
		repository:       mocks.MLockEventRepository,
		scheduledService: mocks.MScheduledService,
		logger:           config.GetLoggerFor("Lock Event Service"),
	}
}
