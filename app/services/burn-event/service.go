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
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	bridgeAccount    hedera.AccountID
	payerAccount     hedera.AccountID
	repository       repository.BurnEvent
	hederaNodeClient client.HederaNode
	mirrorNodeClient client.MirrorNode
	logger           *log.Entry
}

func NewService(
	bridgeAccount string,
	payerAccount string,
	hederaNodeClient client.HederaNode,
	mirrorNodeClient client.MirrorNode,
	repository repository.BurnEvent) *Service {

	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid bridge threshold account: [%s].", bridgeAccount)
	}

	payer, err := hedera.AccountIDFromString(payerAccount)
	if err != nil {
		log.Fatalf("Invalid payer account: [%s].", payerAccount)
	}

	return &Service{
		bridgeAccount:    bridgeAcc,
		payerAccount:     payer,
		repository:       repository,
		hederaNodeClient: hederaNodeClient,
		mirrorNodeClient: mirrorNodeClient,
		logger:           config.GetLoggerFor("Burn Event Service"),
	}
}

func (s Service) ProcessEvent(event burn_event.BurnEvent) {
	err := s.repository.Create(event.Id, event.Amount, event.Recipient.String())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create a burn event record. Error [%s].", event.Id, err)
		return
	}

	var transactionResponse *hedera.TransactionResponse
	if event.NativeToken == constants.Hbar {
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledHbarTransferTransaction(event.Amount, event.Recipient, s.bridgeAccount, s.payerAccount, event.Id)
	} else {
		tokenID, err := hedera.TokenIDFromString(event.NativeToken)
		if err != nil {
			s.logger.Errorf("[%s] - failed to parse native token [%s] to TokenID. Error [%s].", event.Id, event.NativeToken, err)
			return
		}
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledTokenTransferTransaction(event.Amount, tokenID, event.Recipient, s.bridgeAccount, s.payerAccount, event.Id)
	}
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit scheduled transaction. Error [%s].", event.Id, err)
		return
	}

	s.logger.Infof("[%s] - Successfully submitted scheduled transaction [%s] for [%s] to receive [%d] [%s] .",
		event.Id,
		transactionResponse.TransactionID,
		event.Recipient,
		event.Amount,
		event.NativeToken)

	txReceipt, err := transactionResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for [%s]", event.Id, transactionResponse.TransactionID)
		return
	}

	switch txReceipt.Status {
	case hedera.StatusIdenticalScheduleAlreadyCreated:
		s.handleScheduleSign(event.Id, *txReceipt.ScheduleID)
	case hedera.StatusSuccess:
		s.logger.Infof("[%s] - Updating db status to Submitted with TransactionID [%s].",
			event.Id,
			hederahelper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String()))
	default:
		s.logger.Errorf("[%s] - TX [%s] - Scheduled Transaction resolved with [%s].", event.Id, transactionResponse.TransactionID, txReceipt.Status)

		err := s.repository.UpdateStatusFailed(transactionResponse.TransactionID.String())
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", transactionResponse.TransactionID.String(), err)
			return
		}
		return
	}

	transactionID := hederahelper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String())
	err = s.repository.UpdateStatusSubmitted(event.Id, txReceipt.ScheduleID.String(), transactionID)
	if err != nil {
		s.logger.Errorf(
			"[%s] - Failed to update submitted status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
			event.Id, transactionID, txReceipt.ScheduleID.String(), err)
		return
	}

	onSuccess, onFail := s.scheduledTxExecutionCallbacks(transactionID)
	s.mirrorNodeClient.WaitForScheduledTransferTransaction(transactionID, onSuccess, onFail)
}

func (s *Service) handleScheduleSign(id string, scheduleID hedera.ScheduleID) {
	s.logger.Debugf("[%s] - Scheduled transaction already created - Executing Scheduled Sign for [%s].", id, scheduleID)
	txResponse, err := s.hederaNodeClient.SubmitScheduleSign(scheduleID)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit schedule sign [%s]. Error: [%s].", id, scheduleID, err)
		return
	}

	receipt, err := txResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for schedule sign [%s]. Error: [%s].", id, scheduleID, err)
		return
	}

	switch receipt.Status {
	case hedera.StatusSuccess:
		s.logger.Debugf("[%s] - Successfully executed schedule sign for [%s].", id, scheduleID)
	case hedera.StatusScheduleAlreadyExecuted:
		s.logger.Debugf("[%s] - Scheduled Sign [%s] already executed.", id, scheduleID)
	default:
		s.logger.Errorf("[%s] - Schedule Sign [%s] failed with [%s].", id, scheduleID, receipt.Status)
	}
}

func (s *Service) scheduledTxExecutionCallbacks(txId string) (onSuccess, onFail func()) {
	onSuccess = func() {
		s.logger.Debugf("[%s] - Scheduled TX execution successful.", txId)
		err := s.repository.UpdateStatusCompleted(txId)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status completed. Error [%s].", txId, err)
			return
		}
	}

	onFail = func() {
		s.logger.Debugf("[%s] - Scheduled TX execution has failed.", txId)
		err := s.repository.UpdateStatusFailed(txId)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onFail
}
