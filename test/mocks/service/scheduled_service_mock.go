package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/mock"
)

type MockScheduledService struct {
	mock.Mock
}

func (mss *MockScheduledService) ExecuteScheduledMintTransaction(id, asset string, amount int64, onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) error {
	args := mss.Called(id, asset, amount)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mss *MockScheduledService) ExecuteScheduledTransferTransaction(id, nativeAsset string, transfers []transfer.Hedera, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	// TODO: Find a way to mock these functions properly, without rewriting them once more in the unit test file.
	mss.Called(id, nativeAsset, transfers)
}
