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
	processutils "github.com/limechain/hedera-eth-bridge-validator/app/helper/process"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	joined "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	consensusmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
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
func (r *Recovery) Recover() (int64, error) {
	from := r.getStartTimestampFor(r.accountStatusRepository, r.accountID.String())
	to := time.Now().UnixNano()
	if from < 0 {
		log.Info("Nothing to recover. Proceeding to start watchers and handlers.")
		return to, nil
	}

	log.Infof("Crypto Transfer Recovery for Account [%s]", r.accountID.String())
	now, err := r.transfersRecovery(from, to)
	if err != nil {
		r.logger.Errorf("Error - could not finish crypto transfer recovery process: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Crypto Transfer Recovery for Account [%s]", r.accountID.String())

	log.Infof("Consensus Message Recovery for Topic [%s]", r.topicID.String())

	now, err = r.consensusMessageRecovery(now)
	if err != nil {
		r.logger.Errorf("Error - could not finish consensus message recovery process: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Consensus Message Recovery for Topic [%s]", r.topicID.String())

	// TODO Handle unprocessed TXs
	// 1. Get all Skipped TX (DONE)
	// 2. Get all message records for the set of TX IDs (from the Skipped TX records) (DONE)
	// 3. Group messages and TX IDs into a map (TX ID->Messages) (DONE)
	// 4. Go through all TX ID -> Messages. If current validator node haven't submitted a signature message -> sign and submit signature message to topic (DONE)

	log.Infof("Starting to process skipped Transactions")
	err = r.processSkipped()
	if err != nil {
		r.logger.Errorf("Error - could not finish processing skipped transactions: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Process of Skipped Transactions")

	return now, nil
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

func (r *Recovery) transfersRecovery(from int64, to int64) (int64, error) {
	result, err := r.mirrorClient.GetSuccessfulAccountCreditTransactionsAfterDate(r.accountID, from)
	if err != nil {
		return 0, err
	}
	// TODO filter all TX after `to`
	//if recent(tx.ConsensusTimestamp, to) {
	//	break
	//}

	r.logger.Infof("Unprocessed transactions found: [%d]", len(result.Transactions))
	for _, tx := range result.Transactions {
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

		r.logger.Debugf("Adding a transaction with ID [%s] unprocessed transactions with status [%s]", tx.TransactionID, transaction.StatusSkipped)
		// TODO use bridge.persistTransfer
		err = r.transactionRepository.Skip(&proto.CryptoTransferMessage{
			TransactionId: tx.TransactionID,
			EthAddress:    m.EthereumAddress,
			Amount:        amount,
			Fee:           m.TxReimbursementFee,
			GasPriceGwei:  m.GasPriceGwei,
		})
		if err != nil {
			return 0, err
		}
		// TODO end

		// TODO at the end, update the last fetched timestamp (no need for updates on every TX)
		timestamp, err := timestampHelper.FromString(tx.ConsensusTimestamp)
		if err != nil {
			return 0, err
		}

		err = r.accountStatusRepository.UpdateLastFetchedTimestamp(r.accountID.String(), timestamp)
		if err != nil {
			return 0, err
		}
	}

	return to, nil
}

func (r *Recovery) consensusMessageRecovery(now int64) (int64, error) {
	result, err := r.mirrorClient.GetHederaTopicMessagesAfterTimestamp(r.topicID, r.getStartTimestampFor(r.topicStatusRepository, r.topicID.String()))
	if err != nil {
		r.logger.Errorf("Error - could not retrieve messages for recovery: [%s]", err)
		return 0, err
	}
	// TODO filter all TX after `to`

	r.logger.Infof("Unprocessed topic messages: [%d]", len(result.Messages))
	for _, msg := range result.Messages {
		if recent(msg.ConsensusTimestamp, now) {
			break
		}

		timestamp, err := timestampHelper.FromString(msg.ConsensusTimestamp)
		if err != nil {
			r.logger.Errorf("Error - could not parse timestamp string to int64: [%s]", err)
			continue
		}

		contents, err := base64.StdEncoding.DecodeString(msg.Contents)
		if err != nil {
			r.logger.Errorf("Error - could not decode contents of topic message: [%s]", err)
			continue
		}

		m, err := consensusmessage.PrepareMessage(contents, timestamp)
		if err != nil {
			r.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}

		switch m.Type {
		case validatorproto.TopicSubmissionType_EthSignature:
			_, _, err = r.bridgeService.ValidateAndSaveSignature(m)
		case validatorproto.TopicSubmissionType_EthTransaction:
			err = r.checkStatusAndUpdate(m.GetTopicEthTransactionMessage())
		default:
			err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
		}

		if err != nil {
			r.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}
	}

	return now, nil
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

func recent(timestamp string, now int64) bool {
	consensusTimestampParams := strings.Split(timestamp, ".")
	microseconds, _ := strconv.ParseInt(consensusTimestampParams[0], 10, 64)
	nanoseconds, _ := strconv.ParseInt(consensusTimestampParams[1], 10, 64)
	ct := microseconds*1000 + nanoseconds
	if ct > now {
		return true
	}
	return false
}
