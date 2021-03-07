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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/bridge"
	"time"

	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
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

	err = cth.bridgeService.VerifyFee(*transferMsg)
	if err != nil {
		cth.logger.Errorf("Fee validation failed for TX [%s]. Skipping further execution", transferMsg.TransactionId)
	}

	err = cth.bridgeService.ProcessTransfer(*transferMsg)
	if err != nil {
		cth.logger.Errorf("Processing of TX [%s] failed", transferMsg.TransactionId)
	}
}
