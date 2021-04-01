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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	mocks "github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"testing"
)

const (
	topicID   = "0.0.125563"
	accountID = "0.0.99661"
)

var (
	addresses = []string{
		"0xsomeaddress",
		"0xsomeaddress2",
		"0xsomeaddress3",
	}
	// Value of the serviceFeePercent in percentage. Range 0% to 99.999% multiplied my 1000
	serviceFeePercent    uint64 = 10000
	protoTransferMessage        = &protomsg.TransferMessage{
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

func GetTestData() (encoding.TransferMessage, hedera.TopicID, hedera.AccountID) {
	ctm := encoding.TransferMessage{TransferMessage: protoTransferMessage}
	topicID, _ := hedera.TopicIDFromString(topicID)
	accID, _ := hedera.AccountIDFromString(accountID)

	return ctm, topicID, accID
}

func Test_Handle(t *testing.T) {
	ctm, _, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:         ctm.TransactionId,
		Receiver:              ctm.Receiver,
		Amount:                ctm.Amount,
		NativeToken:           ctm.NativeToken,
		WrappedToken:          ctm.WrappedToken,
		TxReimbursement:       ctm.TxReimbursement,
		GasPrice:              ctm.GasPrice,
		ExecuteEthTransaction: ctm.ExecuteEthTransaction,
		Status:                transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)
	mockedService.On("VerifyFee", ctm).Return(nil)
	mockedService.On("ProcessTransfer", ctm).Return(nil)

	ctHandler.Handle(&ctm)

	mockedService.AssertCalled(t, "InitiateNewTransfer", ctm)
	mockedService.AssertCalled(t, "VerifyFee", ctm)
	mockedService.AssertCalled(t, "ProcessTransfer", ctm)
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
	ctm, _, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	mockedService.On("InitiateNewTransfer", ctm).Return(nil, errors.New("some-error"))

	ctHandler.Handle(&ctm)

	mockedService.AssertNotCalled(t, "VerifyFee")
	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_StatusNotInitial_Fails(t *testing.T) {
	ctm, _, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:   ctm.TransactionId,
		Receiver:        ctm.Receiver,
		Amount:          ctm.Amount,
		TxReimbursement: ctm.TxReimbursement,
		Status:          transfer.StatusCompleted,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)

	ctHandler.Handle(&ctm)

	mockedService.AssertNotCalled(t, "VerifyFee")
	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_VerifyFee_Fails(t *testing.T) {
	ctm, _, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:   ctm.TransactionId,
		Receiver:        ctm.Receiver,
		Amount:          ctm.Amount,
		TxReimbursement: ctm.TxReimbursement,
		Status:          transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)
	mockedService.On("VerifyFee", ctm).Return(errors.New("some-error"))

	ctHandler.Handle(&ctm)

	mockedService.AssertNotCalled(t, "ProcessTransfer")
}

func Test_Handle_ProcessTransfer_Fails(t *testing.T) {
	ctm, _, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID:   ctm.TransactionId,
		Receiver:        ctm.Receiver,
		Amount:          ctm.Amount,
		TxReimbursement: ctm.TxReimbursement,
		Status:          transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)
	mockedService.On("VerifyFee", ctm).Return(nil)
	mockedService.On("ProcessTransfer", ctm).Return(errors.New("some-error"))

	ctHandler.Handle(&ctm)
}
