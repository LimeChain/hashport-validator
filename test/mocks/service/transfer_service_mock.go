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
	"fmt"

	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/stretchr/testify/mock"
)

type MockTransferService struct {
	mock.Mock
}

func (mts *MockTransferService) GetByTransactionId(txId string) (*entity.Transfer, error) {
	panic("implement me")
}

func (mts *MockTransferService) GetWithFee(txId string) (*entity.Transfer, error) {
	panic("implement me")
}

func (mts *MockTransferService) GetWithPreloads(txId string) (*entity.Transfer, error) {
	panic("implement me")
}

func (mts *MockTransferService) Create(ct *payload.Transfer) (*entity.Transfer, error) {
	panic("implement me")
}

func (mts *MockTransferService) UpdateStatusCompleted(txId string) error {
	panic("implement me")
}

func (mts *MockTransferService) UpdateStatusFailed(txId string) error {
	panic("implement me")
}

func (mts *MockTransferService) ProcessNativeTransfer(tm payload.Transfer) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) ProcessNativeNftTransfer(tm payload.Transfer) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) ProcessWrappedTransfer(tm payload.Transfer) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) SanityCheckTransfer(tx transaction.Transaction) transfer.SanityCheckResult {
	args := mts.Called(tx)

	return args.Get(0).(transfer.SanityCheckResult)
}

func (mts *MockTransferService) InitiateNewTransfer(tm payload.Transfer) (*entity.Transfer, error) {
	args := mts.Called(tm)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return args.Get(0).(*entity.Transfer), args.Get(1).(error)
}

func (mts *MockTransferService) TransferData(txId string) (interface{}, error) {
	args := mts.Called(txId)
	if args.Get(0) == nil {
		return service.TransferData{}, args.Get(1).(error)
	}

	return args.Get(0).(service.TransferData), args.Error(1)
}

func (mts *MockTransferService) Paged(filter *transfer.PagedRequest) (*transfer.Paged, error) {
	panic("implement me")
}

func (mts *MockTransferService) UpdateTransferStatusCompleted(txId string) error {
	args := mts.Called(txId)
	if args.Get(0) == nil {
		return nil
	}

	return fmt.Errorf("error")
}
