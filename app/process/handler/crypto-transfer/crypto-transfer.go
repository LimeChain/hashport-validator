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

package cryptotransfer

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/process"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	clients "github.com/limechain/hedera-eth-bridge-validator/app/domain/clients/hedera"
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
)

// Crypto Transfer event handler
type CryptoTransferHandler struct {
	pollingInterval    time.Duration
	topicID            hedera.TopicID
	ethSigner          *eth.Signer
	hederaMirrorClient clients.HederaMirrorClient
	hederaNodeClient   clients.HederaNodeClient
	transactionRepo    repositories.TransactionRepository
	logger             *log.Entry
	feeCalculator      *fees.FeeCalculator
	processingService  *process.ProcessingService
}

func NewCryptoTransferHandler(
	c config.CryptoTransferHandler,
	ethSigner *eth.Signer,
	hederaMirrorClient clients.HederaMirrorClient,
	hederaNodeClient clients.HederaNodeClient,
	transactionRepository repositories.TransactionRepository,
	processingService *process.ProcessingService) *CryptoTransferHandler {
	topicID, err := hedera.TopicIDFromString(c.TopicId)
	if err != nil {
		log.Fatalf("Invalid Topic ID provided: [%s]", c.TopicId)
	}

	return &CryptoTransferHandler{
		pollingInterval:    c.PollingInterval,
		topicID:            topicID,
		ethSigner:          ethSigner,
		hederaMirrorClient: hederaMirrorClient,
		hederaNodeClient:   hederaNodeClient,
		transactionRepo:    transactionRepository,
		logger:             config.GetLoggerFor("Account Transfer Handler"),
		processingService:  processingService,
	}
}

// Recover mechanism
func (cth *CryptoTransferHandler) Recover(q *queue.Queue) {

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

		if dbTransaction.Status != txRepo.StatusInitial {
			cth.logger.Infof("Previously added Transaction with TransactionID [%s] has status [%s]. Skipping further execution.", ctm.TransactionId, dbTransaction.Status)
			return
		}
	}

	encodedSignature, err := cth.processingService.ValidateAndSignTxn(&ctm)
	if err != nil {
		cth.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
	}

	topicMessageSubmissionTx, err := cth.processingService.HandleTopicSubmission(&ctm, encodedSignature)
	if err != nil {
		cth.logger.Errorf("Failed to submit topic consensus message for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		return
	}
	topicMessageSubmissionTxId := tx.FromHederaTransactionID(topicMessageSubmissionTx)

	err = cth.transactionRepo.UpdateStatusSignatureSubmitted(ctm.TransactionId, topicMessageSubmissionTxId.String(), encodedSignature)
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
				cth.logger.Debugf("Updating status to [%s] for TX ID [%s] and Topic Submission ID [%s].", txRepo.StatusSignatureProvided, transactionId, fmt.Sprintf(topicMessageSubmissionTxId))
				err := cth.transactionRepo.UpdateStatusSignatureProvided(transactionId)
				if err != nil {
					cth.logger.Errorf("Failed to update status to [%s] status for TransactionID [%s]. Error [%s].", txRepo.StatusSignatureProvided, transactionId, err)
				}
			} else {
				cth.logger.Debugf("Updating status to [%s] for TX ID [%s] and Topic Submission ID [%s].", txRepo.StatusSignatureFailed, transactionId, fmt.Sprintf(topicMessageSubmissionTxId))
				err := cth.transactionRepo.UpdateStatusSignatureFailed(transactionId)
				if err != nil {
					cth.logger.Errorf("Failed to update status to [%s] transaction with TransactionID [%s]. Error [%s].", txRepo.StatusSignatureFailed, transactionId, err)
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
