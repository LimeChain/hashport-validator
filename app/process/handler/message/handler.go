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

package message

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	transferRepository repository.Transfer
	messageRepository  repository.Message
	contracts          service.Contracts
	messages           service.Messages
	logger             *log.Entry
}

func NewHandler(
	configuration config.ConsensusMessageHandler,
	transferRepository repository.Transfer,
	messageRepository repository.Message,
	contractsService service.Contracts,
	messages service.Messages,
) *Handler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &Handler{
		transferRepository: transferRepository,
		messageRepository:  messageRepository,
		contracts:          contractsService,
		messages:           messages,
		logger:             config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
	}
}

func (cmh Handler) Handle(payload []byte) {
	m, err := encoding.NewTopicMessageFromBytes(payload)
	if err != nil {
		log.Errorf("Error could not unmarshal payload. Error [%s].", err)
		return
	}

	switch m.Type {
	case validatorproto.TopicMessageType_EthSignature:
		cmh.handleSignatureMessage(*m)
	case validatorproto.TopicMessageType_EthTransaction:
		cmh.handleEthTxMessage(*m)
	default:
		err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
	}

	if err != nil {
		cmh.logger.Errorf("Error - could not handle payload: [%s]", err)
		return
	}
}

func (cmh Handler) handleEthTxMessage(tm encoding.TopicMessage) {
	ethTxMessage := tm.GetTopicEthTransactionMessage()
	isValid, err := cmh.messages.VerifyEthereumTxAuthenticity(tm)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to verify Ethereum TX [%s]. Error: [%s]", ethTxMessage.TransferID, ethTxMessage.EthTxHash, err)
		return
	}
	if !isValid {
		cmh.logger.Infof("[%s] - Provided Ethereum TX [%s] is not the required Mint Transaction", ethTxMessage.TransferID, ethTxMessage.EthTxHash)
		return
	}

	// Process Ethereum Transaction Message
	err = cmh.messages.ProcessEthereumTxMessage(tm)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to process Ethereum TX Message", ethTxMessage.TransferID)
		return
	}
}

// handleSignatureMessage is the main component responsible for the processing of new incoming Signature Messages
func (cmh Handler) handleSignatureMessage(tm encoding.TopicMessage) {
	tsm := tm.GetTopicSignatureMessage()
	valid, err := cmh.messages.SanityCheckSignature(tm)
	if err != nil {
		cmh.logger.Errorf("Failed to perform sanity check on incoming signature [%s] for TX [%s]", tsm.GetSignature(), tsm.TransferID)
		return
	}
	if !valid {
		cmh.logger.Errorf("Incoming signature for TX [%s] is invalid", tsm.TransferID)
		return
	}

	err = cmh.messages.ProcessSignature(tm)
	if err != nil {
		cmh.logger.Errorf("Could not process Signature [%s] for TX [%s]", tsm.GetSignature(), tsm.TransferID)
		return
	}

	majorityReached, shouldExecute, err := cmh.checkMajorityAndExecution(tsm.TransferID)
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not determine whether majority was reached", tsm.TransferID)
		return
	}

	if shouldExecute {
		if majorityReached {
			cmh.logger.Debugf("[%s] - Collected Majority of signatures", tsm.TransferID)
			err = cmh.messages.ScheduleEthereumTxForSubmission(tsm.TransferID)
			if err != nil {
				cmh.logger.Errorf("[%s] - Could not schedule for submission", tsm.TransferID)
			}
		}
	} else {
		cmh.logger.Infof("[%s] - will not be scheduled for submission.", tsm.TransferID)

		if majorityReached {
			err = cmh.transferRepository.UpdateStatusCompleted(tsm.TransferID)
			if err != nil {
				cmh.logger.Errorf("[%s] - Failed to complete. Error: [%s]", tsm.TransferID, err)
			}
		}
	}
}

func (cmh *Handler) checkMajorityAndExecution(transferID string) (majorityReached, shouldExecute bool, err error) {
	signatureMessages, err := cmh.messageRepository.Get(transferID)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to query all Signature Messages. Error: [%s]", transferID, err)
		return false, false, err
	}

	requiredSigCount := len(cmh.contracts.GetMembers())/2 + 1
	cmh.logger.Infof("[%s] - Collected [%d/%d] Signatures", transferID, len(signatureMessages), len(cmh.contracts.GetMembers()))

	return len(signatureMessages) >= requiredSigCount,
		signatureMessages[0].Transfer.ExecuteEthTransaction,
		nil
}
