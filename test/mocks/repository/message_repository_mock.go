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

package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(message *entity.Message) error {
	args := m.Called(message)
	if args[0] == nil {
		return nil
	}
	return args[0].(error)
}

func (m *MockMessageRepository) Exist(transferID, signature, hash string) (bool, error) {
	args := m.Called(transferID, signature, hash)
	if args[0] == nil {
		return args[0].(bool), nil
	}
	return args[0].(bool), args[0].(error)
}

func (m *MockMessageRepository) Get(transferID string) ([]entity.Message, error) {
	args := m.Called(transferID)
	if args[1] == nil {
		return args[0].([]entity.Message), nil
	}
	return args[0].([]entity.Message), args[1].(error)
}

func (m *MockMessageRepository) GetMessageWith(transferID, signature, hash string) (*entity.Message, error) {
	args := m.Called(transferID, signature, hash)
	if args[0] == nil {
		return args[0].(*entity.Message), nil
	}
	return args[0].(*entity.Message), args[0].(error)
}
