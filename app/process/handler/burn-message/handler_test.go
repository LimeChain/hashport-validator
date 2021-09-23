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

package burn_message

import (
	"errors"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
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
)

func InitializeHandler() (*Handler, *service.MockTransferService) {
	mocks.Setup()

	return NewHandler(mocks.MTransferService), mocks.MTransferService
}

func Test_Handle_ProcessWrappedTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mt.TransactionId,
		Receiver:      mt.Receiver,
		Amount:        mt.Amount,
		Status:        status.Initial,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)
	mockedService.On("ProcessWrappedTransfer", mt).Return(errors.New("some-error"))

	ctHandler.Handle(&mt)
}

func Test_Handle_NotInitial(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mt.TransactionId,
		Receiver:      mt.Receiver,
		Amount:        mt.Amount,
		Status:        status.Completed,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)
	ctHandler.Handle(&mt)
	mockedService.AssertNotCalled(t, "ProcessWrappedTransfer", mock.Anything)
}

func Test_Handle_InitiateNewTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()
	mockedService.On("InitiateNewTransfer", mt).Return(nil, errors.New("some-error"))
	ctHandler.Handle(&mt)
	mockedService.AssertNotCalled(t, "ProcessWrappedTransfer", mock.Anything)
}

func Test_Handle_Payload_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()
	ctHandler.Handle("string")
	mockedService.AssertNotCalled(t, "InitiateNewTransfer", mock.Anything)
	mockedService.AssertNotCalled(t, "ProcessWrappedTransfer", mock.Anything)
}
