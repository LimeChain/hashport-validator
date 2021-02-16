package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockExchangeRateProvider struct {
	mock.Mock
}

func (m MockExchangeRateProvider) GetEthVsHbarRate() (float64, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(float64), nil
	}
	return args.Get(0).(float64), args.Get(1).(error)
}
