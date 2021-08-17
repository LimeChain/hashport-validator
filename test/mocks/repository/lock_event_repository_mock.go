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

type MockLockEventRepository struct {
	mock.Mock
}

func (berm *MockLockEventRepository) Create(id string, amount int64, recipient, nativeAsset, wrappedAsset string, sourceChainId, targetChainId int64) error {
	args := berm.Called(id, amount, recipient, nativeAsset, wrappedAsset, sourceChainId, targetChainId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusScheduledTokenMintSubmitted(ethTxHash, scheduleID, transactionId string) error {
	args := berm.Called(ethTxHash, scheduleID, transactionId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusScheduledTokenTransferSubmitted(ethTxHash, scheduleID, transactionId string) error {
	args := berm.Called(ethTxHash, scheduleID, transactionId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusScheduledTokenMintCompleted(txId string) error {
	args := berm.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusCompleted(id string) error {
	args := berm.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) UpdateStatusFailed(id string) error {
	args := berm.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (berm *MockLockEventRepository) Get(id string) (*entity.LockEvent, error) {
	args := berm.Called(id)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.LockEvent), nil
	}
	return nil, args.Get(1).(error)
}
