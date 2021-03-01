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
	"math/big"

	"github.com/stretchr/testify/mock"
)

type MockBridgeContract struct {
	mock.Mock
}

func (m *MockBridgeContract) GetServiceFee() *big.Int {
	args := m.Called()
	return new(big.Int).SetUint64(args.Get(0).(uint64))
}

func (m *MockBridgeContract) GetMembers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}
