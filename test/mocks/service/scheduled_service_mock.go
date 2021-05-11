package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/mock"
)

type MockScheduledService struct {
	mock.Mock
}

func (mss *MockScheduledService) Execute(id, nativeAsset string, transfers []transfer.Hedera, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	// TODO: Find a way to mock these functions properly, without rewriting them once more in the unit test file.
	mss.Called(id, nativeAsset, transfers)
}
