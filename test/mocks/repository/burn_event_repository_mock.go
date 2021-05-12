package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockBurnEventRepository struct {
	mock.Mock
}

func (berm *MockBurnEventRepository) Create(id string, amount int64, recipient string) error {
	args := berm.Called(id, amount, recipient)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockBurnEventRepository) UpdateStatusSubmitted(ethTxHash, scheduleID, transactionId string) error {
	args := berm.Called(ethTxHash, scheduleID, transactionId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockBurnEventRepository) UpdateStatusCompleted(id string) error {
	args := berm.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockBurnEventRepository) UpdateStatusFailed(id string) error {
	args := berm.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockBurnEventRepository) Get(id string) (*entity.BurnEvent, error) {
	args := berm.Called(id)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.BurnEvent), nil
	}
	return nil, args.Get(1).(error)
}
