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
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
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
	statusMessagesRepo      repository.Status
	transactionsRepo        repository.Transaction
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
	statusMessagesRepo repository.Status,
	transactionsRepo repository.Transaction,
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
		statusMessagesRepo:      statusMessagesRepo,
		transactionsRepo:        transactionsRepo,
		mirrorClient:            mirrorClient,
		nodeClient:              nodeClient,
		accountID:               account,
		topicID:                 topic,
		configRecoveryTimestamp: c.Recovery.Timestamp,
		logger:                  config.GetLoggerFor(fmt.Sprintf("Recovery")),
	}, nil
}

// ComputeIntervals calculates the `from` and `to` unix nano timestamps to be used for the recovery process for both the transfers and messages recovery
func (r Recovery) ComputeIntervals() (transfersFrom int64, messagesFrom int64, to int64, err error) {
	to = time.Now().UnixNano()
	if r.configRecoveryTimestamp > 0 {
		transfersFrom = r.configRecoveryTimestamp
		messagesFrom = r.configRecoveryTimestamp
	} else {
		lastFetchedTransfer, err := r.statusTransferRepo.GetLastFetchedTimestamp(r.accountID.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				transfersFrom = 0
			} else {
				return 0, 0, to, err
			}
		}
		lastFetchedMessage, err := r.statusMessagesRepo.GetLastFetchedTimestamp(r.topicID.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				messagesFrom = 0
			} else {
				return 0, 0, to, err
			}
		}
		transfersFrom = lastFetchedTransfer
		messagesFrom = lastFetchedMessage
	}
	return transfersFrom, messagesFrom, to, nil
}

// Start starts the main recovery process
func (r Recovery) Start(transfersFrom, messagesFrom, to int64) error {
	r.logger.Infof("Starting Recovery Process for Transfers with interval [%s; %s]", timestamp.ToHumanReadable(transfersFrom), timestamp.ToHumanReadable(to))
	r.logger.Infof("Starting Recovery Process for Messages with interval [%s; %s]", timestamp.ToHumanReadable(messagesFrom), timestamp.ToHumanReadable(to))

	err := r.transfersRecovery(transfersFrom, to)
	if err != nil {
		r.logger.Errorf("Transfers Recovery failed: [%s]", err)
		return err
	}

	err = r.topicMessagesRecovery(messagesFrom, to)
	if err != nil {
		r.logger.Errorf("Topic Messages Recovery failed: [%s]", err)
		return err
	}

	err = r.processUnfinishedOperations()
	if err != nil {
		r.logger.Error("Failed to process unfinished operations")
		return err
	}

	log.Infof("Recovery process finished successfully")
	return nil
}

// transfersRecovery queries all incoming Transfer Transactions for the specified AccountID occurring between `from` and `to`
// Performs sanity checks and persists them in the database
func (r Recovery) transfersRecovery(from int64, to int64) error {
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
		amount, err := tx.GetIncomingAmountFor(r.accountID.String())
		if err != nil {
			r.logger.Errorf("Skipping recovery of TX [%s]. Invalid amount. Error: [%s]", tx.TransactionID, err)
			continue
		}
		m, err := r.transfers.SanityCheckTransfer(tx)
		if err != nil {
			r.logger.Errorf("Skipping recovery of [%s]. Failed sanity check. Error: [%s]", tx.TransactionID, err)
			continue
		}
		err = r.transfers.SaveRecoveredTxn(tx.TransactionID, amount, *m)
		if err != nil {
			r.logger.Errorf("Skipping recovery of [%s]. Unable to persist TX. Err: [%s]", tx.TransactionID, err)
			continue
		}
		r.logger.Debugf("Recovered transfer with TXn ID [%s]", tx.TransactionID)
	}

	r.logger.Infof("Successfully recovered [%d] transfer TXns for Account [%s]", len(txns), r.accountID)
	return nil
}

// topicMessagesRecovery queries all missed Topic messages between the provided timestamps
// Performs sanity checks on the missed messages and persists them in the DB
func (r Recovery) topicMessagesRecovery(from, to int64) error {
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
			err = r.recoverEthereumTXMessage(*m)
		default:
			err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
		}

		if err != nil {
			r.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}
	}

	r.logger.Infof("Successfully recovered [%d] Messages for TOpic [%s]", len(messages), r.topicID)
	return nil
}

func (r Recovery) recoverEthereumTXMessage(tm encoding.TopicMessage) error {
	ethTxMessage := tm.GetTopicEthTransactionMessage()
	isValid, err := r.messages.VerifyEthereumTxAuthenticity(tm)
	if err != nil {
		r.logger.Errorf("Failed to verify Ethereum TX [%s] authenticity for TX [%s]", ethTxMessage.EthTxHash, ethTxMessage.TransactionId)
		return err
	}
	if !isValid {
		r.logger.Infof("Provided Ethereum TX [%s] is not the required Mint Transaction", ethTxMessage.EthTxHash)
		return nil
	}

	err = r.messages.ProcessEthereumTxMessage(tm)
	if err != nil {
		r.logger.Errorf("Failed to process Ethereum TX Message for TX[%s]", ethTxMessage.TransactionId)
		return nil
	}
	return nil
}

func (r Recovery) processUnfinishedOperations() error {
	// TODO messagesRepo.getUnprocessedMessages() <- should return all Messages whose TX ID is status INITIAL or RECOVERED
	//unprocessed, err := r.transactionsRepo.GetSkippedOrInitialTransactionsAndMessages()
	//if err != nil {
	//	r.logger.Fatalf("Failed to get all unprocessed messages. Error: %s", err)
	//	return err
	//}

	// Combine all Messages records into map({txId, amount, fee, gasPrice, etc}, []signatures)
	// Iterate all keys
	// 	If validator has not signed -> sign and submit
	//for txn, txnSignatures := range unprocessed {
	//hasSubmittedSignature, ctm := r.hasSubmittedSignature(txn, txnSignatures)
	//
	//if !hasSubmittedSignature {
	//	r.logger.Infof("Validator has not yet submitted signature for Transaction with ID [%s]. Proceeding now...", txn)
	//	// TODO
	//	err = r.transfersService.VerifyFee(ctm)
	//	if err != nil {
	//		r.logger.Errorf("Fee validation failed for TX [%s]. Skipping further execution", transferMsg.TransactionId)
	//	}
	//
	//	signature, err := r.transfersService.ValidateAndSignTxn(ctm)
	//	if err != nil {
	//		r.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", txn, err)
	//	}
	//
	//	_, err = r.transfersService.HandleTopicSubmission(ctm, signature)
	//	if err != nil {
	//		return errors.New(fmt.Sprintf("Could not submit Signature [%s] to Topic [%s] - Error: [%s]", signature, r.topicID, err))
	//	}
	//	r.logger.Infof("Successfully Validated")
	//}
	//}
	return nil
}

func (r Recovery) hasSubmittedSignature(data joined.CTMKey, signatures []string) (bool, *validatorproto.TransferMessage) {
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
	return false, nil
}
