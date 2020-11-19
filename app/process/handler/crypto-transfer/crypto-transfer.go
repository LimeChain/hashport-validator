package cryptotransfer

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

func (cth *CryptoTransferHandler) Handle(payload []byte) {
	var ctm protomsg.CryptoTransferMessage
	err := proto.Unmarshal(payload, &ctm)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to parse incoming payload. Error [%s].", err))
		return
	}

	exists, err := cth.transactionRepo.Exists(ctm.TransactionId)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to check existence of record with TransactionID [%s]. Error [%s].", ctm.TransactionId, err))
		return
	}

	if exists {
		log.Info(fmt.Sprintf("Transaction with TransactionID [%s] has already been added. Skipping further execution.", ctm.TransactionId))
		return
	}

	log.Info(fmt.Sprintf("Creating a transaction record for TransactionID [%s].", ctm.TransactionId))
	err = cth.transactionRepo.Create(&ctm)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create a transaction record for TransactionID [%s]. Error [%s].", ctm.TransactionId, err))
		return
	}

	validFee, err := fees.ValidateExecutionFee(ctm.Fee)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to validate fee for TransactionID [%s]. Error [%s].", ctm.TransactionId, err))
		return
	}

	if !validFee {
		log.Info(fmt.Sprintf("Cancelling transaction [%s] due to invalid fee provided: [%s]", ctm.TransactionId, ctm.Fee))
		err = cth.transactionRepo.UpdateStatusCancelled(ctm.TransactionId)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to cancel transaction with TransactionID [%s]. Error [%s].", ctm.TransactionId, err))
			return
		}

		return
	}

	hash := crypto.Keccak256([]byte(ctm.String()))
	signature, err := cth.ethSigner.Sign(hash)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to sign transaction data for TransactionID [%s], Hash [%s]. Error [%s].", ctm.TransactionId, hash, err))
		return
	}

	encodedSignature := hex.EncodeToString(signature)

	topicMsgSubmissionTxId, err := cth.handleTopicSubmission(&ctm, encodedSignature)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to submit topic consensus message for TransactionID [%s]. Error [%s].", ctm.TransactionId, err))
		return
	}

	err = cth.transactionRepo.UpdateStatusSubmitted(ctm.TransactionId, topicMsgSubmissionTxId, encodedSignature)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to update submitted status for TransactionID [%s]. Error [%s].", ctm.TransactionId, err))
	}
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

	log.Info(fmt.Sprintf("Submitting Topic Consensus Message for Topic ID [%s] and Transaction ID [%s]", cth.topicID, message.TransactionId))
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
