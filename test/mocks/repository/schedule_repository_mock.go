package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) Get(txId string) (*entity.Schedule, error) {
	panic("implement me")
}

func (m *MockScheduleRepository) Create(entity *entity.Schedule) error {
	args := m.Called(entity)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockScheduleRepository) UpdateStatusCompleted(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockScheduleRepository) UpdateStatusFailed(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockScheduleRepository) GetTransferByTransactionID(id string) (*entity.Schedule, error) {
	args := m.Called(id)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Schedule), nil
	}
	return nil, args.Get(1).(error)
}
