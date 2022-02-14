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

package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) GetAllSubmittedIds() ([]*entity.Schedule, error) {
	args := m.Called()
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Schedule), nil
	}
	return args.Get(0).([]*entity.Schedule), args.Get(1).(error)
}

func (m *MockScheduleRepository) Get(txId string) (*entity.Schedule, error) {
	panic("implement me")
}

func (m *MockScheduleRepository) Create(entity *entity.Schedule) error {
	args := m.Called(entity)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockScheduleRepository) UpdateStatusCompleted(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockScheduleRepository) UpdateStatusFailed(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockScheduleRepository) GetReceiverTransferByTransactionID(id string) (*entity.Schedule, error) {
	args := m.Called(id)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Schedule), nil
	}
	return nil, args.Get(1).(error)
}
