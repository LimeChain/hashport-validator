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

package evm_client

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEVMClient struct {
	mock.Mock
}

func (m *MockEVMClient) SetChainID(chainId uint64) {
	m.Called(chainId)
}

func (m *MockEVMClient) GetChainID() uint64 {
	args := m.Called()
	return args.Get(0).(uint64)
}

func (m *MockEVMClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)

	if args.Get(1) == nil {
		return args.Get(0).(uint64), nil
	}

	return args.Get(0).(uint64), args.Get(1).(error)
}

func (m *MockEVMClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, q)

	if args.Get(1) == nil {
		return args.Get(0).([]types.Log), nil
	}

	return args.Get(0).([]types.Log), args.Get(1).(error)
}

func (m *MockEVMClient) RetryBlockNumber() (uint64, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(uint64), nil
	}

	return args.Get(0).(uint64), args.Get(1).(error)
}

func (m *MockEVMClient) RetryFilterLogs(q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(q)

	if args.Get(1) == nil {
		return args.Get(0).([]types.Log), nil
	}

	return args.Get(0).([]types.Log), args.Get(1).(error)
}

func (m *MockEVMClient) ChainID(ctx context.Context) (*big.Int, error) {
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

func (m *MockEVMClient) GetClient() client.Core {
	args := m.Called()
	return args.Get(0).(client.Core)
}

func (m *MockEVMClient) GetBlockTimestamp(blockNumber *big.Int) uint64 {
	args := m.Called(blockNumber)

	return args.Get(0).(uint64)
}

func (m *MockEVMClient) ValidateContractDeployedAt(contractAddress string) (*common.Address, error) {
	args := m.Called(contractAddress)

	if args.Get(1) == nil {
		return args.Get(0).(*common.Address), nil
	}
	return args.Get(0).(*common.Address), args.Get(1).(error)
}

func (m *MockEVMClient) WaitForTransaction(hex string, onSuccess, onRevert func(), onError func(err error)) {
	m.Called(hex, onSuccess, onRevert, onError)
}

func (m *MockEVMClient) WaitForConfirmations(raw types.Log) error {
	args := m.Called(raw)

	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockEVMClient) GetPrivateKey() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockEVMClient) BlockConfirmations() uint64 {
	args := m.Called()

	return args.Get(0).(uint64)
}
