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
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/stretchr/testify/mock"
)

type MockBridgeContract struct {
	mock.Mock
}

func (m *MockBridgeContract) IsMember(address string) bool {
	panic("implement me")
}

func (m *MockBridgeContract) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterBurn) (event.Subscription, error) {
	panic("implement me")
}

func (m *MockBridgeContract) WatchLockEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterLock) (event.Subscription, error) {
	panic("implement me")
}

func (m *MockBridgeContract) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000000")
}

func (m *MockBridgeContract) GetMembers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}
