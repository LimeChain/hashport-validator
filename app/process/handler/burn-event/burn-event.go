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

package burn_event

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	bridgeThresholdAccount hedera.AccountID
	payerAccount           hedera.AccountID
	hederaNodeClient       client.HederaNode
	repository             repository.BurnEvent
	logger                 *log.Entry
}

func NewHandler(
	c config.BurnEventHandler,
	hederaNodeClient client.HederaNode,
	repository repository.BurnEvent) *Handler {
	bridgeThresholdAccount, err := hedera.AccountIDFromString(c.BridgeThresholdAccount)
	if err != nil {
		log.Fatalf("Invalid bridge threshold account: [%s]", c.BridgeThresholdAccount)
	}

	payerAccount, err := hedera.AccountIDFromString(c.PayerAccount)
	if err != nil {
		log.Fatalf("Invalid payer account: [%s]", c.PayerAccount)
	}

	return &Handler{
		bridgeThresholdAccount: bridgeThresholdAccount,
		payerAccount:           payerAccount,
		hederaNodeClient:       hederaNodeClient,
		repository:             repository,
		logger:                 config.GetLoggerFor("Scheduled Transaction Handler"),
	}
}

func (sth Handler) Handle(payload interface{}) {
	burnEvent, ok := payload.(*burn_event.BurnEvent)
	if !ok {
		sth.logger.Errorf("Error could not cast BurnEvent payload [%s]", payload)
		return
	}

	err := sth.repository.Create(burnEvent.TxHash, burnEvent.Amount, burnEvent.Recipient.String())
	if err != nil {
		sth.logger.Errorf("[%s] - Failed to create a burn event record. Error [%s].", burnEvent.TxHash, err)
		return
	}

	var transactionID *hedera.TransactionID
	var scheduleID *hedera.ScheduleID

	if burnEvent.NativeToken == "HBAR" {
		transactionID, scheduleID, err = sth.hederaNodeClient.
			SubmitScheduledHbarTransferTransaction(burnEvent.Amount, burnEvent.Recipient, sth.bridgeThresholdAccount, sth.payerAccount, burnEvent.TxHash)
	} else {
		tokenID, err := hedera.TokenIDFromString(burnEvent.NativeToken)
		if err != nil {
			sth.logger.Errorf("[%s] - failed to parse native token [%s] to TokenID. Error [%s]", burnEvent.TxHash, burnEvent.NativeToken, err)
			return
		}
		transactionID, scheduleID, err = sth.hederaNodeClient.
			SubmitScheduledTokenTransferTransaction(burnEvent.Amount, tokenID, burnEvent.Recipient, sth.bridgeThresholdAccount, sth.payerAccount, burnEvent.TxHash)
	}

	if err != nil {
		sth.logger.Errorf("[%s] - Failed to submit scheduled transaction. Error [%s].", burnEvent.TxHash, err)
		return
	}

	sth.logger.Infof("[%s] - Successfully submitted scheduled transaction [%s] for [%s] to receive [%d] tinybars.",
		burnEvent.TxHash,
		transactionID, burnEvent.Recipient, burnEvent.Amount)

	//submissionTxID := tx.FromHederaTransactionID(transactionID)

	err = sth.repository.UpdateStatusSubmitted(burnEvent.TxHash, scheduleID.String(), transactionID.String())
	if err != nil {
		sth.logger.Errorf(
			"[%s] - Failed to update submitted status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
			burnEvent.TxHash, transactionID, scheduleID, err)
		return
	}

	// TODO: query mirror node for the final status (similar to topic submission)
}
