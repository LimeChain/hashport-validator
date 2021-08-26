package repository

import "github.com/stretchr/testify/mock"

type MockStatusRepository struct {
	mock.Mock
}

func (msr *MockStatusRepository) GetLastFetchedTimestamp(entityID string) (int64, error) {
	args := msr.Called(entityID)
	if args[1] == nil {
		return args[0].(int64), nil
	}
	return args[0].(int64), args[1].(error)
}

func (msr *MockStatusRepository) UpdateLastFetchedTimestamp(entityID string, timestamp int64) error {
	args := msr.Called(entityID, timestamp)
	if args[0] == nil {
		return nil
	}
	return args[0].(error)
}

func (msr *MockStatusRepository) CreateTimestamp(entityID string, timestamp int64) error {
	args := msr.Called(entityID, timestamp)
	if args[0] == nil {
		return nil
	}
	return args[0].(error)
}
