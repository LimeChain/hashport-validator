package cryptotransfer

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// Crypto Transfer event handler
type CryptoTransferHandler struct {
	pollingInterval    time.Duration
	topicID            hedera.ConsensusTopicID
	ethSigner          *eth.Signer
	hederaMirrorClient *hederaClient.HederaMirrorClient
	hederaNodeClient   *hederaClient.HederaNodeClient
	transactionRepo    repositories.TransactionRepository
}

func (cth *CryptoTransferHandler) Handle(payload []byte) {
	var ctm protomsg.CryptoTransferMessage
	err := proto.Unmarshal(payload, &ctm)
	if err != nil {
		log.Errorf("Failed to parse incoming payload. Error [%s].", err)
		return
	}

	dbTransaction, err := cth.transactionRepo.GetByTransactionId(ctm.TransactionId)
	if err != nil {
		log.Errorf("Failed to check existence of record with TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	if dbTransaction == nil {
		log.Infof("Creating a transaction record for TransactionID [%s].", ctm.TransactionId)

		err = cth.transactionRepo.Create(&ctm)
		if err != nil {
			log.Errorf("Failed to create a transaction record for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
			return
		}
	} else {
		log.Infof("Transaction with TransactionID [%s] has already been added. Continuing execution.", ctm.TransactionId)

		if dbTransaction.Status != transaction.StatusPending {
			log.Infof("Previously added Transaction with TransactionID [%s] has status [%s]. Skipping further execution.", ctm.TransactionId, dbTransaction.Status)
			return
		}
	}

	validFee, err := fees.ValidateExecutionFee(ctm.Fee)
	if err != nil {
		log.Errorf("Failed to validate fee for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	if !validFee {
		log.Infof("Cancelling transaction [%s] due to invalid fee provided: [%s]", ctm.TransactionId, ctm.Fee)
		err = cth.transactionRepo.UpdateStatusCancelled(ctm.TransactionId)
		if err != nil {
			log.Errorf("Failed to cancel transaction with TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
			return
		}

		return
	}

	hash := crypto.Keccak256([]byte(ctm.String()))
	signature, err := cth.ethSigner.Sign(hash)
	if err != nil {
		log.Errorf("Failed to sign transaction data for TransactionID [%s], Hash [%s]. Error [%s].", ctm.TransactionId, hash, err)
		return
	}

	encodedSignature := hex.EncodeToString(signature)

	topicMessageSubmissionTx, err := cth.handleTopicSubmission(&ctm, encodedSignature)
	if err != nil {
		log.Errorf("Failed to submit topic consensus message for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	err = cth.transactionRepo.UpdateStatusSubmitted(ctm.TransactionId, topicMessageSubmissionTx.String(), encodedSignature)
	if err != nil {
		log.Errorf("Failed to update submitted status for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	go cth.checkForTransactionCompletion(ctm.TransactionId, topicMessageSubmissionTx)
}

func (cth *CryptoTransferHandler) checkForTransactionCompletion(transactionId string, topicMessageSubmissionTx *hedera.TransactionID) {
	topicMessageSubmissionTxId := formatTxId(topicMessageSubmissionTx)
	timestamp := strconv.Itoa(int(topicMessageSubmissionTx.ValidStart.Unix()))

	log.Infof("Checking for mirror node completion for TransactionID [%s] and Topic Submission TransactionID [%s].", transactionId, topicMessageSubmissionTxId)

	for {
		txs, err := cth.hederaMirrorClient.GetAccountConsensusSubmitMessagesTransactionsAfterDate(topicMessageSubmissionTx.AccountID, timestamp)
		if err != nil {
			log.Error("Error while trying to get account transactions after data: [%s].", err.Error())
			return
		}

		if len(txs.Transactions) > 0 {
			for _, tx := range txs.Transactions {
				if tx.TransactionID == topicMessageSubmissionTxId {
					if tx.Result == hedera.StatusSuccess.String() {
						log.Infof("Completing status for Transaction ID [%s].", transactionId)
						err := cth.transactionRepo.UpdateStatusCompleted(transactionId)
						if err != nil {
							log.Errorf("Failed to update completed status for TransactionID [%s]. Error [%s].", transactionId, err)
						}
					} else {
						log.Infof("Cancelling unsuccessful Transaction ID [%s], Submission Message TxID [%s] with Result [%s].", transactionId, topicMessageSubmissionTxId, tx.Result)
						err := cth.transactionRepo.UpdateStatusCancelled(transactionId)
						if err != nil {
							log.Errorf("Failed to cancel transaction with TransactionID [%s]. Error [%s].", transactionId, err)
						}
					}
					return
				}
			}
		}

		time.Sleep(cth.pollingInterval * time.Second)
	}
}

func (cth *CryptoTransferHandler) handleTopicSubmission(message *protomsg.CryptoTransferMessage, signature string) (*hedera.TransactionID, error) {
	topicSigMessage := &protomsg.TopicSignatureMessage{
		TransactionId: message.TransactionId,
		EthAddress:    message.EthAddress,
		Amount:        message.Amount,
		Fee:           message.Fee,
		Signature:     signature,
	}

	topicSigMessageBytes, err := proto.Marshal(topicSigMessage)
	if err != nil {
		return nil, err
	}

	log.Infof("Submitting Topic Consensus Message for Topic ID [%s] and Transaction ID [%s]", cth.topicID, message.TransactionId)
	return cth.hederaNodeClient.SubmitTopicConsensusMessage(cth.topicID, topicSigMessageBytes)
}

func NewCryptoTransferHandler(
	config config.CryptoTransferHandler,
	ethSigner *eth.Signer,
	hederaMirrorClient *hederaClient.HederaMirrorClient,
	hederaNodeClient *hederaClient.HederaNodeClient,
	transactionRepository repositories.TransactionRepository) *CryptoTransferHandler {
	topicID, err := hedera.TopicIDFromString(config.TopicId)
	if err != nil {
		log.Fatalf("Invalid Topic ID provided: [%s]", config.TopicId)
	}

	return &CryptoTransferHandler{
		pollingInterval:    config.PollingInterval,
		topicID:            topicID,
		ethSigner:          ethSigner,
		hederaMirrorClient: hederaMirrorClient,
		hederaNodeClient:   hederaNodeClient,
		transactionRepo:    transactionRepository,
	}
}

func formatTxId(txId *hedera.TransactionID) string {
	return fmt.Sprintf("%s-%v-%v",
		txId.AccountID.String(),
		txId.ValidStart.Unix(),
		int32(txId.ValidStart.UnixNano()-(txId.ValidStart.Unix()*1e+9)))
}
