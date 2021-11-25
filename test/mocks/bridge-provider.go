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

package mocks

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockBridgeContract struct {
	mock.Mock
}

func (m *MockBridgeContract) AddDecimals(amount *big.Int, asset string) (*big.Int, error) {
	args := m.Called(amount, asset)

	if args[1] == nil {
		return args[0].(*big.Int), nil
	}
	return args[0].(*big.Int), args[1].(error)
}

func (m *MockBridgeContract) RemoveDecimals(amount *big.Int, asset string) (*big.Int, error) {
	args := m.Called(amount, asset)

	if args[1] == nil {
		return args[0].(*big.Int), nil
	}
	return args[0].(*big.Int), args[1].(error)
}

func (m *MockBridgeContract) GetClient() client.Core {
	panic("implement me")
}

func (m *MockBridgeContract) ParseBurnLog(log types.Log) (*router.RouterBurn, error) {
	args := m.Called(log)
	if args[0] == nil {
		return nil, args.Get(1).(error)
	}
	if args[1] == nil {
		return args.Get(0).(*router.RouterBurn), nil
	}
	return args.Get(0).(*router.RouterBurn), args.Get(1).(error)
}

func (m *MockBridgeContract) ParseLockLog(log types.Log) (*router.RouterLock, error) {
	args := m.Called(log)
	if args[0] == nil {
		return nil, args.Get(1).(error)
	}
	if args[1] == nil {
		return args.Get(0).(*router.RouterLock), nil
	}
	return args.Get(0).(*router.RouterLock), args.Get(1).(error)
}

func (m *MockBridgeContract) ParseBurnERC721Log(log types.Log) (*router.RouterBurnERC721, error) {
	args := m.Called(log)
	if args[0] == nil {
		return nil, args.Get(1).(error)
	}
	if args[1] == nil {
		return args.Get(0).(*router.RouterBurnERC721), nil
	}
	return args.Get(0).(*router.RouterBurnERC721), args.Get(1).(error)
}

func (m *MockBridgeContract) IsMember(address string) bool {
	panic("implement me")
}

func (m *MockBridgeContract) HasValidSignaturesLength(signaturesLength *big.Int) (bool, error) {
	args := m.Called(signaturesLength)
	if args[0] == nil {
		return false, args.Get(1).(error)
	}
	return args.Get(0).(bool), nil
}

func (m *MockBridgeContract) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterBurn) (event.Subscription, error) {
	args := m.Called(opts, sink)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args.Get(1).(error)
	}
	if args[1] == nil {
		return args.Get(0).(event.Subscription), nil
	}
	return args.Get(0).(event.Subscription), args.Get(1).(error)
}

func (m *MockBridgeContract) WatchLockEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterLock) (event.Subscription, error) {
	args := m.Called(opts, sink)
	if args[0] == nil {
		return nil, args.Get(1).(error)
	}
	if args[1] == nil {
		return args.Get(0).(event.Subscription), nil
	}
	return args.Get(0).(event.Subscription), args.Get(1).(error)
}

func (m *MockBridgeContract) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000000")
}

func (m *MockBridgeContract) GetMembers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockBridgeContract) ReloadMembers() {
	m.Called()
}
