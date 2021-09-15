/*
 * Copyright 2021 LimeChain Ltd.
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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEVMClient struct {
	mock.Mock
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

func (m *MockEVMClient) ChainID() *big.Int {
	args := m.Called()
	return args.Get(0).(*big.Int)
}

func (m *MockEVMClient) GetClient() *ethclient.Client {
	args := m.Called()
	return args.Get(0).(*ethclient.Client)
}

func (m *MockEVMClient) GetBlockTimestamp(blockNumber *big.Int) (uint64, error) {
	args := m.Called(blockNumber)

	if args.Get(1) == nil {
		return args.Get(0).(uint64), nil
	}

	return args.Get(0).(uint64), args.Get(1).(error)
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

func (m *MockEVMClient) GetRouterContractAddress() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockEVMClient) GetPrivateKey() string {
	args := m.Called()
	return args.Get(0).(string)
}
