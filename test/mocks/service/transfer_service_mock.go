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

package service

import (
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockTransferService struct {
	mock.Mock
}

func (mts *MockTransferService) ProcessNativeTransfer(tm transfer.Transfer) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) ProcessWrappedTransfer(tm transfer.Transfer) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) SanityCheckTransfer(tx mirror_node.Transaction) (int64, string, error) {
	args := mts.Called(tx)
	if args.Get(2) == nil {
		return args.Get(0).(int64), args.Get(1).(string), nil
	}
	return args.Get(0).(int64), args.Get(1).(string), args.Get(2).(error)
}

func (mts *MockTransferService) SaveRecoveredTxn(txId, amount, nativeAsset, wrappedAsset, evmAddress string) error {
	args := mts.Called(txId, amount, nativeAsset, wrappedAsset, evmAddress)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) InitiateNewTransfer(tm transfer.Transfer) (*entity.Transfer, error) {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return args.Get(0).(*entity.Transfer), args.Get(1).(error)
}

func (mts *MockTransferService) TransferData(txId string) (service.TransferData, error) {
	args := mts.Called(txId)
	if args.Get(0) == nil {
		return service.TransferData{}, args.Get(1).(error)
	}

	return args.Get(0).(service.TransferData), args.Get(0).(error)
}
