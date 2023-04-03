/*
 * Copyright 2022 LimeChain Ltd.
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
	msgHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math"
	"math/big"
)

type Handler struct {
	transferRepository     repository.Transfer
	messageRepository      repository.Message
	contracts              map[uint64]service.Contracts
	messages               service.Messages
	logger                 *log.Entry
	participationRateGauge prometheus.Gauge
	prometheusService      service.Prometheus
	assetsService          service.Assets
}

func NewHandler(
	topicId string,
	transferRepository repository.Transfer,
	messageRepository repository.Message,
	contractServices map[uint64]service.Contracts,
	messages service.Messages,
	prometheusService service.Prometheus,
	assetsService service.Assets,
) *Handler {
	topicID, err := hedera.TopicIDFromString(topicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", topicId)
	}

	var participationRate prometheus.Gauge
	if prometheusService.GetIsMonitoringEnabled() {
		participationRate = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
			Name: constants.ValidatorsParticipationRateGaugeName,
			Help: constants.ValidatorsParticipationRateGaugeHelp,
		})
		// Set to 100 in order not to fire alerts until a transaction is processed
		participationRate.Set(constants.ValidatorsParticipationRateInitialValue)
	}

	return &Handler{
		transferRepository:     transferRepository,
		messageRepository:      messageRepository,
		contracts:              contractServices,
		messages:               messages,
		logger:                 config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
		prometheusService:      prometheusService,
		participationRateGauge: participationRate,
		assetsService:          assetsService,
	}
}

func (cmh Handler) Handle(payload interface{}) {
	m, ok := payload.(*message.Message)
	if !ok {
		cmh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	switch msg := m.Message.(type) {
	case *proto.TopicMessage_FungibleSignatureMessage:
		msgHelper.UpdateHederaChainIdOfFungibleMsg(msg.FungibleSignatureMessage)
		cmh.handleFungibleSignatureMessage(msg.FungibleSignatureMessage, m.TransactionTimestamp)
		break
	case *proto.TopicMessage_NftSignatureMessage:
		msgHelper.UpdateHederaChainIdOfNftMsg(msg.NftSignatureMessage)
		cmh.handleNftSignatureMessage(msg.NftSignatureMessage, m.TransactionTimestamp)
		break
	default:
		cmh.logger.Errorf("Invalid topic message provided: [%v]", msg)
		break
	}
}

// handleFungibleSignatureMessage is the main component responsible for the processing of new incoming Signature Messages
func (cmh Handler) handleFungibleSignatureMessage(tsm *proto.TopicEthSignatureMessage, timestamp int64) {

	valid, err := cmh.messages.SanityCheckFungibleSignature(tsm)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to perform sanity check on incoming signature [%s].", tsm.TransferID, tsm.GetSignature())
		return
	}
	if !valid {
		cmh.logger.Errorf("[%s] - Incoming signature is invalid", tsm.TransferID)
		return
	}

	// Parse incoming message
	authMsgBytes, err := auth_message.EncodeFungibleBytesFrom(tsm.SourceChainId, tsm.TargetChainId, tsm.TransferID, tsm.Asset, tsm.Recipient, tsm.Amount)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", tsm.TransferID, err)
		return
	}

	err = cmh.messages.ProcessSignature(tsm.TransferID, tsm.Signature, tsm.TargetChainId, timestamp, authMsgBytes)
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not process signature [%s]", tsm.TransferID, tsm.GetSignature())
		return
	}

	cmh.completeTransfer(tsm.TransferID, tsm.TargetChainId, tsm.SourceChainId, tsm.Asset, false)
}

// handleNftSignatureMessage is the main component responsible for the processing of new incoming Signature Messages
func (cmh Handler) handleNftSignatureMessage(tsm *proto.TopicEthNftSignatureMessage, timestamp int64) {
	valid, err := cmh.messages.SanityCheckNftSignature(tsm)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to perform sanity check on nft incoming signature [%s].", tsm.TransferID, tsm.GetSignature())
		return
	}
	if !valid {
		cmh.logger.Errorf("[%s] - Incoming nft signature is invalid", tsm.TransferID)
		return
	}

	// Parse incoming message
	authMsgBytes, err := auth_message.EncodeNftBytesFrom(tsm.SourceChainId, tsm.TargetChainId, tsm.TransferID, tsm.Asset, int64(tsm.TokenId), tsm.Metadata, tsm.Recipient)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to encode the authorisation nft signature. Error: [%s]", tsm.TransferID, err)
		return
	}

	err = cmh.messages.ProcessSignature(tsm.TransferID, tsm.Signature, tsm.TargetChainId, timestamp, authMsgBytes)
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not process nft signature [%s]", tsm.TransferID, tsm.GetSignature())
		return
	}

	cmh.completeTransfer(tsm.TransferID, tsm.TargetChainId, tsm.SourceChainId, tsm.Asset, true)
}

func (cmh Handler) completeTransfer(transferID string, targetChainId, sourceChainId uint64, asset string, isNFT bool) {
	majorityReached, err := cmh.checkMajority(transferID, targetChainId)
	if err != nil {
		cmh.logger.Errorf("[%s] - Could not determine whether majority was reached. Error: [%s]", transferID, err)
		return
	}

	if majorityReached {
		if !isNFT { // metrics for fungible only
			oppositeAsset := cmh.assetsService.OppositeAsset(sourceChainId, targetChainId, asset)
			metrics.SetMajorityReached(
				sourceChainId,
				targetChainId,
				oppositeAsset,
				transferID,
				cmh.prometheusService,
				cmh.logger,
			)
		}
		err = cmh.transferRepository.UpdateStatusCompleted(transferID)
		if err != nil {
			cmh.logger.Errorf("[%s] - Failed to complete. Error: [%s]", transferID, err)
		}
	}
}

func (cmh *Handler) checkMajority(transferID string, targetChainId uint64) (majorityReached bool, err error) {
	signatureMessages, err := cmh.messageRepository.Get(transferID)
	if err != nil {
		cmh.logger.Errorf("[%s] - Failed to query all Signature Messages. Error: [%s]", transferID, err)
		return false, err
	}

	membersCount := len(cmh.contracts[targetChainId].GetMembers())
	bnSignaturesLength := big.NewInt(int64(len(signatureMessages)))
	cmh.setParticipationRate(signatureMessages, membersCount)
	cmh.logger.Infof("[%s] - Collected [%d/%d] Signatures", transferID, len(signatureMessages), membersCount)

	return cmh.contracts[targetChainId].HasValidSignaturesLength(bnSignaturesLength)
}

func (cmh *Handler) setParticipationRate(signatureMessages []entity.Message, membersCount int) {
	if !cmh.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	participationRate := math.Round(percent.PercentOf(len(signatureMessages), membersCount)*100) / 100
	cmh.logger.Debugf("Percentage callc [%f]", participationRate)
	cmh.participationRateGauge.Set(participationRate)
}
