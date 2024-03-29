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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEvmNonFungibleToken struct {
	mock.Mock
}

func (m *MockEvmNonFungibleToken) Name(opts *bind.CallOpts) (string, error) {
	args := m.Called(opts)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockEvmNonFungibleToken) Symbol(opts *bind.CallOpts) (string, error) {
	args := m.Called(opts)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockEvmNonFungibleToken) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	args := m.Called(opts, owner)
	return args.Get(0).(*big.Int), args.Error(1)
}
