package service

import "github.com/stretchr/testify/mock"

type MockFeeService struct {
	mock.Mock
}

func (mfs *MockFeeService) CalculateFee(amount int64) (fee, remainder int64) {
	args := mfs.Called(amount)
	return args.Get(0).(int64), args.Get(1).(int64)
}
