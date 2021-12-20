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
	"github.com/dariubs/percent"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	prometheusServices "github.com/limechain/hedera-eth-bridge-validator/app/services/prometheus"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math"
	"math/big"
)

type Handler struct {
	transferRepository repository.Transfer
	messageRepository  repository.Message
	contracts          map[int64]service.Contracts
	messages           service.Messages
	logger             *log.Entry
	participationRate  prometheus.Gauge
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

	participationRate := prometheusServices.NewGaugeMetric("validators_participation_rate", "Participation rate: Track validators' activity in %.")

	return &Handler{
		transferRepository: transferRepository,
		messageRepository:  messageRepository,
		contracts:          contractServices,
		messages:           messages,
		logger:             config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
		participationRate: 	participationRate,
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

	majorityReached, err := cmh.checkMajority(tsm.TransferID, int64(tsm.TargetChainId))
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not determine whether majority was reached. Error: [%s]", tsm.TransferID, err)
		return
	}

	if majorityReached {
		err = cmh.transferRepository.UpdateStatusCompleted(tsm.TransferID)
		if err != nil {
			cmh.logger.Errorf("[%s] - Failed to complete. Error: [%s]", tsm.TransferID, err)
		}
	}
}

func (cmh *Handler) checkMajority(transferID string, targetChainId int64) (majorityReached bool, err error) {
	signatureMessages, err := cmh.messageRepository.Get(transferID)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to query all Signature Messages. Error: [%s]", transferID, err)
		return false, err
	}

	membersCount := len(cmh.contracts[targetChainId].GetMembers())
	bnSignaturesLength := big.NewInt(int64(len(signatureMessages)))
	cmh.setParticipationRate(signatureMessages, membersCount)
	cmh.logger.Infof("[%s] - Collected [%d/%d] Signatures", transferID, len(signatureMessages), membersCount)

	return cmh.contracts[targetChainId].
		HasValidSignaturesLength(bnSignaturesLength)
}

func (cmh *Handler) setParticipationRate(signatureMessages []entity.Message, membersCount int) {
	participationRate := math.Round(percent.PercentOf(len(signatureMessages), membersCount)*100) / 100
	cmh.logger.Infof("Percentage callc [%f]", participationRate)
	cmh.participationRate.Set(participationRate)
}
