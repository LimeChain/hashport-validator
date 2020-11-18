package crypto_transfer

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

// Crypto Transfer event handler
type CryptoTransferHandler struct {
	topicID         hedera.ConsensusTopicID
	ethSigner       *eth.Signer
	hederaClient    *hederaClient.HederaNodeClient
	transactionRepo repositories.TransactionRepository
}

func (cth *CryptoTransferHandler) Handle(payload []byte) error {
	// TODO: logs instead of todos
	var ctm protomsg.CryptoTransferMessage
	err := proto.Unmarshal(payload, &ctm)
	if err != nil {
		return err
	}

	exists, err := cth.transactionRepo.Exists(ctm.TransactionId)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	err = cth.transactionRepo.Create(&ctm)
	if err != nil {
		return err
	}

	validFee, err := fees.ValidateExecutionFee(ctm.Fee)
	if err != nil {
		return err
	}

	if !validFee {
		err = cth.transactionRepo.UpdateStatusCancelled(ctm.TransactionId)
		if err != nil {
			return err
		}

		return nil
	}

	hash := crypto.Keccak256([]byte(ctm.String()))
	signature, err := cth.ethSigner.Sign(hash)
	if err != nil {
		return err
	}

	encodedSignature := hex.EncodeToString(signature)

	topicMsgSubmissionTxId, err := cth.handleTopicSubmission(&ctm, encodedSignature)
	if err != nil {
		return err
	}

	return cth.transactionRepo.UpdateStatusSubmitted(ctm.TransactionId, topicMsgSubmissionTxId, encodedSignature)
}

func (cth *CryptoTransferHandler) handleTopicSubmission(message *protomsg.CryptoTransferMessage, signature string) (string, error) {
	topicSigMessage := &protomsg.TopicSignatureMessage{
		TransactionId: message.TransactionId,
		EthAddress:    message.EthAddress,
		Amount:        message.Amount,
		Fee:           message.Fee,
		Signature:     signature,
	}

	topicSigMessageBytes, err := proto.Marshal(topicSigMessage)
	if err != nil {
		return "", err
	}

	return cth.hederaClient.SubmitTopicConsensusMessage(cth.topicID, topicSigMessageBytes)
}

func NewCryptoTransferHandler(
	config config.CryptoTransferHandler,
	ethSigner *eth.Signer,
	hederaClient *hederaClient.HederaNodeClient,
	transactionRepository repositories.TransactionRepository) *CryptoTransferHandler {
	topicID, err := hedera.TopicIDFromString(config.TopicId)
	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid Topic ID provided: [%s]", config.TopicId))
	}

	return &CryptoTransferHandler{
		topicID:         topicID,
		ethSigner:       ethSigner,
		hederaClient:    hederaClient,
		transactionRepo: transactionRepository,
	}
}
