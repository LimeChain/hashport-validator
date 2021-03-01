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

package consensusmessage

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	processutils "github.com/limechain/hedera-eth-bridge-validator/app/helper/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"strings"
)

type ConsensusMessageHandler struct {
	processingService     *process.ProcessingService
	ethereumClient        *ethereum.EthereumClient
	hederaNodeClient      *hederaClient.HederaNodeClient
	bridgeContractAddress string
	messageRepository     repositories.MessageRepository
	transactionRepository repositories.TransactionRepository
	scheduler             *scheduler.Scheduler
	signer                *eth.Signer
	topicID               hedera.TopicID
	logger                *log.Entry
	bridge                *bridge.BridgeContractService
}

func NewConsensusMessageHandler(
	configuration config.ConsensusMessageHandler,
	bridgeContractAddress string,
	messageRepository repositories.MessageRepository,
	transactionRepository repositories.TransactionRepository,
	ethereumClient *ethereum.EthereumClient,
	hederaNodeClient *hederaClient.HederaNodeClient,
	scheduler *scheduler.Scheduler,
	bridge *bridge.BridgeContractService,
	signer *eth.Signer,
	processingService *process.ProcessingService,
) *ConsensusMessageHandler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &ConsensusMessageHandler{
		processingService:     processingService,
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		bridgeContractAddress: bridgeContractAddress,
		hederaNodeClient:      hederaNodeClient,
		ethereumClient:        ethereumClient,
		topicID:               topicID,
		scheduler:             scheduler,
		signer:                signer,
		logger:                config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
		bridge:                bridge,
	}
}

func (cmh ConsensusMessageHandler) Recover(queue *queue.Queue) {
	// TODO: (Suggestion) Move this whole function before start of any watchers / handlers

	skippedTransactions, err := cmh.transactionRepository.GetSkipped()
	if err != nil {
		cmh.logger.Fatalf("Failed to retrieve transactions with status [%s] - Error: [%s]", transaction.StatusSkipped, err)
	}

	for _, tx := range skippedTransactions {
		ctm := &validatorproto.CryptoTransferMessage{
			TransactionId: tx.TransactionId,
			EthAddress:    tx.EthAddress,
			Amount:        tx.Amount,
			Fee:           tx.Fee,
		}

		encodedData, err := ethhelper.EncodeData(ctm)
		if err != nil {
			cmh.logger.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
		}

		hash := crypto.Keccak256(encodedData)
		hexHash := hex.EncodeToString(hash)

		err = cmh.scheduleIfReady(tx.TransactionId, hexHash, ctm)
		if err != nil {
			cmh.logger.Fatalf("Failed to schedule execution of Transaction with ID [%s] and hash [%s] - Error: [%s]", tx.TransactionId, hexHash, err)
		}
	}
}

func (cmh ConsensusMessageHandler) Handle(payload []byte) {
	m := &validatorproto.TopicSubmissionMessage{}
	err := proto.Unmarshal(payload, m)
	if err != nil {
		log.Errorf("Error could not unmarshal payload. Error [%s].", err)
	}

	switch m.Type {
	case validatorproto.TopicSubmissionType_EthSignature:
		err = cmh.handleSignatureMessage(m)
	case validatorproto.TopicSubmissionType_EthTransaction:
		err = cmh.handleEthTxMessage(m.GetTopicEthTransactionMessage())
	default:
		err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
	}

	if err != nil {
		cmh.logger.Errorf("Error - could not handle payload: [%s]", err)
		return
	}
}

func (cmh ConsensusMessageHandler) handleEthTxMessage(m *validatorproto.TopicEthTransactionMessage) error {
	isValid, err := cmh.verifyEthTxAuthenticity(m)
	if err != nil {
		cmh.logger.Errorf("[%s] - ETH TX [%s] - Error while trying to verify TX authenticity.", m.TransactionId, m.EthTxHash)
		return err
	}

	if !isValid {
		cmh.logger.Infof("[%s] - Eth TX [%s] - Invalid authenticity.", m.TransactionId, m.EthTxHash)
		return nil
	}

	err = cmh.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId, m.EthTxHash)
	if err != nil {
		cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
		return err
	}

	go cmh.processingService.AcknowledgeTransactionSuccess(m)

	return cmh.scheduler.Cancel(m.TransactionId)
}

func (cmh ConsensusMessageHandler) verifyEthTxAuthenticity(m *validatorproto.TopicEthTransactionMessage) (bool, error) {
	tx, _, err := cmh.ethereumClient.Client.TransactionByHash(context.Background(), common.HexToHash(m.EthTxHash))
	if err != nil {
		cmh.logger.Warnf("[%s] - Failed to get eth transaction by hash [%s]. Error [%s].", m.TransactionId, m.EthTxHash, err)
		return false, err
	}

	if strings.ToLower(tx.To().String()) != strings.ToLower(cmh.bridgeContractAddress) {
		cmh.logger.Debugf("[%s] - ETH TX [%s] - Failed authenticity - Different To Address [%s].", m.TransactionId, m.EthTxHash, tx.To().String())
		return false, nil
	}

	txMessage, signatures, err := ethhelper.DecodeBridgeMintFunction(tx.Data())
	if err != nil {
		return false, err
	}

	if txMessage.TransactionId != m.TransactionId {
		cmh.logger.Debugf("[%s] - ETH TX [%s] - Different txn id [%s].", m.TransactionId, m.EthTxHash, txMessage.TransactionId)
		return false, nil
	}

	dbTx, err := cmh.transactionRepository.GetByTransactionId(m.TransactionId)
	if err != nil {
		return false, err
	}
	if dbTx == nil {
		cmh.logger.Debugf("[%s] - ETH TX [%s] - Transaction not found in database.", m.TransactionId, m.EthTxHash)
		return false, nil
	}

	if dbTx.Amount != txMessage.Amount ||
		dbTx.EthAddress != txMessage.EthAddress ||
		dbTx.Fee != txMessage.Fee {
		cmh.logger.Debugf("[%s] - ETH TX [%s] - Invalid arguments.", m.TransactionId, m.EthTxHash)
		return false, nil
	}

	encodedData, err := ethhelper.EncodeData(txMessage)
	if err != nil {
		return false, err
	}
	hash := ethhelper.KeccakData(encodedData)

	checkedAddresses := make(map[string]bool)
	for _, signature := range signatures {
		address, err := ethhelper.GetAddressBySignature(hash, signature)
		if err != nil {
			return false, err
		}
		if checkedAddresses[address] {
			return false, err
		}

		if !processutils.IsValidAddress(address, cmh.operatorsEthAddresses) {
			cmh.logger.Debugf("[%s] - ETH TX [%s] - Invalid operator process - [%s].", m.TransactionId, m.EthTxHash, address)
			return false, nil
		}
		checkedAddresses[address] = true
	}

	return true, nil
}

func (cmh ConsensusMessageHandler) acknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage) {
	cmh.logger.Infof("Waiting for Transaction with ID [%s] to be mined.", m.TransactionId)

	isSuccessful, err := cmh.ethereumClient.WaitForTransactionSuccess(common.HexToHash(m.EthTxHash))
	if err != nil {
		cmh.logger.Errorf("Failed to await TX ID [%s] with ETH TX [%s] to be mined. Error [%s].", m.TransactionId, m.Hash, err)
		return
	}

	if !isSuccessful {
		cmh.logger.Infof("Transaction with ID [%s] was reverted. Updating status to [%s].", m.TransactionId, transaction.StatusEthTxReverted)
		err = cmh.transactionRepository.UpdateStatusEthTxReverted(m.TransactionId)
		if err != nil {
			cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxReverted, m.TransactionId, err)
			return
		}
	} else {
		cmh.logger.Infof("Transaction with ID [%s] was successfully mined. Updating status to [%s].", m.TransactionId, transaction.StatusCompleted)
		err = cmh.transactionRepository.UpdateStatusCompleted(m.TransactionId)
		if err != nil {
			cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusCompleted, m.TransactionId, err)
			return
		}
	}
}

func (cmh ConsensusMessageHandler) handleSignatureMessage(msg *validatorproto.TopicSubmissionMessage) error {
	hash, message, err := cmh.processingService.ValidateAndSaveSignature(msg)
	if err != nil {
		cmh.logger.Errorf("Could not Validate and Save Signature for Transaction with ID [%s] and hash [%s] - Error: [%s]", message.TransactionId, hash, err)
		return err
	}

	return cmh.scheduleIfReady(message.TransactionId, hash, message)
}

func (cmh ConsensusMessageHandler) scheduleIfReady(txId string, hash string, message *validatorproto.CryptoTransferMessage) error {
	txMessages, err := cmh.messageRepository.GetTransactions(txId, hash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve transaction messages for Transaction ID [%s]. Error [%s]", txId, err))
	}

	if cmh.enoughSignaturesCollected(txMessages, txId) {
		cmh.logger.Debugf("TX [%s] - Enough signatures have been collected.", txId)

		slot, isFound := cmh.computeExecutionSlot(txMessages)
		if !isFound {
			cmh.logger.Debugf("TX [%s] - Operator [%s] has not been found as signer amongst the signatures collected.", txId, cmh.signer.Address())
			return nil
		}

		submission := &ethsubmission.Submission{
			CryptoTransferMessage: message,
			Messages:              txMessages,
			Slot:                  slot,
			TransactOps:           cmh.signer.NewKeyTransactor(),
		}

		err := cmh.scheduler.Schedule(txId, *submission)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmh ConsensusMessageHandler) enoughSignaturesCollected(txSignatures []message.TransactionMessage, transactionId string) bool {
	requiredSigCount := len(cmh.bridge.GetMembers())/2 + 1
	cmh.logger.Infof("Collected [%d/%d] Signatures for TX ID [%s] ", len(txSignatures), len(cmh.bridge.GetMembers()), transactionId)
	return len(txSignatures) >= requiredSigCount
}

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (cmh ConsensusMessageHandler) computeExecutionSlot(messages []message.TransactionMessage) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].SignerAddress) == strings.ToLower(cmh.signer.Address()) {
			return int64(i), true
		}
	}

	return -1, false
}