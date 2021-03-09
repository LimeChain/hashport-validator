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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct {
	ethereumClient        clients.Ethereum
	hederaNodeClient      clients.HederaNode
	messageRepository     repositories.Message
	transactionRepository repositories.Transaction
	scheduler             *scheduler.Scheduler
	signer                *eth.Signer
	topicID               hedera.TopicID
	logger                *log.Entry
	bridge                *bridge.ContractService
}

func NewHandler(
	configuration config.ConsensusMessageHandler,
	messageRepository repositories.Message,
	transactionRepository repositories.Transaction,
	ethereumClient clients.Ethereum,
	hederaNodeClient clients.HederaNode,
	scheduler *scheduler.Scheduler,
	bridge *bridge.ContractService,
	signer *eth.Signer,
) *Handler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &Handler{
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		hederaNodeClient:      hederaNodeClient,
		ethereumClient:        ethereumClient,
		topicID:               topicID,
		scheduler:             scheduler,
		signer:                signer,
		logger:                config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
		bridge:                bridge,
	}
}

func (cmh Handler) Recover(queue *queue.Queue) {
	cmh.logger.Println("Recovery method not implemented yet.")
}

func (cmh Handler) Handle(payload []byte) {
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

func (cmh Handler) handleEthTxMessage(m *validatorproto.TopicEthTransactionMessage) error {
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

	go cmh.acknowledgeTransactionSuccess(m)

	return cmh.scheduler.Cancel(m.TransactionId)
}

func (cmh Handler) verifyEthTxAuthenticity(m *validatorproto.TopicEthTransactionMessage) (bool, error) {
	tx, _, err := cmh.ethereumClient.GetClient().TransactionByHash(context.Background(), common.HexToHash(m.EthTxHash))
	if err != nil {
		cmh.logger.Warnf("[%s] - Failed to get eth transaction by hash [%s]. Error [%s].", m.TransactionId, m.EthTxHash, err)
		return false, err
	}

	if strings.ToLower(tx.To().String()) != strings.ToLower(cmh.bridge.GetContractAddress().String()) {
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

		if !cmh.isValidAddress(address) {
			cmh.logger.Debugf("[%s] - ETH TX [%s] - Invalid operator address - [%s].", m.TransactionId, m.EthTxHash, address)
			return false, nil
		}
		checkedAddresses[address] = true
	}

	return true, nil
}

func (cmh Handler) acknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage) {
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

func (cmh Handler) handleSignatureMessage(msg *validatorproto.TopicSubmissionMessage) error {
	m := msg.GetTopicSignatureMessage()
	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
	}

	cmh.logger.Debugf("Signature for TX ID [%s] was received", m.TransactionId)

	encodedData, err := ethhelper.EncodeData(ctm)
	if err != nil {
		cmh.logger.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
	}

	ethHash := ethhelper.KeccakData(encodedData)

	hexHash := hex.EncodeToString(ethHash)

	decodedSig, ethSig, err := ethhelper.DecodeSignature(m.GetSignature())
	m.Signature = ethSig
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	exists, err := cmh.alreadyExists(m, ethSig, hexHash)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	key, err := crypto.Ecrecover(ethHash, decodedSig)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to recover public key. Hash - [%s] - [%s]", m.TransactionId, hexHash, err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to unmarshal public key. - [%s]", m.TransactionId, err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if !cmh.isValidAddress(address.String()) {
		return errors.New(fmt.Sprintf("[%s] - Address is not valid - [%s]", m.TransactionId, address.String()))
	}

	err = cmh.messageRepository.Create(&message.TransactionMessage{
		TransactionId:        m.TransactionId,
		EthAddress:           m.EthAddress,
		Amount:               m.Amount,
		Fee:                  m.Fee,
		Signature:            ethSig,
		Hash:                 hexHash,
		SignerAddress:        address.String(),
		TransactionTimestamp: msg.TransactionTimestamp,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s] - [%s]", m.TransactionId, ethSig, err))
	}

	cmh.logger.Debugf("Verified and saved signature for TX ID [%s]", m.TransactionId)

	txMessages, err := cmh.messageRepository.GetTransactions(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve transaction messages for Transaction ID [%s]. Error [%s]", m.TransactionId, err))
	}

	if cmh.enoughSignaturesCollected(txMessages, m.TransactionId) {
		cmh.logger.Debugf("TX [%s] - Enough signatures have been collected.", m.TransactionId)

		slot, isFound := cmh.computeExecutionSlot(txMessages)
		if !isFound {
			cmh.logger.Debugf("TX [%s] - Operator [%s] has not been found as signer amongst the signatures collected.", m.TransactionId, cmh.signer.Address())
			return nil
		}

		keyTransactor, err := cmh.signer.NewKeyTransactor(cmh.ethereumClient.ChainID())
		if err != nil {
			cmh.logger.Errorf("Failed to establish key transactor. Error [%s].", err)
		}

		submission := &ethsubmission.Submission{
			CryptoTransferMessage: ctm,
			Messages:              txMessages,
			Slot:                  slot,
			TransactOps:           keyTransactor,
		}

		err = cmh.scheduler.Schedule(m.TransactionId, *submission)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmh Handler) alreadyExists(m *validatorproto.TopicEthSignatureMessage, ethSig, hexHash string) (bool, error) {
	_, err := cmh.messageRepository.GetTransaction(m.TransactionId, ethSig, hexHash)
	notFound := errors.Is(err, gorm.ErrRecordNotFound)

	if err != nil && !notFound {
		return false, errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}
	return !notFound, nil
}

func (cmh Handler) enoughSignaturesCollected(txSignatures []message.TransactionMessage, transactionId string) bool {
	requiredSigCount := len(cmh.bridge.GetMembers())/2 + 1
	cmh.logger.Infof("Collected [%d/%d] Signatures for TX ID [%s] ", len(txSignatures), len(cmh.bridge.GetMembers()), transactionId)
	return len(txSignatures) >= requiredSigCount
}

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (cmh Handler) computeExecutionSlot(messages []message.TransactionMessage) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].SignerAddress) == strings.ToLower(cmh.signer.Address()) {
			return int64(i), true
		}
	}

	return -1, false
}

func (cmh Handler) isValidAddress(key string) bool {
	for _, k := range cmh.bridge.GetMembers() {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}
