package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockFeeRepository struct {
	mock.Mock
}

func (mfr *MockFeeRepository) Create(entity *entity.Fee) error {
	args := mfr.Called(entity)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mfr *MockFeeRepository) UpdateStatusCompleted(id string) error {
	args := mfr.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mfr *MockFeeRepository) UpdateStatusFailed(id string) error {
	args := mfr.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mfr *MockFeeRepository) Get(id string) (*entity.Fee, error) {
	args := mfr.Called(id)
	if args.Get(0) == nil {
		return args.Get(0).(*entity.Fee), nil
	}
	return nil, args.Get(1).(error)
}
