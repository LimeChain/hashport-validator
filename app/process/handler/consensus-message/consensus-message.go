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
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/handler"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"strings"
)

type ConsensusMessageHandler struct {
	ethereumClient        *ethereum.EthereumClient
	hederaNodeClient      *hederaClient.HederaNodeClient
	operatorsEthAddresses []string
	messageRepository     repositories.MessageRepository
	transactionRepository repositories.TransactionRepository
	scheduler             *scheduler.Scheduler
	signer                *eth.Signer
	topicID               hedera.TopicID
	logger                *log.Entry
}

func NewConsensusMessageHandler(
	configuration config.ConsensusMessageHandler,
	messageRepository repositories.MessageRepository,
	transactionRepository repositories.TransactionRepository,
	ethereumClient *ethereum.EthereumClient,
	hederaNodeClient *hederaClient.HederaNodeClient,
	scheduler *scheduler.Scheduler,
	signer *eth.Signer,
) *ConsensusMessageHandler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &ConsensusMessageHandler{
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		operatorsEthAddresses: configuration.Addresses,
		hederaNodeClient:      hederaNodeClient,
		ethereumClient:        ethereumClient,
		topicID:               topicID,
		scheduler:             scheduler,
		signer:                signer,
		logger:                config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
	}
}

func (cmh ConsensusMessageHandler) Recover(queue *queue.Queue) {
	cmh.logger.Println("Recovery method not implemented yet.")
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
	// TODO: verify authenticity of transaction hash

	err := cmh.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId, m.EthTxHash)
	if err != nil {
		cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
		return err
	}

	go handler.AcknowledgeTransactionSuccess(m, cmh.logger, cmh.ethereumClient, cmh.transactionRepository)

	return cmh.scheduler.Cancel(m.TransactionId)
}

func (cmh ConsensusMessageHandler) handleSignatureMessage(msg *validatorproto.TopicSubmissionMessage) error {
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

	hash := crypto.Keccak256(encodedData)
	hexHash := hex.EncodeToString(hash)

	decodedSig, ethSig, err := ethhelper.DecodeSignature(m.GetSignature())
	m.Signature = ethSig
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	exists, err := handler.AlreadyExists(cmh.messageRepository, m, ethSig, hexHash)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	key, err := crypto.Ecrecover(hash, decodedSig)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to recover public key. Hash - [%s] - [%s]", m.TransactionId, hexHash, err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to unmarshal public key. - [%s]", m.TransactionId, err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if !handler.IsValidAddress(address.String(), cmh.operatorsEthAddresses) {
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

		submission := &ethsubmission.Submission{
			CryptoTransferMessage: ctm,
			Messages:              txMessages,
			Slot:                  slot,
			TransactOps:           cmh.signer.NewKeyTransactor(),
		}

		err := cmh.scheduler.Schedule(m.TransactionId, *submission)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmh ConsensusMessageHandler) enoughSignaturesCollected(txSignatures []message.TransactionMessage, transactionId string) bool {
	requiredSigCount := len(cmh.operatorsEthAddresses)/2 + 1
	cmh.logger.Infof("Collected [%d/%d] Signatures for TX ID [%s] ", len(txSignatures), len(cmh.operatorsEthAddresses), transactionId)
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
