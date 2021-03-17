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
	"errors"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	joined "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Recovery struct {
	transfers               service.Transfers
	messages                service.Messages
	statusTransferRepo      repository.Status
	mirrorClient            client.MirrorNode
	nodeClient              client.HederaNode
	accountID               hederasdk.AccountID
	topicID                 hederasdk.TopicID
	configRecoveryTimestamp int64
	logger                  *log.Entry
}

func NewProcess(
	c config.Hedera,
	transfers service.Transfers,
	messagesService service.Messages,
	statusTransferRepo repository.Status,
	mirrorClient client.MirrorNode,
	nodeClient client.HederaNode,
) (*Recovery, error) {
	account, err := hederasdk.AccountIDFromString(c.Watcher.CryptoTransfer.Account.Id)
	if err != nil {
		return nil, err
	}

	topic, err := hederasdk.TopicIDFromString(c.Watcher.ConsensusMessage.Topic.Id)
	if err != nil {
		return nil, err
	}

	return &Recovery{
		transfers:               transfers,
		messages:                messagesService,
		statusTransferRepo:      statusTransferRepo,
		mirrorClient:            mirrorClient,
		nodeClient:              nodeClient,
		accountID:               account,
		topicID:                 topic,
		configRecoveryTimestamp: c.Recovery.Timestamp,
		logger:                  config.GetLoggerFor(fmt.Sprintf("Recovery")),
	}, nil
}

// ComputeInterval calculates the `from` and `to` unix nano timestamps to be used for the recovery process
func (r *Recovery) ComputeInterval() (int64, int64, error) {
	var from int64 = 0
	to := time.Now().UnixNano()
	if r.configRecoveryTimestamp > 0 {
		from = r.configRecoveryTimestamp
	} else {
		lastFetched, err := r.statusTransferRepo.GetLastFetchedTimestamp(r.accountID.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, to, nil
			} else {
				return 0, to, err
			}
		}
		from = lastFetched
	}
	return from, to, nil
}

// Start starts the main recovery process
func (r *Recovery) Start(from, to int64) error {
	r.logger.Infof("Starting Recovery Process for interval [%d; %d]", from, to)

	err := r.transfersRecovery(from, to)
	if err != nil {
		r.logger.Errorf("Transfers Recovery failed: [%s]", err)
		return err
	}

	err = r.topicMessagesRecovery(from, to)
	if err != nil {
		r.logger.Errorf("Topic Messages Recovery failed", err)
		return err
	}

	// TODO Handle unprocessed TXs
	// 1. Get all Skipped TX (DONE)
	// 2. Get all message records for the set of TX IDs (from the Skipped TX records) (DONE)
	// 3. Group messages and TX IDs into a map (TX ID->Messages) (DONE)
	// 4. Go through all TX ID -> Messages. If current validator node haven't submitted a signature message -> sign and submit signature message to topic (DONE)

	//log.Infof("Starting to process skipped Transactions")
	//err = r.processSkipped()
	//if err != nil {
	//	r.logger.Errorf("Error - could not finish processing skipped transactions: [%s]", err)
	//	return err
	//}
	//log.Infof("[SUCCESSFUL] Process of Skipped Transactions")

	return nil
}

// transfersRecovery queries all incoming Transfer Transactions for the specified AccountID occurring between `from` and `to`
// Performs sanity checks and persists them in the database
func (r *Recovery) transfersRecovery(from int64, to int64) error {
	txns, err := r.mirrorClient.GetAccountCreditTransactionsBetween(r.accountID, from, to)
	if err != nil {
		return err
	}

	if len(txns) == 0 {
		r.logger.Infof("No Transfers found to recover for Account [%s]", r.accountID)
		return nil
	}

	r.logger.Infof("Found [%d] unprocessed TXns for Account [%s]", len(txns), r.accountID)
	for _, tx := range txns {
		amount, asset, err := tx.GetIncomingTransfer(r.accountID.String())
		if err != nil {
			r.logger.Errorf("Skipping recovery of TX [%s]. Invalid amount. Error: [%s]", tx.TransactionID, err)
			continue
		}
		m, err := r.transfers.SanityCheckTransfer(tx)
		if err != nil {
			r.logger.Errorf("Skipping recovery of [%s]. Failed sanity check. Error: [%s]", tx.TransactionID, err)
			continue
		}
		err = r.transfers.SaveRecoveredTxn(tx.TransactionID, amount, asset, *m)
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

	if len(messages) == 0 {
		r.logger.Infof("No Messages found to recover for Topic [%s]", r.topicID)
		return nil
	}

	r.logger.Debugf("Found [%d] unprocessed messages for Topic [%s]", len(messages), r.topicID)
	for _, msg := range messages {
		m, err := encoding.NewTopicMessageFromString(msg.Contents, msg.ConsensusTimestamp)
		if err != nil {
			r.logger.Errorf("Skipping recovery of Topic MSG with TS [%s]. Could not decode message. Error: [%s]", msg.ConsensusTimestamp, err)
			continue
		}

		switch m.Type {
		case validatorproto.TopicMessageType_EthSignature:
			err = r.messages.ProcessSignature(*m)
		case validatorproto.TopicMessageType_EthTransaction:
			// TODO resolve the recovery
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
	//unprocessed, err := r.transactionRepository.GetSkippedOrInitialTransactionsAndMessages()
	//if err != nil {
	//	return errors.New(fmt.Sprintf("Error - could not go through all skipped transactions: [%s]", err))
	//}
	//
	//for txn, txnSignatures := range unprocessed {
	//	hasSubmittedSignature, ctm := r.hasSubmittedSignature(txn, txnSignatures)
	//
	//	if !hasSubmittedSignature {
	//		r.logger.Infof("Validator has not yet submitted signature for Transaction with ID [%s]. Proceeding now...", txn)
	//		// TODO
	//		err = r.transfersService.VerifyFee(ctm)
	//		if err != nil {
	//			r.logger.Errorf("Fee validation failed for TX [%s]. Skipping further execution", transferMsg.TransactionId)
	//		}
	//
	//		signature, err := r.transfersService.ValidateAndSignTxn(ctm)
	//		if err != nil {
	//			r.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", txn, err)
	//		}
	//
	//		_, err = r.transfersService.HandleTopicSubmission(ctm, signature)
	//		if err != nil {
	//			return errors.New(fmt.Sprintf("Could not submit Signature [%s] to Topic [%s] - Error: [%s]", signature, r.topicID, err))
	//		}
	//		r.logger.Infof("Successfully Validated")
	//	}
	//}
	return nil
}

func (r *Recovery) hasSubmittedSignature(data joined.CTMKey, signatures []string) (bool, *validatorproto.TransferMessage) {
	//ctm := &validatorproto.TransferMessage{
	//	TransactionId: data.TransactionId,
	//	EthAddress:    data.EthAddress,
	//	Amount:        data.Amount,
	//	Fee:           data.Fee,
	//	GasPriceGwei:  data.GasPriceGwei,
	//}
	//
	//signature, err := r.transfersService.ValidateAndSignTxn(ctm)
	//if err != nil {
	//	r.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", data.TransactionId, err)
	//}
	//
	//for _, s := range signatures {
	//	if signature == s {
	//		return true, nil
	//	}
	//}
	//return false, ctm
	return false, nil
}

func (r *Recovery) checkStatusAndUpdate(m *validatorproto.TopicEthTransactionMessage) error {
	//err := r.transactionRepository.UpdateEthTxSubmitted(m.TransactionId, m.EthTxHash)
	//if err != nil {
	//	r.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
	//	return err
	//}
	//
	//go r.transfersService.AcknowledgeTransactionSuccess(m)
	return nil
}
