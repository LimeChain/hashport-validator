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
	mocks "github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"testing"
)

var (
	addresses = []string{
		"0xsomeaddress",
		"0xsomeaddress2",
		"0xsomeaddress3",
	}
	// Value of the serviceFeePercent in percentage. Range 0% to 99.999% multiplied my 1000
	serviceFeePercent uint64 = 10000
	tx                       = &model.Transfer{
		TransactionId:         "0.0.0-0000000-1234",
		Receiver:              "0x12345",
		Amount:                "10000000000",
		TxReimbursement:       "500000000",
		GasPrice:              "100000000",
		NativeToken:           "HBAR",
		WrappedToken:          "0x45678",
		ExecuteEthTransaction: true,
	}
)

func InitializeHandler() (*Handler, *service.MockTransferService) {
	mocks.Setup()

	return NewHandler(mocks.MTransferService), mocks.MTransferService
}

func Test_Handle(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:         tx.TransactionId,
		Receiver:              tx.Receiver,
		Amount:                tx.Amount,
		NativeToken:           tx.NativeToken,
		WrappedToken:          tx.WrappedToken,
		TxReimbursement:       tx.TxReimbursement,
		GasPrice:              tx.GasPrice,
		ExecuteEthTransaction: tx.ExecuteEthTransaction,
		Status:                transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", tx).Return(tx, nil)
	mockedService.On("VerifyFee", tx).Return(nil)
	mockedService.On("ProcessTransfer", tx).Return(nil)

	ctHandler.Handle(&tx)

	mockedService.AssertCalled(t, "InitiateNewTransfer", tx)
	mockedService.AssertCalled(t, "VerifyFee", tx)
	mockedService.AssertCalled(t, "ProcessTransfer", tx)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	invalidTransferPayload := []byte{1, 2, 1}

	ctHandler.Handle(invalidTransferPayload)

	mockedService.AssertNotCalled(t, "InitiateNewTransfer")
	mockedService.AssertNotCalled(t, "VerifyFee")
	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_InitiateNewTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	mockedService.On("InitiateNewTransfer", tx).Return(nil, errors.New("some-error"))

	ctHandler.Handle(&tx)

	mockedService.AssertNotCalled(t, "VerifyFee")
	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_StatusNotInitial_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:   tx.TransactionId,
		Receiver:        tx.Receiver,
		Amount:          tx.Amount,
		TxReimbursement: tx.TxReimbursement,
		Status:          transfer.StatusCompleted,
	}

	mockedService.On("InitiateNewTransfer", tx).Return(tx, nil)

	ctHandler.Handle(&tx)

	mockedService.AssertNotCalled(t, "VerifyFee")
	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_VerifyFee_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:   tx.TransactionId,
		Receiver:        tx.Receiver,
		Amount:          tx.Amount,
		TxReimbursement: tx.TxReimbursement,
		Status:          transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", tx).Return(tx, nil)
	mockedService.On("VerifyFee", tx).Return(errors.New("some-error"))

	ctHandler.Handle(&tx)

	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_ProcessTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:   tx.TransactionId,
		Receiver:        tx.Receiver,
		Amount:          tx.Amount,
		TxReimbursement: tx.TxReimbursement,
		Status:          transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", tx).Return(tx, nil)
	mockedService.On("VerifyFee", tx).Return(nil)
	mockedService.On("ProcessTransfer", tx).Return(errors.New("some-error"))

	ctHandler.Handle(&tx)
}
