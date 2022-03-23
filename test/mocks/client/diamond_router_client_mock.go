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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockDiamondRouter struct {
	mock.Mock
}

func (m *MockDiamondRouter) WatchLock(opts *bind.WatchOpts, sink chan<- *router.RouterLock) (event.Subscription, error) {
	args := m.Called(opts, sink)
	return args.Get(0).(event.Subscription), args.Error(1)
}

func (m *MockDiamondRouter) HasValidSignaturesLength(opts *bind.CallOpts, _n *big.Int) (bool, error) {
	args := m.Called(opts, _n)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockDiamondRouter) ParseMint(log types.Log) (*router.RouterMint, error) {
	args := m.Called(log)
	return args.Get(0).(*router.RouterMint), args.Error(1)
}

func (m *MockDiamondRouter) ParseBurn(log types.Log) (*router.RouterBurn, error) {
	args := m.Called(log)
	return args.Get(0).(*router.RouterBurn), args.Error(1)
}

func (m *MockDiamondRouter) ParseLock(log types.Log) (*router.RouterLock, error) {
	args := m.Called(log)
	return args.Get(0).(*router.RouterLock), args.Error(1)
}

func (m *MockDiamondRouter) ParseUnlock(log types.Log) (*router.RouterUnlock, error) {
	args := m.Called(log)
	return args.Get(0).(*router.RouterUnlock), args.Error(1)
}

func (m *MockDiamondRouter) ParseBurnERC721(log types.Log) (*router.RouterBurnERC721, error) {
	args := m.Called(log)
	return args.Get(0).(*router.RouterBurnERC721), args.Error(1)
}

func (m *MockDiamondRouter) WatchBurn(opts *bind.WatchOpts, sink chan<- *router.RouterBurn) (event.Subscription, error) {
	args := m.Called(opts, sink)
	return args.Get(0).(event.Subscription), args.Error(1)
}

func (m *MockDiamondRouter) MembersCount(opts *bind.CallOpts) (*big.Int, error) {
	args := m.Called(opts)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockDiamondRouter) MemberAt(opts *bind.CallOpts, _index *big.Int) (common.Address, error) {
	args := m.Called(opts, _index)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *MockDiamondRouter) TokenFeeData(opts *bind.CallOpts, _token common.Address) (struct {
	ServiceFeePercentage *big.Int
	FeesAccrued          *big.Int
	PreviousAccrued      *big.Int
	Accumulator          *big.Int
}, error) {
	args := m.Called(opts, _token)

	return args.Get(0).(struct {
		ServiceFeePercentage *big.Int
		FeesAccrued          *big.Int
		PreviousAccrued      *big.Int
		Accumulator          *big.Int
	}), args.Error(1)
}
