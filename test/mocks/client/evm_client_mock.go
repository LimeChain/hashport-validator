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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEVM struct {
	mock.Mock
}

func (m *MockEVM) SetChainID(chainId uint64) {
	m.Called(chainId)
}

func (m *MockEVM) GetChainID() uint64 {
	args := m.Called()
	return args.Get(0).(uint64)
}

func (m *MockEVM) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)

	if args.Get(1) == nil {
		return args.Get(0).(uint64), nil
	}

	return args.Get(0).(uint64), args.Get(1).(error)
}

func (m *MockEVM) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, q)

	if args.Get(1) == nil {
		return args.Get(0).([]types.Log), nil
	}

	return args.Get(0).([]types.Log), args.Get(1).(error)
}

func (m *MockEVM) RetryBlockNumber() (uint64, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(uint64), nil
	}

	return args.Get(0).(uint64), args.Get(1).(error)
}

func (m *MockEVM) RetryFilterLogs(q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(q)

	if args.Get(1) == nil {
		return args.Get(0).([]types.Log), nil
	}

	return args.Get(0).([]types.Log), args.Get(1).(error)
}

func (m *MockEVM) ChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*big.Int), nil
	}
	return args.Get(0).(*big.Int), args.Get(1).(error)
}

func (m *MockEVM) GetClient() client.Core {
	args := m.Called()
	return args.Get(0).(client.Core)
}

func (m *MockEVM) GetBlockTimestamp(blockNumber *big.Int) uint64 {
	args := m.Called(blockNumber)

	return args.Get(0).(uint64)
}

func (m *MockEVM) ValidateContractDeployedAt(contractAddress string) (*common.Address, error) {
	args := m.Called(contractAddress)

	if args.Get(1) == nil {
		return args.Get(0).(*common.Address), nil
	}
	return args.Get(0).(*common.Address), args.Get(1).(error)
}

func (m *MockEVM) WaitForTransaction(hex string, onSuccess, onRevert func(), onError func(err error)) {
	m.Called(hex, onSuccess, onRevert, onError)
}

func (m *MockEVM) WaitForConfirmations(raw types.Log) error {
	args := m.Called(raw)

	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockEVM) GetPrivateKey() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockEVM) BlockConfirmations() uint64 {
	args := m.Called()

	return args.Get(0).(uint64)
}

func (m *MockEVM) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, contract, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVM) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVM) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEVM) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	args := m.Called(ctx, account)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVM) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEVM) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEVM) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEVM) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	args := m.Called(ctx, call)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEVM) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEVM) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	args := m.Called(ctx, query, ch)
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

func (m *MockEVM) WaitForTransactionReceipt(hash common.Hash) (txReceipt *types.Receipt, err error) {
	args := m.Called(hash)
	if err, ok := args.Get(1).(error); ok {
		return nil, err
	}
	return args.Get(0).(*types.Receipt), nil
}
