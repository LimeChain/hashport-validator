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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/mock"
)

type MockBurnService struct {
	mock.Mock
}

func (m *MockBurnService) TransactionID(id string) (string, error) {
	args := m.Called(id)

	if args[1] == nil {
		return args[0].(string), nil
	}
	return args[0].(string), args[1].(error)
}

func (m *MockBurnService) ProcessEvent(event transfer.Transfer) {
	m.Called(event)
}
