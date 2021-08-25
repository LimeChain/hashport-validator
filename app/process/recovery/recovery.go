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
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Recovery struct {
	transfers               service.Transfers
	messages                service.Messages
	contractServices        map[int64]service.Contracts
	statusTransferRepo      repository.Status
	statusMessagesRepo      repository.Status
	transferRepo            repository.Transfer
	mirrorClient            client.MirrorNode
	nodeClient              client.HederaNode
	accountID               hederasdk.AccountID
	topicID                 hederasdk.TopicID
	configRecoveryTimestamp int64
	logger                  *log.Entry
	mappings                config.AssetMappings
}

func NewProcess(
	c config.Validator,
	transfers service.Transfers,
	messages service.Messages,
	contractServices map[int64]service.Contracts,
	statusTransferRepo repository.Status,
	statusMessagesRepo repository.Status,
	transferRepo repository.Transfer,
	mirrorClient client.MirrorNode,
	nodeClient client.HederaNode,
	mappings config.AssetMappings,
) (*Recovery, error) {
	account, err := hederasdk.AccountIDFromString(c.Clients.Hedera.BridgeAccount)
	if err != nil {
		return nil, err
	}

	topic, err := hederasdk.TopicIDFromString(c.Clients.Hedera.TopicId)
	if err != nil {
		return nil, err
	}

	return &Recovery{
		transfers:               transfers,
		messages:                messages,
		contractServices:        contractServices,
		statusTransferRepo:      statusTransferRepo,
		statusMessagesRepo:      statusMessagesRepo,
		transferRepo:            transferRepo,
		mirrorClient:            mirrorClient,
		nodeClient:              nodeClient,
		accountID:               account,
		topicID:                 topic,
		configRecoveryTimestamp: c.Recovery.StartTimestamp,
		logger:                  config.GetLoggerFor(fmt.Sprintf("Recovery")),
		mappings:                mappings,
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
		amount, nativeAsset, err := tx.GetIncomingTransfer(r.accountID.String())
		if err != nil {
			r.logger.Errorf("[%s] - Skipping recovery. Invalid amount. Error: [%s]", tx.TransactionID, err)
			continue
		}

		wrappedAsset := r.mappings.NativeToWrapped(nativeAsset, 0, 1)
		if wrappedAsset == "" {
			r.logger.Errorf("[%s] - Could not parse native asset [%s] - Error: [%s]", tx.TransactionID, nativeAsset, err)
			continue
		}

		_, m, err := r.transfers.SanityCheckTransfer(tx)
		if err != nil {
			r.logger.Errorf("[%s] - Skipping recovery. Failed sanity check. Error: [%s]", tx.TransactionID, err)
			continue
		}
		err = r.transfers.SaveRecoveredTxn(tx.TransactionID, amount, nativeAsset, wrappedAsset, m)
		if err != nil {
			r.logger.Errorf("[%s] - Skipping recovery. Unable to persist TX. Error: [%s]", tx.TransactionID, err)
			continue
		}
		r.logger.Debugf("[%s] - Recovered transfer", tx.TransactionID)
	}

	r.logger.Infof("[%s] - Successfully recovered [%d] transfer TXns", r.accountID, len(txns))
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
		m, err := message.FromString(msg.Contents, msg.ConsensusTimestamp)
		if err != nil {
			r.logger.Errorf("Skipping recovery of Topic Message with timestamp [%s]. Could not decode message. Error: [%s]", msg.ConsensusTimestamp, err)
			continue
		}

		err = r.messages.ProcessSignature(*m)
		if err != nil {
			r.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}
	}

	r.logger.Infof("Successfully recovered [%d] Messages for Topic [%s]", len(messages), r.topicID)
	return nil
}

func (r Recovery) processUnfinishedOperations() error {
	unprocessedTransfers, err := r.transferRepo.GetUnprocessedTransfers()
	if err != nil {
		r.logger.Errorf("Failed to get unprocessed transfers. Error: [%s]", err)
		return err
	}

	for _, t := range unprocessedTransfers {
		// TODO: remove mockChainID and add targetChainID in t (*entity.Transfer)
		mockChainID := int64(80001)
		transferMsg := transfer.NewNative(
			t.TransactionID,
			t.Receiver,
			t.NativeAsset,
			t.WrappedAsset,
			t.Amount,
			r.contractServices[mockChainID].Address().String())

		err = r.transfers.ProcessNativeTransfer(transferMsg.Transfer)
		if err != nil {
			r.logger.Errorf("Processing of TX [%s] failed", transferMsg.TransactionId)
			continue
		}
	}
	return nil
}
