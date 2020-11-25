package cryptotransfer

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
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

// Recover mechanism
func (cth *CryptoTransferHandler) Recover(q *queue.Queue) {
	log.Info("[Recovery - CryptoTransfer Handler] Executing Recovery mechanism for CryptoTransfer Handler.")
	log.Info("[Recovery - CryptoTransfer Handler] Database GET [PENDING] [SUBMITTED] transactions.")

	transactions, err := cth.transactionRepo.GetPendingOrSubmittedTransactions()
	if err != nil {
		log.Errorf("[Recovery - CryptoTransfer] Failed to Database GET transactions. Error [%s]", err)
		return
	}

	for _, transaction := range transactions {
		if transaction.Status == txRepo.StatusPending {
			log.Infof("[Recovery - CryptoTransfer Handler] Submit TransactionID [%s] to Handler.", transaction.TransactionId)
			go cth.submitTx(transaction, q)
		} else {
			go cth.checkForTransactionCompletion(transaction.TransactionId, transaction.SubmissionTxId)
		}
	}
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
		log.Errorf("Failed to get record with TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
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

		if dbTransaction.Status != txRepo.StatusPending {
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

	hash := crypto.Keccak256([]byte(fmt.Sprintf("%s-%s-%d-%s", ctm.TransactionId, ctm.EthAddress, ctm.Amount, ctm.Fee)))
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
	topicMessageSubmissionTxId := tx.FromHederaTransactionID(topicMessageSubmissionTx)

	err = cth.transactionRepo.UpdateStatusSubmitted(ctm.TransactionId, topicMessageSubmissionTxId.String(), encodedSignature)
	if err != nil {
		log.Errorf("Failed to update submitted status for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	go cth.checkForTransactionCompletion(ctm.TransactionId, topicMessageSubmissionTxId.String())
}

func (cth *CryptoTransferHandler) checkForTransactionCompletion(transactionId string, topicMessageSubmissionTxId string) {
	log.Infof("Checking for mirror node completion for TransactionID [%s] and Topic Submission TransactionID [%s].",
		transactionId,
		fmt.Sprintf(topicMessageSubmissionTxId))

	for {
		txs, err := cth.hederaMirrorClient.GetAccountTransaction(topicMessageSubmissionTxId)
		if err != nil {
			log.Errorf("Error while trying to get account TransactionID [%s]. Error [%s].", topicMessageSubmissionTxId, err.Error())
			return
		}

		if len(txs.Transactions) > 0 {
			success := false
			for _, tx := range txs.Transactions {
				if tx.Result == hedera.StatusSuccess.String() {
					success = true
				}
			}

			if success {
				log.Infof("Completing status for Transaction ID [%s].", transactionId)
				err := cth.transactionRepo.UpdateStatusCompleted(transactionId)
				if err != nil {
					log.Errorf("Failed to update completed status for TransactionID [%s]. Error [%s].", transactionId, err)
				}
			} else {
				log.Infof("Cancelling unsuccessful Transaction ID [%s], Submission Message TxID [%s] with Result [%s].", transactionId, topicMessageSubmissionTxId)
				err := cth.transactionRepo.UpdateStatusCancelled(transactionId)
				if err != nil {
					log.Errorf("Failed to cancel transaction with TransactionID [%s]. Error [%s].", transactionId, err)
				}
			}
			return
		}

		time.Sleep(cth.pollingInterval * time.Second)
	}
}

func (cth *CryptoTransferHandler) submitTx(tx *txRepo.Transaction, q *queue.Queue) {
	ctm := &protomsg.CryptoTransferMessage{
		TransactionId: tx.TransactionId,
		EthAddress:    tx.EthAddress,
		Amount:        tx.Amount,
		Fee:           tx.Fee,
	}
	publisher.Publish(ctm, "HCS_CRYPTO_TRANSFER", cth.topicID, q)
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
