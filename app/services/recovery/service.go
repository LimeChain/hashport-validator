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
	processutils "github.com/limechain/hedera-eth-bridge-validator/app/helper/process"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	joined "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	consensusmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/process"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type RecoveryService struct {
	transactionRepository   repositories.Transaction
	topicStatusRepository   repositories.Status
	accountStatusRepository repositories.Status
	mirrorClient            clients.MirrorNode
	nodeClient              clients.HederaNode
	accountID               hederasdk.AccountID
	topicID                 hederasdk.TopicID
	cryptoTransferTS        int64
	logger                  *log.Entry
	processingService       *process.ProcessingService
}

func NewRecoveryService(
	processingService *process.ProcessingService,
	transactionRepository repositories.Transaction,
	topicStatusRepository repositories.Status,
	accountStatusRepository repositories.Status,
	mirrorClient clients.MirrorNode,
	nodeClient clients.HederaNode,
	accountID hederasdk.AccountID,
	topicID hederasdk.TopicID,
	cryptoTS int64,
) *RecoveryService {
	return &RecoveryService{
		processingService:       processingService,
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

func (rs *RecoveryService) Recover() (int64, error) {
	from := rs.getStartTimestampFor(rs.accountStatusRepository, rs.accountID.String())
	fmt.Println(from)
	to := time.Now().UnixNano()
	if from < 0 {
		log.Info("Nothing to recover. Proceeding to start watchers and handlers.")
		return to, nil
	}

	log.Infof("Crypto Transfer Recovery for Account [%s]", rs.accountID.String())
	now, err := rs.cryptoTransferRecovery(from, to)
	if err != nil {
		rs.logger.Errorf("Error - could not finish crypto transfer recovery process: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Crypto Transfer Recovery for Account [%s]", rs.accountID.String())

	log.Infof("Consensus Message Recovery for Topic [%s]", rs.topicID.String())

	now, err = rs.consensusMessageRecovery(now)
	if err != nil {
		rs.logger.Errorf("Error - could not finish consensus message recovery process: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Consensus Message Recovery for Topic [%s]", rs.topicID.String())

	// TODO Handle unprocessed TXs
	// 1. Get all Skipped TX (DONE)
	// 2. Get all message records for the set of TX IDs (from the Skipped TX records) (DONE)
	// 3. Group messages and TX IDs into a map (TX ID->Messages) (DONE)
	// 4. Go through all TX ID -> Messages. If current validator node haven't submitted a signature message -> sign and submit signature message to topic (DONE)

	log.Infof("Starting to process skipped Transactions")
	err = rs.processSkipped()
	if err != nil {
		rs.logger.Errorf("Error - could not finish processing skipped transactions: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Process of Skipped Transactions")

	return now, nil
}

func (rs *RecoveryService) processSkipped() error {
	unprocessed, err := rs.transactionRepository.GetSkippedOrInitialTransactionsAndMessages()
	if err != nil {
		return errors.New(fmt.Sprintf("Error - could not go through all skipped transactions: [%s]", err))
	}

	for txn, txnSignatures := range unprocessed {
		hasSubmittedSignature, ctm := rs.hasSubmittedSignature(txn, txnSignatures)

		if !hasSubmittedSignature {
			rs.logger.Infof("Validator has not yet submitted signature for Transaction with ID [%s]. Proceeding now...", txn)

			signature, err := rs.processingService.ValidateAndSignTxn(ctm)
			if err != nil {
				rs.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", txn, err)
			}

			_, err = rs.processingService.HandleTopicSubmission(ctm, signature)
			if err != nil {
				return errors.New(fmt.Sprintf("Could not submit Signature [%s] to Topic [%s] - Error: [%s]", signature, rs.topicID, err))
			}
			rs.logger.Infof("Successfully Validated")
		}
	}
	return nil
}

func (rs *RecoveryService) hasSubmittedSignature(data joined.CTMKey, signatures []string) (bool, *validatorproto.CryptoTransferMessage) {
	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: data.TransactionId,
		EthAddress:    data.EthAddress,
		Amount:        data.Amount,
		Fee:           data.Fee,
		GasPriceGwei:  data.GasPriceGwei,
	}

	signature, err := rs.processingService.ValidateAndSignTxn(ctm)
	if err != nil {
		rs.logger.Errorf("Failed to Validate and Sign TransactionID [%s]. Error [%s].", data.TransactionId, err)
	}

	for _, s := range signatures {
		if signature == s {
			return true, nil
		}
	}
	return false, ctm
}

func (rs *RecoveryService) cryptoTransferRecovery(from int64, to int64) (int64, error) {
	result, err := rs.mirrorClient.GetSuccessfulAccountCreditTransactionsAfterDate(rs.accountID, from)
	if err != nil {
		return 0, err
	}

	rs.logger.Infof("Unprocessed transactions found: [%d]", len(result.Transactions))
	for _, tr := range result.Transactions {
		if recent(tr.ConsensusTimestamp, to) {
			break
		}

		memoInfo, err := processutils.DecodeMemo(tr.MemoBase64)
		if err != nil {
			rs.logger.Errorf("Could not decode memo for Transaction with ID [%s] - Error: [%s]", tr.TransactionID, err)
			continue
		}

		rs.logger.Debugf("Adding a transaction with ID [%s] unprocessed transactions with status [%s]", tr.TransactionID, transaction.StatusSkipped)

		err = rs.transactionRepository.Skip(&proto.CryptoTransferMessage{
			TransactionId: tr.TransactionID,
			EthAddress:    memoInfo.EthAddress,
			Amount:        strconv.Itoa(int(processutils.ExtractAmount(tr, rs.accountID))),
			Fee:           memoInfo.Fee,
			GasPriceGwei:  memoInfo.GasPriceGwei,
		})

		if err != nil {
			return 0, err
		}

		timestamp, err := timestampHelper.FromString(tr.ConsensusTimestamp)
		if err != nil {
			return 0, err
		}

		err = rs.accountStatusRepository.UpdateLastFetchedTimestamp(rs.accountID.String(), timestamp)
		if err != nil {
			return 0, err
		}
	}

	return to, nil
}

func (rs *RecoveryService) consensusMessageRecovery(now int64) (int64, error) {
	result, err := rs.mirrorClient.GetHederaTopicMessagesAfterTimestamp(rs.topicID, rs.getStartTimestampFor(rs.topicStatusRepository, rs.topicID.String()))
	if err != nil {
		rs.logger.Errorf("Error - could not retrieve messages for recovery: [%s]", err)
		return 0, err
	}

	rs.logger.Infof("Unprocessed topic messages: [%d]", len(result.Messages))
	for _, msg := range result.Messages {
		if recent(msg.ConsensusTimestamp, now) {
			break
		}

		timestamp, err := timestampHelper.FromString(msg.ConsensusTimestamp)
		if err != nil {
			rs.logger.Errorf("Error - could not parse timestamp string to int64: [%s]", err)
			continue
		}

		contents, err := base64.StdEncoding.DecodeString(msg.Contents)
		if err != nil {
			rs.logger.Errorf("Error - could not decode contents of topic message: [%s]", err)
			continue
		}

		m, err := consensusmessage.PrepareMessage(contents, timestamp)
		if err != nil {
			rs.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}

		switch m.Type {
		case validatorproto.TopicSubmissionType_EthSignature:
			_, _, err = rs.processingService.ValidateAndSaveSignature(m)
		case validatorproto.TopicSubmissionType_EthTransaction:
			err = rs.checkStatusAndUpdate(m.GetTopicEthTransactionMessage())
		default:
			err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
		}

		if err != nil {
			rs.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
			continue
		}
	}

	return now, nil
}

func (rs *RecoveryService) getStartTimestampFor(repository repositories.Status, address string) int64 {
	if rs.cryptoTransferTS > 0 {
		return rs.cryptoTransferTS
	}

	timestamp, err := repository.GetLastFetchedTimestamp(address)
	if err == nil {
		return timestamp
	}

	return -1
}

func (rs *RecoveryService) checkStatusAndUpdate(m *validatorproto.TopicEthTransactionMessage) error {
	err := rs.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId, m.EthTxHash)
	if err != nil {
		rs.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
		return err
	}

	go rs.processingService.AcknowledgeTransactionSuccess(m)
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
