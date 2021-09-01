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

package transfer

import (
	"errors"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	mocks "github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	mt = model.Transfer{
		TransactionId: "0.0.0-0000000-1234",
		Receiver:      "0x12345",
		Amount:        "10000000000",
		NativeAsset:   constants.Hbar,
		TargetAsset:   "0x45678",
	}
	mnt = &model.NativeTransfer{Transfer: mt}
	mwt = &model.WrappedTransfer{Transfer: mt}
)

func InitializeHandler() (*Handler, *service.MockTransferService) {
	mocks.Setup()

	return NewHandler(mocks.MTransferService), mocks.MTransferService
}

func Test_Handle(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mt.TransactionId,
		Receiver:      mt.Receiver,
		Amount:        mt.Amount,
		NativeAsset:   mt.NativeAsset,
		TargetAsset:   mt.TargetAsset,
		Status:        transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)
	mockedService.On("ProcessNativeTransfer", mt).Return(nil)

	ctHandler.Handle(mnt)

	mockedService.AssertCalled(t, "InitiateNewTransfer", mt)
	mockedService.AssertCalled(t, "ProcessNativeTransfer", mt)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	invalidTransferPayload := []byte{1, 2, 1}

	ctHandler.Handle(invalidTransferPayload)

	mockedService.AssertNotCalled(t, "InitiateNewTransfer")
	mockedService.AssertNotCalled(t, "ProcessNativeTransfer")
}

func Test_Handle_InitiateNewTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	mockedService.On("InitiateNewTransfer", mt).Return(nil, errors.New("some-error"))

	ctHandler.Handle(mnt)

	mockedService.AssertNotCalled(t, "ProcessNativeTransfer")
}

func Test_Handle_StatusNotInitial_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mnt.TransactionId,
		Receiver:      mnt.Receiver,
		Amount:        mnt.Amount,
		Status:        transfer.StatusCompleted,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)

	ctHandler.Handle(mnt)

	mockedService.AssertNotCalled(t, "ProcessNativeTransfer")
}

func Test_Handle_ProcessNativeTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mnt.TransactionId,
		Receiver:      mnt.Receiver,
		Amount:        mnt.Amount,
		Status:        transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)
	mockedService.On("ProcessNativeTransfer", mt).Return(errors.New("some-error"))
	mockedService.AssertNotCalled(t, "ProcessWrappedTransfer", mock.Anything)

	ctHandler.Handle(mnt)
}

func Test_Handle_ProcessWrappedTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mwt.TransactionId,
		Receiver:      mwt.Receiver,
		Amount:        mwt.Amount,
		Status:        transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)
	mockedService.On("ProcessWrappedTransfer", mt).Return(errors.New("some-error"))
	mockedService.AssertNotCalled(t, "ProcessNativeTransfer", mock.Anything)

	ctHandler.Handle(mwt)
}
