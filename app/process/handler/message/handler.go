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
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	transferRepository repository.Transfer
	messageRepository  repository.Message
	contracts          map[int64]service.Contracts
	messages           service.Messages
	logger             *log.Entry
}

func NewHandler(
	topicId string,
	transferRepository repository.Transfer,
	messageRepository repository.Message,
	contractServices map[int64]service.Contracts,
	messages service.Messages,
) *Handler {
	topicID, err := hedera.TopicIDFromString(topicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", topicId)
	}

	return &Handler{
		transferRepository: transferRepository,
		messageRepository:  messageRepository,
		contracts:          contractServices,
		messages:           messages,
		logger:             config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
	}
}

func (cmh Handler) Handle(payload interface{}) {
	m, ok := payload.(*message.Message)
	if !ok {
		cmh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	cmh.handleSignatureMessage(*m)
}

// handleSignatureMessage is the main component responsible for the processing of new incoming Signature Messages
func (cmh Handler) handleSignatureMessage(tsm message.Message) {
	valid, err := cmh.messages.SanityCheckSignature(tsm)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to perform sanity check on incoming signature [%s].", tsm.TransferID, tsm.GetSignature())
		return
	}
	if !valid {
		cmh.logger.Errorf("[%s] - Incoming signature is invalid", tsm.TransferID)
		return
	}

	err = cmh.messages.ProcessSignature(tsm)
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not process signature [%s]", tsm.TransferID, tsm.GetSignature())
		return
	}

	majorityReached, err := cmh.checkMajority(tsm.TransferID)
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not determine whether majority was reached", tsm.TransferID)
		return
	}

	if majorityReached {
		err = cmh.transferRepository.UpdateStatusCompleted(tsm.TransferID)
		if err != nil {
			cmh.logger.Errorf("[%s] - Failed to complete. Error: [%s]", tsm.TransferID, err)
		}
	}
}

func (cmh *Handler) checkMajority(transferID string) (majorityReached bool, err error) {
	signatureMessages, err := cmh.messageRepository.Get(transferID)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to query all Signature Messages. Error: [%s]", transferID, err)
		return false, err
	}

	// TODO: remove mockChainID and add actual parameter
	mockChainID := int64(80001)
	membersCount := len(cmh.contracts[mockChainID].GetMembers())
	requiredSigCount := membersCount/2 + 1
	cmh.logger.Infof("[%s] - Collected [%d/%d] Signatures", transferID, len(signatureMessages), membersCount)

	return len(signatureMessages) >= requiredSigCount,
		nil
}
