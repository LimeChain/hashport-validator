package crypto_transfer

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	mocks "github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"gorm.io/gorm"
)

func Test_CryptoTransferHandler(t *testing.T) {
	cthConfig := config.CryptoTransferHandler{
		TopicId:         "0.0.125563",
		PollingInterval: 5,
	}

	var ethPk string = "bb9282ba72b55a531fa5e7cc83e92e9055c6905648d673f4d57ad663a317da49"

	transactionRepo := &mocks.MockTransactionRepository{}
	hederaNodeClient := &mocks.MockHederaNodeClient{}
	hederaMirrorClient := &mocks.MockHederaMirrorClient{}
	ethSigner := eth.NewEthSigner(ethPk)

	ctHandler := NewCryptoTransferHandler(cthConfig, ethSigner, hederaMirrorClient, hederaNodeClient, transactionRepo)
	payload := []byte{10, 30, 48, 46, 48, 46, 57, 57, 54, 54, 49, 45, 49, 54, 49, 51, 51, 56, 57, 49, 50, 49, 45, 49, 51, 53, 54, 51, 49, 52, 54, 49, 18, 42, 48, 120, 55, 99, 70, 97, 101, 50, 100, 101, 70, 49, 53, 100, 70, 56, 54, 67, 102, 100, 65, 57, 102, 50, 100, 50, 53, 65, 51, 54, 49, 102, 49, 49, 50, 51, 70, 52, 50, 101, 68, 68, 24, 144, 78, 34, 13, 49, 49, 50, 54, 50, 50, 49, 50, 51, 55, 50, 49, 49}

	var ctm protomsg.CryptoTransferMessage
	proto.Unmarshal(payload, &ctm)
	tx := &transaction.Transaction{
		Model:          gorm.Model{},
		TransactionId:  ctm.TransactionId,
		EthAddress:     ctm.EthAddress,
		Amount:         ctm.Amount,
		Fee:            ctm.Fee,
		Signature:      "",
		SubmissionTxId: "",
		Status:         "INITIAL",
		EthHash:        "",
	}

	topicID, _ := hedera.TopicIDFromString("0.0.125563")
	accID, _ := hedera.AccountIDFromString("0.0.99661")
	topicSubmissionMessageBytes := []byte{0x12, 0xe3, 0x1, 0xa, 0x1e, 0x30, 0x2e, 0x30, 0x2e, 0x39, 0x39, 0x36, 0x36, 0x31, 0x2d, 0x31, 0x36, 0x31, 0x33, 0x33, 0x38, 0x39, 0x31, 0x32, 0x31, 0x2d, 0x31, 0x33, 0x35, 0x36, 0x33, 0x31, 0x34, 0x36, 0x31, 0x12, 0x2a, 0x30, 0x78, 0x37, 0x63, 0x46, 0x61, 0x65, 0x32, 0x64, 0x65, 0x46, 0x31, 0x35, 0x64, 0x46, 0x38, 0x36, 0x43, 0x66, 0x64, 0x41, 0x39, 0x66, 0x32, 0x64, 0x32, 0x35, 0x41, 0x33, 0x36, 0x31, 0x66, 0x31, 0x31, 0x32, 0x33, 0x46, 0x34, 0x32, 0x65, 0x44, 0x44, 0x18, 0x90, 0x4e, 0x22, 0xd, 0x31, 0x31, 0x32, 0x36, 0x32, 0x32, 0x31, 0x32, 0x33, 0x37, 0x32, 0x31, 0x31, 0x2a, 0x82, 0x1, 0x35, 0x37, 0x62, 0x61, 0x30, 0x33, 0x65, 0x63, 0x35, 0x39, 0x36, 0x39, 0x30, 0x38, 0x61, 0x66, 0x66, 0x61, 0x64, 0x65, 0x62, 0x32, 0x32, 0x62, 0x32, 0x61, 0x34, 0x37, 0x31, 0x64, 0x66, 0x62, 0x61, 0x66, 0x61, 0x34, 0x35, 0x61, 0x62, 0x39, 0x66, 0x62, 0x63, 0x31, 0x66, 0x62, 0x35, 0x66, 0x37, 0x31, 0x61, 0x31, 0x34, 0x38, 0x35, 0x65, 0x61, 0x65, 0x61, 0x63, 0x33, 0x32, 0x39, 0x35, 0x30, 0x38, 0x63, 0x31, 0x33, 0x66, 0x64, 0x64, 0x35, 0x38, 0x61, 0x62, 0x34, 0x32, 0x64, 0x64, 0x34, 0x61, 0x38, 0x37, 0x64, 0x63, 0x66, 0x37, 0x63, 0x65, 0x39, 0x39, 0x38, 0x33, 0x33, 0x35, 0x64, 0x32, 0x36, 0x39, 0x36, 0x36, 0x38, 0x35, 0x66, 0x37, 0x34, 0x61, 0x32, 0x38, 0x62, 0x31, 0x34, 0x33, 0x33, 0x31, 0x64, 0x30, 0x35, 0x32, 0x35, 0x36, 0x31, 0x62, 0x64, 0x36, 0x36, 0x37, 0x31, 0x62}
	hederaTrans := new(hedera.TransactionID)
	hederaTrans.AccountID = accID

	transactionRepo.On("GetByTransactionId", ctm.TransactionId).Return(tx, nil)
	transactionRepo.On("UpdateStatusSignatureSubmitted", "0.0.99661-1613389121-135631461", "0.0.99661--62135596800-0", "57ba03ec596908affadeb22b2a471dfbafa45ab9fbc1fb5f71a1485eaeac329508c13fdd58ab42dd4a87dcf7ce998335d2696685f74a28b14331d052561bd6671b").Return(nil)
	hederaNodeClient.On("SubmitTopicConsensusMessage", topicID, topicSubmissionMessageBytes).Return(hederaTrans, nil)
	ctHandler.Handle(payload)

}
