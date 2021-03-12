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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/hedera-mirror-client"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/hedera-node-client"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/message"
	transaction2 "github.com/limechain/hedera-eth-bridge-validator/test/mocks/transaction"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	mocks "github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"

	txn "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	fees "github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
)

const (
	topicID         = "0.0.125563"
	accountID       = "0.0.99661"
	pollingInterval = 5
	ethPrivateKey   = "bb9282ba72b55a531fa5e7cc83e92e9055c6905648d673f4d57ad663a317da49"
	submissionTxID  = "0.0.99661--62135596800-0"
	signature       = "f9f9c16aa2ac71b8341d9187c37c2b8dd8152c4a27fe70f8fcf60d56456166ce704c3f1df4831d66e26879a32cb764d928de346418c1f0f116cba14d78a4dfac1b"
	exchangeRate    = 0.00007
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

func getHederaConfig() config.Hedera {
	hederaConfig := config.Hedera{}
	hederaConfig.Client.BaseGasUsage = 130000
	hederaConfig.Client.GasPerValidator = 54000
	return hederaConfig
}

func getEthereumConfig() config.Ethereum {
	ethereumConfig := config.Ethereum{
		NodeUrl: "wss://ropsten.infura.io/ws/v3/8b64d65996d24dc0aae2e0c6029e5a9b",
	}
	return ethereumConfig
}

func InitializeHandler() (*Handler, *transaction2.MockTransactionRepository, *hedera_node_client.MockHederaNodeClient, *hedera_mirror_client.MockHederaMirrorClient, *fees.Calculator) {
	cthConfig := config.CryptoTransferHandler{
		TopicId:         topicID,
		PollingInterval: pollingInterval,
	}
	mocks.Setup()
	ethSigner := eth.NewEthSigner(ethPrivateKey)
	feeCalculator := fees.NewCalculator(mocks.MExchangeRateProvider, getHederaConfig(), mocks.MBridgeContractService)
	ethClient := ethereum.NewClient(getEthereumConfig())
	transactionRepo := &transaction2.MockTransactionRepository{}
	messageRepo := &message.MockMessageRepository{}
	hederaNodeClient := &hedera_node_client.MockHederaNodeClient{}
	hederaMirrorClient := &hedera_mirror_client.MockHederaMirrorClient{}
	processingService := process.NewProcessingService(ethClient,
		transactionRepo,
		messageRepo,
		addresses,
		feeCalculator,
		ethSigner,
		hederaNodeClient,
		topicID)

	return NewHandler(cthConfig, ethSigner, hederaMirrorClient, hederaNodeClient, transactionRepo, processingService), transactionRepo, hederaNodeClient, hederaMirrorClient, feeCalculator
}

func GetTestData() (protomsg.CryptoTransferMessage, hedera.TopicID, hedera.AccountID, []byte, []byte) {
	ctm := protomsg.CryptoTransferMessage{}
	topicID, _ := hedera.TopicIDFromString(topicID)
	accID, _ := hedera.AccountIDFromString(accountID)

	cryptoTransferPayload := []byte{10, 30, 48, 46, 48, 46, 57, 57, 54, 54, 49, 45, 49, 54, 49, 51, 54, 54, 50, 55, 54, 52, 45, 51, 55, 52, 53, 48, 50, 48, 54, 51, 18, 42, 48, 120, 55, 99, 70, 97, 101, 50, 100, 101, 70, 49, 53, 100, 70, 56, 54, 67, 102, 100, 65, 57, 102, 50, 100, 50, 53, 65, 51, 54, 49, 102, 49, 49, 50, 51, 70, 52, 50, 101, 68, 68, 26, 10, 49, 48, 48, 48, 48, 48, 48, 48, 48, 48, 34, 9, 54, 48, 48, 48, 48, 48, 48, 48, 48, 42, 1, 49}
	topicSubmissionMessageBytes := []byte{0x12, 0xe8, 0x1, 0xa, 0x1e, 0x30, 0x2e, 0x30, 0x2e, 0x39, 0x39, 0x36, 0x36, 0x31, 0x2d, 0x31, 0x36, 0x31, 0x33, 0x36, 0x36, 0x32, 0x37, 0x36, 0x34, 0x2d, 0x33, 0x37, 0x34, 0x35, 0x30, 0x32, 0x30, 0x36, 0x33, 0x12, 0x2a, 0x30, 0x78, 0x37, 0x63, 0x46, 0x61, 0x65, 0x32, 0x64, 0x65, 0x46, 0x31, 0x35, 0x64, 0x46, 0x38, 0x36, 0x43, 0x66, 0x64, 0x41, 0x39, 0x66, 0x32, 0x64, 0x32, 0x35, 0x41, 0x33, 0x36, 0x31, 0x66, 0x31, 0x31, 0x32, 0x33, 0x46, 0x34, 0x32, 0x65, 0x44, 0x44, 0x1a, 0xa, 0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x22, 0x9, 0x36, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x2a, 0x82, 0x1, 0x66, 0x39, 0x66, 0x39, 0x63, 0x31, 0x36, 0x61, 0x61, 0x32, 0x61, 0x63, 0x37, 0x31, 0x62, 0x38, 0x33, 0x34, 0x31, 0x64, 0x39, 0x31, 0x38, 0x37, 0x63, 0x33, 0x37, 0x63, 0x32, 0x62, 0x38, 0x64, 0x64, 0x38, 0x31, 0x35, 0x32, 0x63, 0x34, 0x61, 0x32, 0x37, 0x66, 0x65, 0x37, 0x30, 0x66, 0x38, 0x66, 0x63, 0x66, 0x36, 0x30, 0x64, 0x35, 0x36, 0x34, 0x35, 0x36, 0x31, 0x36, 0x36, 0x63, 0x65, 0x37, 0x30, 0x34, 0x63, 0x33, 0x66, 0x31, 0x64, 0x66, 0x34, 0x38, 0x33, 0x31, 0x64, 0x36, 0x36, 0x65, 0x32, 0x36, 0x38, 0x37, 0x39, 0x61, 0x33, 0x32, 0x63, 0x62, 0x37, 0x36, 0x34, 0x64, 0x39, 0x32, 0x38, 0x64, 0x65, 0x33, 0x34, 0x36, 0x34, 0x31, 0x38, 0x63, 0x31, 0x66, 0x30, 0x66, 0x31, 0x31, 0x36, 0x63, 0x62, 0x61, 0x31, 0x34, 0x64, 0x37, 0x38, 0x61, 0x34, 0x64, 0x66, 0x61, 0x63, 0x31, 0x62}

	return ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes
}

func Test_Handle_Not_Initial_Transaction(t *testing.T) {
	ctm, topicID, _, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient, _ := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	tx := &transaction.Transaction{
		Model:          gorm.Model{},
		TransactionId:  ctm.TransactionId,
		EthAddress:     ctm.EthAddress,
		Amount:         ctm.Amount,
		Fee:            ctm.Fee,
		Signature:      signature,
		SubmissionTxId: submissionTxID,
		Status:         txRepo.StatusCompleted,
	}

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, nil)

	ctHandler.Handle(cryptoTransferPayload)

	transactionRepo.AssertNotCalled(t, "UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature)
	hederaNodeClient.AssertNotCalled(t, "SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes)
	hederaMirrorClient.AssertNotCalled(t, "GetAccountTransaction", submissionTxID)
}

func Test_Handle_Initial_Transaction(t *testing.T) {
	ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient, _ := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	expectedTransaction := hedera.TransactionID{
		AccountID:  accID,
		ValidStart: time.Time{},
	}

	tx := &transaction.Transaction{
		Model:          gorm.Model{},
		TransactionId:  ctm.TransactionId,
		EthAddress:     ctm.EthAddress,
		Amount:         ctm.Amount,
		Fee:            ctm.Fee,
		Signature:      signature,
		SubmissionTxId: submissionTxID,
		Status:         txRepo.StatusInitial,
	}

	txs := mirror_node.Response{
		Transactions: []mirror_node.Transaction{},
	}

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, nil)
	transactionRepo.On("UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature).Return(nil)
	transactionRepo.On("UpdateStatusInsufficientFee", ctm.TransactionId).Return(nil)
	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(&expectedTransaction, nil)
	hederaMirrorClient.On("GetAccountTransaction", submissionTxID).Return(&txs, nil)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	mocks.MBridgeContractService.On("GetServiceFee").Return(serviceFeePercent)
	mocks.MBridgeContractService.On("GetMembers").Return(addresses)

	ctHandler.Handle(cryptoTransferPayload)
	time.Sleep(time.Second * pollingInterval)

	transactionRepo.AssertCalled(t, "UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature)
	hederaNodeClient.AssertCalled(t, "SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes)
	hederaMirrorClient.AssertCalled(t, "GetAccountTransaction", submissionTxID)
}

func Test_Handle_Failed(t *testing.T) {
	ctm, topicID, _, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient, _ := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	tx := &transaction.Transaction{
		Model:          gorm.Model{},
		TransactionId:  ctm.TransactionId,
		EthAddress:     ctm.EthAddress,
		Amount:         ctm.Amount,
		Fee:            ctm.Fee,
		Signature:      signature,
		SubmissionTxId: submissionTxID,
		Status:         txRepo.StatusInitial,
	}

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, errors.New("Failed to get record by transaction id"))

	ctHandler.Handle(cryptoTransferPayload)

	transactionRepo.AssertNotCalled(t, "UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature)
	transactionRepo.AssertNotCalled(t, "UpdateStatusInsufficientFee", ctm.TransactionId)
	hederaNodeClient.AssertNotCalled(t, "SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes)
	hederaMirrorClient.AssertNotCalled(t, "GetAccountTransaction", submissionTxID)
}

func Test_HandleTopicSubmission(t *testing.T) {
	ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, _, hederaNodeClient, _, _ := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	expectedTransaction := hedera.TransactionID{
		AccountID:  accID,
		ValidStart: time.Time{},
	}

	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(&expectedTransaction, nil)

	transactionID, err := ctHandler.transfersService.HandleTopicSubmission(&ctm, signature)
	submissionTxn := txn.FromHederaTransactionID(transactionID)

	assert.Nil(t, err)
	assert.Equal(t, submissionTxn.String(), submissionTxID)
}

func Test_CheckForTransactionCompletion(t *testing.T) {
	ctm, _, _, cryptoTransferPayload, _ := GetTestData()
	ctHandler, _, _, hederaMirrorClient, _ := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	txs := mirror_node.Response{
		Transactions: []mirror_node.Transaction{},
	}

	hederaMirrorClient.On("GetAccountTransaction", submissionTxID).Return(&txs, nil)

	go ctHandler.checkForTransactionCompletion(ctm.TransactionId, submissionTxID)
	time.Sleep(time.Second * pollingInterval)

	hederaMirrorClient.AssertCalled(t, "GetAccountTransaction", submissionTxID)
}
