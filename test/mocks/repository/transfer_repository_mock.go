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
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/stretchr/testify/mock"
)

type MockTransferRepository struct {
	mock.Mock
}

func (m *MockTransferRepository) UpdateStatusFailed(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) GetByTransactionId(txId string) (*entity.Transfer, error) {
	args := m.Called(txId)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return nil, args.Get(1).(error)
}

func (m *MockTransferRepository) GetWithFee(txId string) (*entity.Transfer, error) {
	args := m.Called(txId)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return nil, args.Get(1).(error)
}

func (m *MockTransferRepository) GetWithPreloads(txId string) (*entity.Transfer, error) {
	args := m.Called(txId)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return nil, args.Get(1).(error)
}

func (m *MockTransferRepository) Create(ct *payload.Transfer) (*entity.Transfer, error) {
	args := m.Called(ct)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return nil, args.Get(1).(error)
}

func (m *MockTransferRepository) UpdateFee(txId, fee string) error {
	args := m.Called(txId, fee)
	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusCompleted(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) Paged(req *transfer.PagedRequest) ([]*entity.Transfer, int64, error) {
	panic("implement me")
}
