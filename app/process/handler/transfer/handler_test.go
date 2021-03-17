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
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	mocks "github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"gorm.io/gorm"
	"testing"
)

const (
	topicID        = "0.0.125563"
	accountID      = "0.0.99661"
	submissionTxID = "0.0.99661--62135596800-0"
	signature      = "f9f9c16aa2ac71b8341d9187c37c2b8dd8152c4a27fe70f8fcf60d56456166ce704c3f1df4831d66e26879a32cb764d928de346418c1f0f116cba14d78a4dfac1b"
)

var (
	addresses = []string{
		"0xsomeaddress",
		"0xsomeaddress2",
		"0xsomeaddress3",
	}
	// Value of the serviceFeePercent in percentage. Range 0% to 99.999% multiplied my 1000
	serviceFeePercent uint64 = 10000
)

func InitializeHandler() (*Handler, *service.MockTransferService) {
	mocks.Setup()

	return NewHandler(mocks.MTransactionService), mocks.MTransactionService
}

func GetTestData() (encoding.TransferMessage, hedera.TopicID, hedera.AccountID, []byte, []byte) {
	ctm := encoding.TransferMessage{TransferMessage: &protomsg.TransferMessage{}}
	topicID, _ := hedera.TopicIDFromString(topicID)
	accID, _ := hedera.AccountIDFromString(accountID)

	cryptoTransferPayload := []byte{10, 30, 48, 46, 48, 46, 57, 57, 54, 54, 49, 45, 49, 54, 49, 51, 54, 54, 50, 55, 54, 52, 45, 51, 55, 52, 53, 48, 50, 48, 54, 51, 18, 42, 48, 120, 55, 99, 70, 97, 101, 50, 100, 101, 70, 49, 53, 100, 70, 56, 54, 67, 102, 100, 65, 57, 102, 50, 100, 50, 53, 65, 51, 54, 49, 102, 49, 49, 50, 51, 70, 52, 50, 101, 68, 68, 26, 10, 49, 48, 48, 48, 48, 48, 48, 48, 48, 48, 34, 9, 54, 48, 48, 48, 48, 48, 48, 48, 48, 42, 1, 49, 48, 1}
	topicSubmissionMessageBytes := []byte{0x12, 0xe8, 0x1, 0xa, 0x1e, 0x30, 0x2e, 0x30, 0x2e, 0x39, 0x39, 0x36, 0x36, 0x31, 0x2d, 0x31, 0x36, 0x31, 0x33, 0x36, 0x36, 0x32, 0x37, 0x36, 0x34, 0x2d, 0x33, 0x37, 0x34, 0x35, 0x30, 0x32, 0x30, 0x36, 0x33, 0x12, 0x2a, 0x30, 0x78, 0x37, 0x63, 0x46, 0x61, 0x65, 0x32, 0x64, 0x65, 0x46, 0x31, 0x35, 0x64, 0x46, 0x38, 0x36, 0x43, 0x66, 0x64, 0x41, 0x39, 0x66, 0x32, 0x64, 0x32, 0x35, 0x41, 0x33, 0x36, 0x31, 0x66, 0x31, 0x31, 0x32, 0x33, 0x46, 0x34, 0x32, 0x65, 0x44, 0x44, 0x1a, 0xa, 0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x22, 0x9, 0x36, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x2a, 0x82, 0x1, 0x66, 0x39, 0x66, 0x39, 0x63, 0x31, 0x36, 0x61, 0x61, 0x32, 0x61, 0x63, 0x37, 0x31, 0x62, 0x38, 0x33, 0x34, 0x31, 0x64, 0x39, 0x31, 0x38, 0x37, 0x63, 0x33, 0x37, 0x63, 0x32, 0x62, 0x38, 0x64, 0x64, 0x38, 0x31, 0x35, 0x32, 0x63, 0x34, 0x61, 0x32, 0x37, 0x66, 0x65, 0x37, 0x30, 0x66, 0x38, 0x66, 0x63, 0x66, 0x36, 0x30, 0x64, 0x35, 0x36, 0x34, 0x35, 0x36, 0x31, 0x36, 0x36, 0x63, 0x65, 0x37, 0x30, 0x34, 0x63, 0x33, 0x66, 0x31, 0x64, 0x66, 0x34, 0x38, 0x33, 0x31, 0x64, 0x36, 0x36, 0x65, 0x32, 0x36, 0x38, 0x37, 0x39, 0x61, 0x33, 0x32, 0x63, 0x62, 0x37, 0x36, 0x34, 0x64, 0x39, 0x32, 0x38, 0x64, 0x65, 0x33, 0x34, 0x36, 0x34, 0x31, 0x38, 0x63, 0x31, 0x66, 0x30, 0x66, 0x31, 0x31, 0x36, 0x63, 0x62, 0x61, 0x31, 0x34, 0x64, 0x37, 0x38, 0x61, 0x34, 0x64, 0x66, 0x61, 0x63, 0x31, 0x62}

	return ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes
}

func Test_Handle(t *testing.T) {
	ctm, _, _, cryptoTransferPayload, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	tx := &txRepo.Transaction{
		Model:            gorm.Model{},
		TransactionId:    ctm.TransactionId,
		EthAddress:       ctm.EthAddress,
		Amount:           ctm.Amount,
		Fee:              ctm.Fee,
		Signature:        signature,
		SignatureMsgTxId: submissionTxID,
		Status:           txRepo.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)
	mockedService.On("VerifyFee", ctm).Return(nil)
	mockedService.On("ProcessTransfer", ctm).Return(nil)

	ctHandler.Handle(cryptoTransferPayload)

	mockedService.AssertCalled(t, "InitiateNewTransfer", ctm)
	mockedService.AssertCalled(t, "VerifyFee", ctm)
	mockedService.AssertCalled(t, "ProcessTransfer", ctm)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	ctm, _, _, _, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	invalidTransferPayload := []byte{1, 2, 1}

	ctHandler.Handle(invalidTransferPayload)

	mockedService.AssertNotCalled(t, "InitiateNewTransfer", ctm)
	mockedService.AssertNotCalled(t, "VerifyFee", ctm)
	mockedService.AssertNotCalled(t, "ProcessTransfer", ctm)
}

func Test_Handle_InitiateNewTransfer_Fails(t *testing.T) {
	ctm, _, _, cryptoTransferPayload, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	mockedService.On("InitiateNewTransfer", ctm).Return(nil, errors.New("some-error"))

	ctHandler.Handle(cryptoTransferPayload)

	mockedService.AssertNotCalled(t, "VerifyFee", ctm)
	mockedService.AssertNotCalled(t, "ProcessTransfer", ctm)
}

func Test_Handle_StatusNotInitial_Fails(t *testing.T) {
	ctm, _, _, cryptoTransferPayload, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	tx := &txRepo.Transaction{
		Model:            gorm.Model{},
		TransactionId:    ctm.TransactionId,
		EthAddress:       ctm.EthAddress,
		Amount:           ctm.Amount,
		Fee:              ctm.Fee,
		Signature:        signature,
		SignatureMsgTxId: submissionTxID,
		Status:           txRepo.StatusCompleted,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)

	ctHandler.Handle(cryptoTransferPayload)

	mockedService.AssertNotCalled(t, "VerifyFee", ctm)
	mockedService.AssertNotCalled(t, "ProcessTransfer", ctm)
}

func Test_Handle_VerifyFee_Fails(t *testing.T) {
	ctm, _, _, cryptoTransferPayload, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	tx := &txRepo.Transaction{
		Model:            gorm.Model{},
		TransactionId:    ctm.TransactionId,
		EthAddress:       ctm.EthAddress,
		Amount:           ctm.Amount,
		Fee:              ctm.Fee,
		Signature:        signature,
		SignatureMsgTxId: submissionTxID,
		Status:           txRepo.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)
	mockedService.On("VerifyFee", ctm).Return(errors.New("some-error"))

	ctHandler.Handle(cryptoTransferPayload)

	mockedService.AssertNotCalled(t, "ProcessTransfer", ctm)
}

func Test_Handle_ProcessTransfer_Fails(t *testing.T) {
	ctm, _, _, cryptoTransferPayload, _ := GetTestData()
	ctHandler, mockedService := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	tx := &txRepo.Transaction{
		Model:            gorm.Model{},
		TransactionId:    ctm.TransactionId,
		EthAddress:       ctm.EthAddress,
		Amount:           ctm.Amount,
		Fee:              ctm.Fee,
		Signature:        signature,
		SignatureMsgTxId: submissionTxID,
		Status:           txRepo.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", ctm).Return(tx, nil)
	mockedService.On("VerifyFee", ctm).Return(nil)
	mockedService.On("ProcessTransfer", ctm).Return(errors.New("some-error"))

	ctHandler.Handle(cryptoTransferPayload)
}
