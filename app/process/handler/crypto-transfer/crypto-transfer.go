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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/bridge"
	"time"

	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

// Crypto Transfer event handler
type Handler struct {
	pollingInterval    time.Duration
	topicID            hedera.TopicID
	ethSigner          *eth.Signer
	hederaMirrorClient clients.MirrorNode
	hederaNodeClient   clients.HederaNode
	transactionRepo    repositories.Transaction
	logger             *log.Entry
	feeCalculator      *fees.Calculator
	bridgeService      services.Bridge
}

func NewHandler(
	c config.CryptoTransferHandler,
	ethSigner *eth.Signer,
	hederaMirrorClient clients.MirrorNode,
	hederaNodeClient clients.HederaNode,
	transactionRepository repositories.Transaction,
	processingService *bridge.Service) *Handler {
	topicID, err := hedera.TopicIDFromString(c.TopicId)
	if err != nil {
		log.Fatalf("Invalid Topic ID provided: [%s]", c.TopicId)
	}

	return &Handler{
		pollingInterval:    c.PollingInterval,
		topicID:            topicID,
		ethSigner:          ethSigner,
		hederaMirrorClient: hederaMirrorClient,
		hederaNodeClient:   hederaNodeClient,
		transactionRepo:    transactionRepository,
		logger:             config.GetLoggerFor("Account Transfer Handler"),
		bridgeService:      processingService,
	}
}

// Recover mechanism
func (cth Handler) Recover(q *queue.Queue) {

}

func (cth Handler) Handle(payload []byte) {
	transferMsg, err := encoding.NewTransferMessageFromBytes(payload)
	if err != nil {
		cth.logger.Errorf("Failed to parse incoming payload. Error [%s].", err)
		return
	}

	transactionRecord, err := cth.bridgeService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		cth.logger.Errorf("Error occurred while initiating TX ID [%s] processing", transferMsg.TransactionId)
		return
	}

	if transactionRecord.Status != txRepo.StatusInitial {
		cth.logger.Infof("Previously added Transaction with TransactionID [%s] has status [%s]. Skipping further execution.", transactionRecord.TransactionId, transactionRecord.Status)
		return
	}

	// TODO
	encodedSignature, err := cth.bridgeService.ValidateAndSignTxn(*transferMsg)
	if err != nil {
		cth.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", transferMsg.TransactionId, err)
	}

	topicMessageSubmissionTx, err := cth.bridgeService.HandleTopicSubmission(transferMsg, encodedSignature)
	if err != nil {
		cth.logger.Errorf("Failed to submit topic consensus message for TransactionID [%s]. Error [%s].", transferMsg.TransactionId, err)
		return
	}
	topicMessageSubmissionTxId := tx.FromHederaTransactionID(topicMessageSubmissionTx)

	err = cth.transactionRepo.UpdateStatusSignatureSubmitted(transferMsg.TransactionId, topicMessageSubmissionTxId.String(), encodedSignature)
	if err != nil {
		cth.logger.Errorf("Failed to update submitted status for TransactionID [%s]. Error [%s].", transferMsg.TransactionId, err)
		return
	}

	go cth.checkForTransactionCompletion(transferMsg.TransactionId, topicMessageSubmissionTxId.String())
}

func (cth Handler) checkForTransactionCompletion(transactionId string, topicMessageSubmissionTxId string) {
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
