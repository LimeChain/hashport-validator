package service

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockSignerService struct {
	mock.Mock
}

func (m *MockSignerService) Sign(msg []byte) ([]byte, error) {
	args := m.Called(msg)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil
	}
	return args.Get(0).([]byte), args.Get(1).(error)
}

func (m *MockSignerService) NewKeyTransactor(chainId *big.Int) (*bind.TransactOpts, error) {
	args := m.Called(chainId)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*bind.TransactOpts), nil
	}
	return args.Get(0).(*bind.TransactOpts), args.Get(1).(error)
}

func (m *MockSignerService) Address() string {
	args := m.Called()
	return args.Get(0).(string)
}
