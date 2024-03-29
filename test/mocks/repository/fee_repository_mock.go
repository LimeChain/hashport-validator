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

type MockFeeRepository struct {
	mock.Mock
}

func (mfr *MockFeeRepository) GetAllSubmittedIds() ([]*entity.Fee, error) {
	args := mfr.Called()
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Fee), nil
	}
	return args.Get(0).([]*entity.Fee), args.Get(1).(error)
}

func (mfr *MockFeeRepository) Create(entity *entity.Fee) error {
	args := mfr.Called(entity)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mfr *MockFeeRepository) UpdateStatusCompleted(id string) error {
	args := mfr.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mfr *MockFeeRepository) UpdateStatusFailed(id string) error {
	args := mfr.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mfr *MockFeeRepository) Get(id string) (*entity.Fee, error) {
	args := mfr.Called(id)
	if args.Get(0) == nil {
		return args.Get(0).(*entity.Fee), nil
	}
	return nil, args.Get(1).(error)
}
