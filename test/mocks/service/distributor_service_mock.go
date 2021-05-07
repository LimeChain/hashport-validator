package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/mock"
)

type MockDistrubutorService struct {
	mock.Mock
}

func (mds *MockDistrubutorService) CalculateMemberDistribution(validFee int64) ([]transfer.Hedera, error) {
	args := mds.Called(validFee)
	if args.Get(1) == nil {
		return args.Get(0).([]transfer.Hedera), nil
	}
	return nil, args.Get(1).(error)
}

func (mds *MockDistrubutorService) ValidAmount(amount int64) int64 {
	args := mds.Called(amount)
	return args.Get(0).(int64)
}
