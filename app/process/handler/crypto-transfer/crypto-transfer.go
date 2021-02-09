package cryptotransfer

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	exchangerate "github.com/limechain/hedera-eth-bridge-validator/app/clients/exchange-rate"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// Crypto Transfer event handler
type CryptoTransferHandler struct {
	pollingInterval     time.Duration
	topicID             hedera.TopicID
	ethSigner           *eth.Signer
	ethClient           *ethereum.EthereumClient
	hederaMirrorClient  *hederaClient.HederaMirrorClient
	hederaNodeClient    *hederaClient.HederaNodeClient
	transactionRepo     repositories.TransactionRepository
	logger              *log.Entry
	exchangeRateService *exchangerate.ExchangeRateProvider
}

func NewCryptoTransferHandler(
	c config.CryptoTransferHandler,
	ethSigner *eth.Signer,
	ethClient *ethereum.EthereumClient,
	hederaMirrorClient *hederaClient.HederaMirrorClient,
	hederaNodeClient *hederaClient.HederaNodeClient,
	transactionRepository repositories.TransactionRepository,
	exchangeRateService *exchangerate.ExchangeRateProvider) *CryptoTransferHandler {
	topicID, err := hedera.TopicIDFromString(c.TopicId)
	if err != nil {
		log.Fatalf("Invalid Topic ID provided: [%s]", c.TopicId)
	}

	return &CryptoTransferHandler{
		pollingInterval:     c.PollingInterval,
		topicID:             topicID,
		ethSigner:           ethSigner,
		hederaMirrorClient:  hederaMirrorClient,
		hederaNodeClient:    hederaNodeClient,
		transactionRepo:     transactionRepository,
		logger:              config.GetLoggerFor("Account Transfer Handler"),
		exchangeRateService: exchangeRateService,
		ethClient:           ethClient,
	}
}

// Recover mechanism
func (cth *CryptoTransferHandler) Recover(q *queue.Queue) {
	cth.logger.Info("[Recovery] Executing Recovery mechanism for CryptoTransfer Handler.")
	cth.logger.Info("[Recovery] Database GET [PENDING] [SUBMITTED] transactions.")

	transactions, err := cth.transactionRepo.GetPendingOrSubmittedTransactions()
	if err != nil {
		cth.logger.Errorf("[Recovery] Failed to Database GET transactions. Error [%s]", err)
		return
	}

	for _, transaction := range transactions {
		if transaction.Status == txRepo.StatusPending {
			cth.logger.Infof("[Recovery] Submit TransactionID [%s] to Handler.", transaction.TransactionId)
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
		cth.logger.Errorf("Failed to parse incoming payload. Error [%s].", err)
		return
	}

	dbTransaction, err := cth.transactionRepo.GetByTransactionId(ctm.TransactionId)
	if err != nil {
		cth.logger.Errorf("Failed to get record with TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	if dbTransaction == nil {
		cth.logger.Debugf("Persisting TX with ID [%s].", ctm.TransactionId)

		err = cth.transactionRepo.Create(&ctm)
		if err != nil {
			cth.logger.Errorf("Failed to create a transaction record for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
			return
		}
	} else {
		cth.logger.Debugf("Transaction with TransactionID [%s] has already been added. Continuing execution.", ctm.TransactionId)

		if dbTransaction.Status != txRepo.StatusPending {
			cth.logger.Infof("Previously added Transaction with TransactionID [%s] has status [%s]. Skipping further execution.", ctm.TransactionId, dbTransaction.Status)
			return
		}
	}

	exchangeRate, err := cth.exchangeRateService.GetRate()
	if err != nil {
		cth.logger.Errorf("Failed to retrieve exchange rate. Error [%s].", ctm.TransactionId, err)
		return
	}

	validFee, err := fees.ValidateExecutionFee(ctm.Fee, ctm.Amount*10/100, ctm.Amount)
	if err != nil {
		cth.logger.Errorf("Failed to validate fee for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	if !validFee {
		cth.logger.Infof("Cancelling transaction [%s] due to invalid fee provided: [%s]", ctm.TransactionId, ctm.Fee)
		err = cth.transactionRepo.UpdateStatusCancelled(ctm.TransactionId)
		if err != nil {
			cth.logger.Errorf("Failed to cancel transaction with TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
			return
		}

		return
	}

	estimatedGas, err := cth.ethClient.EstimateGas(ctm.Amount)
	if err != nil {
		cth.logger.Errorf("Failed to estimate gas for TransactionID [%s]. Error [%s]", ctm.TransactionId, err)
		return
	}

	slowGasPrice, err := cth.ethClient.GetSlowGasPrice()
	if err != nil {
		cth.logger.Errorf("Failed to get slow gas price. Error [%s]", ctm.TransactionId, err)
		return
	}

	ethPrice := slowGasPrice * estimatedGas

	HBarTxFee := float64(ethPrice) / exchangeRate
	TxFee, err := strconv.ParseFloat(ctm.Fee, 64)
	if err != nil {
		cth.logger.Errorf("Could not parse transaction fee: [%s]. Error: [%s]", ctm.Fee, err)
		return
	}

	if HBarTxFee >= TxFee {
		cth.logger.Errorf("Insufficient transaction fee: [%s]. Error: [%s]", ctm.Fee, err)
		return
	}

	encodedData, err := ethhelper.EncodeData(&ctm)
	if err != nil {
		cth.logger.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	hash := crypto.Keccak256(encodedData)
	signature, err := cth.ethSigner.Sign(hash)
	if err != nil {
		cth.logger.Errorf("Failed to sign transaction data for TransactionID [%s], Hash [%s]. Error [%s].", ctm.TransactionId, hash, err)
		return
	}

	encodedSignature := hex.EncodeToString(signature)

	topicMessageSubmissionTx, err := cth.handleTopicSubmission(&ctm, encodedSignature)
	if err != nil {
		cth.logger.Errorf("Failed to submit topic consensus message for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}
	topicMessageSubmissionTxId := tx.FromHederaTransactionID(topicMessageSubmissionTx)

	err = cth.transactionRepo.UpdateStatusSubmitted(ctm.TransactionId, topicMessageSubmissionTxId.String(), encodedSignature)
	if err != nil {
		cth.logger.Errorf("Failed to update submitted status for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}

	go cth.checkForTransactionCompletion(ctm.TransactionId, topicMessageSubmissionTxId.String())
}

func (cth *CryptoTransferHandler) checkForTransactionCompletion(transactionId string, topicMessageSubmissionTxId string) {
	cth.logger.Debugf("Checking for mirror node completion for TransactionID [%s] and Topic Submission TransactionID [%s].",
		transactionId,
		fmt.Sprintf(topicMessageSubmissionTxId))

	for {
		txs, err := cth.hederaMirrorClient.GetAccountTransaction(topicMessageSubmissionTxId)
		if err != nil {
			cth.logger.Errorf("Error while trying to get account TransactionID [%s]. Error [%s].", topicMessageSubmissionTxId, err.Error())
			return
		}

		if len(txs.Transactions) > 0 {
			success := false
			for _, transaction := range txs.Transactions {
				if transaction.Result == hedera.StatusSuccess.String() {
					success = true
					break
				}
			}

			if success {
				cth.logger.Debugf("Updating status to completed for TX ID [%s] and Topic Submission ID [%s].", transactionId, fmt.Sprintf(topicMessageSubmissionTxId))
				err := cth.transactionRepo.UpdateStatusCompleted(transactionId)
				if err != nil {
					cth.logger.Errorf("Failed to update completed status for TransactionID [%s]. Error [%s].", transactionId, err)
				}
			} else {
				cth.logger.Infof("Cancelling unsuccessful Transaction ID [%s], Submission Message TxID [%s].", transactionId, topicMessageSubmissionTxId)
				err := cth.transactionRepo.UpdateStatusCancelled(transactionId)
				if err != nil {
					cth.logger.Errorf("Failed to cancel transaction with TransactionID [%s]. Error [%s].", transactionId, err)
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
	topicSigMessage := &protomsg.TopicEthSignatureMessage{
		TransactionId: message.TransactionId,
		EthAddress:    message.EthAddress,
		Amount:        message.Amount,
		Fee:           message.Fee,
		Signature:     signature,
	}

	topicSubmissionMessage := &protomsg.TopicSubmissionMessage{
		Type:    protomsg.TopicSubmissionType_EthSignature,
		Message: &protomsg.TopicSubmissionMessage_TopicSignatureMessage{TopicSignatureMessage: topicSigMessage},
	}

	topicSubmissionMessageBytes, err := proto.Marshal(topicSubmissionMessage)
	if err != nil {
		return nil, err
	}

	cth.logger.Infof("Submitting Signature for TX ID [%s] on Topic [%s]", message.TransactionId, cth.topicID)
	return cth.hederaNodeClient.SubmitTopicConsensusMessage(cth.topicID, topicSubmissionMessageBytes)
}
