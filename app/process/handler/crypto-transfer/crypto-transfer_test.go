package cryptotransfer

import (
	"errors"
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
)

const (
	topicID         = "0.0.125563"
	accountID       = "0.0.99661"
	pollingInterval = 5
	ethPrivateKey   = "bb9282ba72b55a531fa5e7cc83e92e9055c6905648d673f4d57ad663a317da49"
	submissionTxID  = "0.0.99661--62135596800-0"
	signature       = "57ba03ec596908affadeb22b2a471dfbafa45ab9fbc1fb5f71a1485eaeac329508c13fdd58ab42dd4a87dcf7ce998335d2696685f74a28b14331d052561bd6671b"
)

func InitializeHandler() (*CryptoTransferHandler, *mocks.MockTransactionRepository, *mocks.MockHederaNodeClient, *mocks.MockHederaMirrorClient) {
	cthConfig := config.CryptoTransferHandler{
		TopicId:         topicID,
		PollingInterval: pollingInterval,
	}

	ethSigner := eth.NewEthSigner(ethPrivateKey)
	transactionRepo := &mocks.MockTransactionRepository{}
	hederaNodeClient := &mocks.MockHederaNodeClient{}
	hederaMirrorClient := &mocks.MockHederaMirrorClient{}

	return NewCryptoTransferHandler(cthConfig, ethSigner, hederaMirrorClient, hederaNodeClient, transactionRepo), transactionRepo, hederaNodeClient, hederaMirrorClient
}

func GetTestData() (protomsg.CryptoTransferMessage, hedera.TopicID, hedera.AccountID, []byte, []byte) {
	ctm := protomsg.CryptoTransferMessage{}
	topicID, _ := hedera.TopicIDFromString(topicID)
	accID, _ := hedera.AccountIDFromString(accountID)

	cryptoTransferPayload := []byte{10, 30, 48, 46, 48, 46, 57, 57, 54, 54, 49, 45, 49, 54, 49, 51, 51, 56, 57, 49, 50, 49, 45, 49, 51, 53, 54, 51, 49, 52, 54, 49, 18, 42, 48, 120, 55, 99, 70, 97, 101, 50, 100, 101, 70, 49, 53, 100, 70, 56, 54, 67, 102, 100, 65, 57, 102, 50, 100, 50, 53, 65, 51, 54, 49, 102, 49, 49, 50, 51, 70, 52, 50, 101, 68, 68, 24, 144, 78, 34, 13, 49, 49, 50, 54, 50, 50, 49, 50, 51, 55, 50, 49, 49}
	topicSubmissionMessageBytes := []byte{0x12, 0xe3, 0x1, 0xa, 0x1e, 0x30, 0x2e, 0x30, 0x2e, 0x39, 0x39, 0x36, 0x36, 0x31, 0x2d, 0x31, 0x36, 0x31, 0x33, 0x33, 0x38, 0x39, 0x31, 0x32, 0x31, 0x2d, 0x31, 0x33, 0x35, 0x36, 0x33, 0x31, 0x34, 0x36, 0x31, 0x12, 0x2a, 0x30, 0x78, 0x37, 0x63, 0x46, 0x61, 0x65, 0x32, 0x64, 0x65, 0x46, 0x31, 0x35, 0x64, 0x46, 0x38, 0x36, 0x43, 0x66, 0x64, 0x41, 0x39, 0x66, 0x32, 0x64, 0x32, 0x35, 0x41, 0x33, 0x36, 0x31, 0x66, 0x31, 0x31, 0x32, 0x33, 0x46, 0x34, 0x32, 0x65, 0x44, 0x44, 0x18, 0x90, 0x4e, 0x22, 0xd, 0x31, 0x31, 0x32, 0x36, 0x32, 0x32, 0x31, 0x32, 0x33, 0x37, 0x32, 0x31, 0x31, 0x2a, 0x82, 0x1, 0x35, 0x37, 0x62, 0x61, 0x30, 0x33, 0x65, 0x63, 0x35, 0x39, 0x36, 0x39, 0x30, 0x38, 0x61, 0x66, 0x66, 0x61, 0x64, 0x65, 0x62, 0x32, 0x32, 0x62, 0x32, 0x61, 0x34, 0x37, 0x31, 0x64, 0x66, 0x62, 0x61, 0x66, 0x61, 0x34, 0x35, 0x61, 0x62, 0x39, 0x66, 0x62, 0x63, 0x31, 0x66, 0x62, 0x35, 0x66, 0x37, 0x31, 0x61, 0x31, 0x34, 0x38, 0x35, 0x65, 0x61, 0x65, 0x61, 0x63, 0x33, 0x32, 0x39, 0x35, 0x30, 0x38, 0x63, 0x31, 0x33, 0x66, 0x64, 0x64, 0x35, 0x38, 0x61, 0x62, 0x34, 0x32, 0x64, 0x64, 0x34, 0x61, 0x38, 0x37, 0x64, 0x63, 0x66, 0x37, 0x63, 0x65, 0x39, 0x39, 0x38, 0x33, 0x33, 0x35, 0x64, 0x32, 0x36, 0x39, 0x36, 0x36, 0x38, 0x35, 0x66, 0x37, 0x34, 0x61, 0x32, 0x38, 0x62, 0x31, 0x34, 0x33, 0x33, 0x31, 0x64, 0x30, 0x35, 0x32, 0x35, 0x36, 0x31, 0x62, 0x64, 0x36, 0x36, 0x37, 0x31, 0x62}

	return ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes
}

func Test_Handle_Not_Initial_Transaction(t *testing.T) {
	ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient := InitializeHandler()

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
		Status:         txRepo.StatusCompleted,
	}

	txs := txn.HederaTransactions{
		Transactions: []txn.HederaTransaction{},
	}

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, nil)
	transactionRepo.On("UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature).Return(nil)
	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(&expectedTransaction, nil)
	hederaMirrorClient.On("GetAccountTransaction", submissionTxID).Return(&txs, nil)

	ctHandler.Handle(cryptoTransferPayload)

	transactionRepo.AssertNotCalled(t, "UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature)
}

func Test_Handle_Initial_Transaction(t *testing.T) {
	ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient := InitializeHandler()

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

	txs := txn.HederaTransactions{
		Transactions: []txn.HederaTransaction{},
	}

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, nil)
	transactionRepo.On("UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature).Return(nil)
	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(&expectedTransaction, nil)
	hederaMirrorClient.On("GetAccountTransaction", submissionTxID).Return(&txs, nil)

	ctHandler.Handle(cryptoTransferPayload)

	transactionRepo.AssertCalled(t, "UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature)
}

func Test_Handle_Failed(t *testing.T) {
	ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient := InitializeHandler()

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

	txs := txn.HederaTransactions{
		Transactions: []txn.HederaTransaction{},
	}

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, errors.New("Failed to get record by transaction id"))
	transactionRepo.On("UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature).Return(nil)
	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(&expectedTransaction, nil)
	hederaMirrorClient.On("GetAccountTransaction", submissionTxID).Return(&txs, nil)

	ctHandler.Handle(cryptoTransferPayload)

	transactionRepo.AssertNotCalled(t, "UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature)
	transactionRepo.AssertNotCalled(t, "UpdateStatusInsufficientFee", ctm.TransactionId)
	hederaNodeClient.AssertNotCalled(t, "SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes)
	hederaMirrorClient.AssertNotCalled(t, "GetAccountTransaction", submissionTxID)
}

func Test_HandleTopicSubmission(t *testing.T) {
	ctm, topicID, accID, cryptoTransferPayload, topicSubmissionMessageBytes := GetTestData()
	ctHandler, transactionRepo, hederaNodeClient, hederaMirrorClient := InitializeHandler()

	proto.Unmarshal(cryptoTransferPayload, &ctm)

	expectedTransaction := hedera.TransactionID{
		AccountID:  accID,
		ValidStart: time.Time{},
	}

	txs := txn.HederaTransactions{
		Transactions: []txn.HederaTransaction{},
	}

	transactionRepo.On("UpdateStatusSignatureSubmitted", ctm.TransactionId, submissionTxID, signature).Return(nil)
	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(&expectedTransaction, nil)
	hederaMirrorClient.On("GetAccountTransaction", submissionTxID).Return(&txs, nil)

	transactionID, err := ctHandler.handleTopicSubmission(&ctm, signature)
	submissionTxn := txn.FromHederaTransactionID(transactionID)

	assert.Nil(t, err)
	assert.Equal(t, submissionTxn.String(), submissionTxID)
}
