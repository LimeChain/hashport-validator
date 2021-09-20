package evm_client

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEVMCoreClient struct {
	mock.Mock
}

func (m *MockEVMCoreClient) ChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].(*big.Int), nil
	}
	return args[0].(*big.Int), args[1].(error)
}

func (m *MockEVMCoreClient) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	args := m.Called(ctx, hash)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].(*types.Block), nil
	}
	return args[0].(*types.Block), args[1].(error)
}

func (m *MockEVMCoreClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].(*types.Block), nil
	}
	return args[0].(*types.Block), args[1].(error)
}

func (m *MockEVMCoreClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	if args[0] == nil && args[1] == nil {
		return uint64(0), nil
	}
	if args[0] == nil {
		return uint64(0), args[1].(error)
	}
	if args[1] == nil {
		return args[0].(uint64), nil
	}
	return args[0].(uint64), args[1].(error)
}

func (m *MockEVMCoreClient) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	args := m.Called(ctx, hash)
	if args[0] == nil && args[2] == nil {
		return nil, args[1].(bool), nil
	}
	if args[0] == nil {
		return nil, args[1].(bool), args[2].(error)
	}
	if args[2] == nil {
		return args[0].(*types.Transaction), args[1].(bool), nil
	}
	return args[0].(*types.Transaction), args[1].(bool), args[2].(error)
}

func (m *MockEVMCoreClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].(*types.Receipt), nil
	}
	return args[0].(*types.Receipt), args[1].(error)
}

func (m *MockEVMCoreClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].([]byte), nil
	}
	return args[0].([]byte), args[1].(error)
}

func (m *MockEVMCoreClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, q)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].([]types.Log), nil
	}
	return args[0].([]types.Log), args[1].(error)
}

func (m *MockEVMCoreClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].([]byte), nil
	}
	return args[0].([]byte), args[1].(error)
}
func (m *MockEVMCoreClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	panic("implement me")
}
func (m *MockEVMCoreClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	args := m.Called(ctx, account)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].([]byte), nil
	}
	return args[0].([]byte), args[1].(error)
}
func (m *MockEVMCoreClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	panic("implement me")
}
func (m *MockEVMCoreClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}
func (m *MockEVMCoreClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}
func (m *MockEVMCoreClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	panic("implement me")
}
func (m *MockEVMCoreClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	panic("implement me")
}
func (m *MockEVMCoreClient) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	panic("implement me")
}
