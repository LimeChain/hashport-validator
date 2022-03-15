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

package evm_token_client

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEVMTokenClient struct {
	mock.Mock
}

func (m *MockEVMTokenClient) Decimals(opts *bind.CallOpts) (uint8, error) {
	args := m.Called(opts)
	return args.Get(0).(uint8), args.Error(1)
}

func (m *MockEVMTokenClient) Name(opts *bind.CallOpts) (string, error) {
	args := m.Called(opts)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockEVMTokenClient) Symbol(opts *bind.CallOpts) (string, error) {
	args := m.Called(opts)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockEVMTokenClient) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEVMTokenClient) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	args := m.Called(opts, account)
	return args.Get(0).(*big.Int), args.Error(1)
}
