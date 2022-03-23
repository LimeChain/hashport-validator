/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEVMCore struct {
	mock.Mock
}

func (m *MockEVMCore) ChainID(ctx context.Context) (*big.Int, error) {
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

func (m *MockEVMCore) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
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

func (m *MockEVMCore) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
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

func (m *MockEVMCore) BlockNumber(ctx context.Context) (uint64, error) {
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

func (m *MockEVMCore) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
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

func (m *MockEVMCore) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
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

func (m *MockEVMCore) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
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

func (m *MockEVMCore) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
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

func (m *MockEVMCore) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
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
func (m *MockEVMCore) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	panic("implement me")
}
func (m *MockEVMCore) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
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
func (m *MockEVMCore) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	panic("implement me")
}
func (m *MockEVMCore) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}
func (m *MockEVMCore) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}
func (m *MockEVMCore) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	panic("implement me")
}
func (m *MockEVMCore) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	panic("implement me")
}
func (m *MockEVMCore) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	panic("implement me")
}
