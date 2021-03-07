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

package recovery

import (
	"encoding/base64"
	"errors"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	joined "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type Recovery struct {
	bridgeService           services.Bridge
	transactionRepository   repositories.Transaction
	topicStatusRepository   repositories.Status
	accountStatusRepository repositories.Status
	mirrorClient            clients.MirrorNode
	nodeClient              clients.HederaNode
	accountID               hederasdk.AccountID
	topicID                 hederasdk.TopicID
	cryptoTransferTS        int64
	logger                  *log.Entry
}

func NewRecoveryProcess(
	bridgeService services.Bridge,
	transactionRepository repositories.Transaction,
	topicStatusRepository repositories.Status,
	accountStatusRepository repositories.Status,
	mirrorClient clients.MirrorNode,
	nodeClient clients.HederaNode,
	accountID hederasdk.AccountID,
	topicID hederasdk.TopicID,
	cryptoTS int64,
) *Recovery {
	return &Recovery{
		bridgeService:           bridgeService,
		transactionRepository:   transactionRepository,
		topicStatusRepository:   topicStatusRepository,
		accountStatusRepository: accountStatusRepository,
		mirrorClient:            mirrorClient,
		nodeClient:              nodeClient,
		accountID:               accountID,
		topicID:                 topicID,
		logger:                  config.GetLoggerFor(fmt.Sprintf("Recovery Service")),
		cryptoTransferTS:        cryptoTS,
	}
}

// Recover starts the main recovery process
func (r *Recovery) Recover(from, to int64) error {
	r.logger.Infof("Starting Recovery Process")

	err := r.transfersRecovery(from, to)
	if err != nil {
		r.logger.Errorf("Transfers Recovery failed: [%s]", err)
		return err
	}

	_, err = r.topicMessagesRecovery(from, to)
	if err != nil {
		r.logger.Errorf("Topic Messages Recovery failed", err)
		return err
	}

	// TODO Handle unprocessed TXs
	// 1. Get all Skipped TX (DONE)
	// 2. Get all message records for the set of TX IDs (from the Skipped TX records) (DONE)
	// 3. Group messages and TX IDs into a map (TX ID->Messages) (DONE)
	// 4. Go through all TX ID -> Messages. If current validator node haven't submitted a signature message -> sign and submit signature message to topic (DONE)

	log.Infof("Starting to process skipped Transactions")
	err = r.processSkipped()
	if err != nil {
		r.logger.Errorf("Error - could not finish processing skipped transactions: [%s]", err)
		return err
	}
	log.Infof("[SUCCESSFUL] Process of Skipped Transactions")

	return nil
}

// transfersRecovery queries all incoming Transfer Transactions for the specified AccountID occurring between `from` and `to`
// Performs sanity checks and persists them in the database
func (r *Recovery) transfersRecovery(from int64, to int64) error {
	txns, err := r.mirrorClient.GetAccountCreditTransactionsBetween(r.accountID, from, to)
	if err != nil {
		return err
	}

	r.logger.Infof("Found [%d] unprocessed TXns for Account [%s]", len(txns), r.accountID)
	for _, tx := range txns {
		amount, err := tx.GetIncomingAmountFor(r.accountID.String())
		if err != nil {
			r.logger.Errorf("Skipping recovery of TX [%s]. Invalid amount. Error: [%s]", tx.TransactionID, err)
			continue
		}
		m, err := r.bridgeService.SanityCheck(tx)
		if err != nil {
			r.logger.Errorf("Skipping recovery of [%s]. Failed sanity check. Error: [%s]", tx.TransactionID, err)
			continue
		}
		err = r.bridgeService.SaveRecoveredTxn(tx.TransactionID, amount, *m)
		if err != nil {
			r.logger.Errorf("Skipping recovery of [%s]. Unable to persist TX. Err: [%s]", tx.TransactionID, err)
			continue
		}
		r.logger.Debugf("Recovered transfer with TXn ID [%s]", tx.TransactionID)
	}

	r.logger.Infof("Successfully recovered [%d] transfer TXns for Account [%s]", len(txns), r.accountID)
	return nil
}

// topicMessagesRecovery
func (r *Recovery) topicMessagesRecovery(from, to int64) error {
	messages, err := r.mirrorClient.GetMessagesForTopicBetween(r.topicID, from, to)
	if err != nil {
		return err
	}

	r.logger.Infof("Found [%d] unprocessed messages for Topic [%s]", len(messages), r.topicID)
	for _, msg := range messages {
		m, err := encoding.NewTopicMessageFromString(msg.Contents, msg.ConsensusTimestamp)
		if err != nil {
			r.logger.Errorf("Skipping recovery of Topic MSG with TS [%s]. Could not decode message. Error: [%s]", msg.ConsensusTimestamp, err)
			continue
		}

		switch m.Type {
		case validatorproto.TopicMessageType_EthSignature:
			_, _, err = r.bridgeService.ValidateAndSaveSignature(m)
		case validatorproto.TopicMessageType_EthTransaction:
			err = r.checkStatusAndUpdate(m.GetTopicEthTransactionMessage())
		default:
			err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
		}

		if err != nil {
			r.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}
	}

	return nil
}

func (r *Recovery) processSkipped() error {
	unprocessed, err := r.transactionRepository.GetSkippedOrInitialTransactionsAndMessages()
	if err != nil {
		return errors.New(fmt.Sprintf("Error - could not go through all skipped transactions: [%s]", err))
	}

	for txn, txnSignatures := range unprocessed {
		hasSubmittedSignature, ctm := r.hasSubmittedSignature(txn, txnSignatures)

		if !hasSubmittedSignature {
			r.logger.Infof("Validator has not yet submitted signature for Transaction with ID [%s]. Proceeding now...", txn)

			signature, err := r.bridgeService.ValidateAndSignTxn(ctm)
			if err != nil {
				r.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", txn, err)
			}

			_, err = r.bridgeService.HandleTopicSubmission(ctm, signature)
			if err != nil {
				return errors.New(fmt.Sprintf("Could not submit Signature [%s] to Topic [%s] - Error: [%s]", signature, r.topicID, err))
			}
			r.logger.Infof("Successfully Validated")
		}
	}
	return nil
}

func (r *Recovery) hasSubmittedSignature(data joined.CTMKey, signatures []string) (bool, *validatorproto.CryptoTransferMessage) {
	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: data.TransactionId,
		EthAddress:    data.EthAddress,
		Amount:        data.Amount,
		Fee:           data.Fee,
		GasPriceGwei:  data.GasPriceGwei,
	}

	signature, err := r.bridgeService.ValidateAndSignTxn(ctm)
	if err != nil {
		r.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", data.TransactionId, err)
	}

	for _, s := range signatures {
		if signature == s {
			return true, nil
		}
	}
	return false, ctm
}

func (r *Recovery) getStartTimestampFor(repository repositories.Status, address string) int64 {
	if r.cryptoTransferTS > 0 {
		return r.cryptoTransferTS
	}

	timestamp, err := repository.GetLastFetchedTimestamp(address)
	if err == nil {
		return timestamp
	}

	return -1
}

func (r *Recovery) checkStatusAndUpdate(m *validatorproto.TopicEthTransactionMessage) error {
	err := r.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId, m.EthTxHash)
	if err != nil {
		r.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
		return err
	}

	go r.bridgeService.AcknowledgeTransactionSuccess(m)
	return nil
}
