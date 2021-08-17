package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockLockEventRepository struct {
	mock.Mock
}

func (berm *MockLockEventRepository) Create(id string, amount int64, recipient, nativeAsset, wrappedAsset string, sourceChainId, targetChainId int64) error {
	args := berm.Called(id, amount, recipient, nativeAsset, wrappedAsset, sourceChainId, targetChainId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusScheduledTokenMintSubmitted(ethTxHash, scheduleID, transactionId string) error {
	args := berm.Called(ethTxHash, scheduleID, transactionId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusScheduledTokenTransferSubmitted(ethTxHash, scheduleID, transactionId string) error {
	args := berm.Called(ethTxHash, scheduleID, transactionId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusScheduledTokenMintCompleted(txId string) error {
	args := berm.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusCompleted(id string) error {
	args := berm.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusFailed(id string) error {
	args := berm.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) Get(id string) (*entity.LockEvent, error) {
	args := berm.Called(id)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.LockEvent), nil
	}
	return nil, args.Get(1).(error)
}
