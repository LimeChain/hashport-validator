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

package message

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(message *entity.Transfer) error {
	args := m.Called(message)
	return args.Get(0).(error)
}

func (m *MockMessageRepository) GetTransaction(txId, signature, hash string) (*entity.Transfer, error) {
	args := m.Called(txId, signature, hash)
	return args.Get(0).(*entity.Transfer), args.Get(1).(error)
}

func (m *MockMessageRepository) GetTransactions(txId string, txHash string) ([]entity.Transfer, error) {
	args := m.Called(txId, txHash)
	return args.Get(0).([]entity.Transfer), args.Get(1).(error)
}
