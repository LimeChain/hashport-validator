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

package scheduledtx

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go/v2"
	clients "github.com/limechain/hedera-eth-bridge-validator/app/domain/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type ScheduledTransactionHandler struct {
	bridgeThresholdAccount hedera.AccountID
	payerAccount           hedera.AccountID
	hederaNodeClient       clients.HederaNodeClient
	scheduledRepository    repositories.ScheduledRepository
	logger                 *log.Entry
}

func NewScheduledMessageHandler(
	c config.ScheduledTransactionHandler,
	hederaNodeClient clients.HederaNodeClient,
	scheduledRepository repositories.ScheduledRepository) *ScheduledTransactionHandler {
	bridgeThresholdAccount, err := hedera.AccountIDFromString(c.BridgeThresholdAccount)
	if err != nil {
		log.Fatalf("Invalid bridge threshold account: [%s]", c.BridgeThresholdAccount)
	}

	payerAccount, err := hedera.AccountIDFromString(c.PayerAccount)
	if err != nil {
		log.Fatalf("Invalid payer account: [%s]", c.PayerAccount)
	}

	return &ScheduledTransactionHandler{
		bridgeThresholdAccount: bridgeThresholdAccount,
		payerAccount:           payerAccount,
		hederaNodeClient:       hederaNodeClient,
		scheduledRepository:    scheduledRepository,
		logger:                 config.GetLoggerFor("Scheduled Transaction Handler"),
	}
}

// Recover mechanism
func (sth *ScheduledTransactionHandler) Recover(q *queue.Queue) {
	// todo:
}

func (sth *ScheduledTransactionHandler) Handle(payload []byte) {
	var stm protomsg.ScheduledTransactionMessage
	err := proto.Unmarshal(payload, &stm)
	if err != nil {
		sth.logger.Errorf("Failed to parse incoming payload. Error [%s].", err)
		return
	}

	recipient, err := hedera.AccountIDFromString(stm.Recipient)
	if err != nil {
		sth.logger.Errorf("[%s] - Failed to parse receiver account [%s]. Error [%s].", stm.Nonce, stm.Recipient, err)
		return
	}

	err = sth.scheduledRepository.Create(stm.Amount, stm.Nonce, stm.Recipient, sth.bridgeThresholdAccount.String(), sth.payerAccount.String())
	if err != nil {
		sth.logger.Errorf("Failed to create a scheduled record for [%s]. Error [%s]", stm.Nonce, err)
		return
	}

	transactionID, scheduleID, err := sth.hederaNodeClient.SubmitScheduledTransaction(stm.Amount, recipient, sth.bridgeThresholdAccount, sth.payerAccount, stm.Nonce)
	if err != nil {
		sth.logger.Errorf("Failed to submit scheduled transaction. Error [%s]", err)
		return
	}

	sth.logger.Infof("[%s] - Successfully submitted scheduled transaction for [%s] to receive [%d] tinybars.",
		transactionID.String(), stm.Recipient, stm.Amount)

	submissionTxID := tx.FromHederaTransactionID(transactionID)

	err = sth.scheduledRepository.UpdateStatusSubmitted(stm.Nonce, scheduleID.String(), submissionTxID.String())
	if err != nil {
		sth.logger.Errorf(
			"Failed to update submitted status of scheduled record [%s] with TransactionID [%s]. Error [%s].",
			stm.Nonce, transactionID, err)
		return
	}

	// TODO: query mirror node for the final status (similar to topic submission)
}
