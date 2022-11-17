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

package fee_policy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirrorNodeMsg "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	fee_policy "github.com/limechain/hedera-eth-bridge-validator/app/model/fee-policy"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"

	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	waitSleepTime = time.Duration(2)
)

type Service struct {
	mirrorNode            client.MirrorNode
	milestoneTimestamp    int64
	queryDefaultLimit     int64
	queryMaxLimit         int64
	config                *config.Config
	parsedFeePolicyConfig *parser.FeePolicy
	feePolicyConfig       *config.FeePolicy
	logger                *log.Entry
}

func NewService(cfg *config.Config, mirrorNode client.MirrorNode) *Service {
	return &Service{
		mirrorNode:        mirrorNode,
		queryMaxLimit:     mirrorNode.QueryMaxLimit(),
		queryDefaultLimit: mirrorNode.QueryDefaultLimit(),
		config:            cfg,
		logger:            config.GetLoggerFor("Fee Policy Config Service"),
		feePolicyConfig:   &config.FeePolicy{},
	}
}

func (svc *Service) FeeAmountFor(networkId uint64, account string, token string, amount int64) (feeAmount int64, exist bool) {
	if svc.feePolicyConfig.StoreAddresses != nil {
		policy := svc.feePolicyConfig.StoreAddresses[account]

		if policy != nil {
			return policy.FeeAmountFor(networkId, token, amount)
		}
	}

	return 0, false
}

func (svc *Service) ProcessLatestFeePolicyConfig(topicID hedera.TopicID) (*parser.FeePolicy, error) {
	latestMessages, err := svc.mirrorNode.GetLatestMessages(topicID, 1)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to get latest messages from topic: [%s]. Err: [%s]", topicID.String(), err)
		return nil, errors.New(errMsg)
	}

	if len(latestMessages) == 0 {
		svc.logger.Infof("No new fee policy config messages to process.")
		return nil, nil
	}

	lastMessage := latestMessages[0]
	latestConsensusTimestamp, _ := timestamp.FromString(lastMessage.ConsensusTimestamp)

	if latestConsensusTimestamp == svc.milestoneTimestamp {
		svc.logger.Infof("No new fee policy config messages to process.")
		return nil, nil
	}

	if lastMessage.ChunkInfo.Total == 1 {
		// The whole config content is in 1 message
		decodedMsgContent, err := svc.decodeMsgContent(lastMessage)
		if err != nil {
			return nil, err
		}
		return svc.processFullMsgContent(decodedMsgContent, lastMessage.ConsensusTimestamp)
	}

	if lastMessage.ChunkInfo.Number < lastMessage.ChunkInfo.Total {
		lastMessage, _ = svc.waitForAllChunks(topicID, lastMessage)
	}

	var messagesToProcess []mirrorNodeMsg.Message
	messagesToProcess, err = svc.fetchAllChunks(topicID, lastMessage)

	if err != nil {
		return nil, err
	}

	return svc.processAllMessages(messagesToProcess)
}

func (svc *Service) fetchAllChunks(topicID hedera.TopicID, lastMessage mirrorNodeMsg.Message) ([]mirrorNodeMsg.Message, error) {
	firstChunkMsgSeqNum := lastMessage.SequenceNumber - (lastMessage.ChunkInfo.Total - 1)
	msg, err := svc.mirrorNode.GetMessageBySequenceNumber(topicID, firstChunkMsgSeqNum)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to get first message chunk by sequence number - [%d]. Err: [%s]", firstChunkMsgSeqNum, err)
		return nil, errors.New(errMsg)
	}

	firstConsensusTimestamp, _ := timestamp.FromString(msg.ConsensusTimestamp)
	if lastMessage.ChunkInfo.Total <= svc.queryMaxLimit {
		allChunks, err := svc.mirrorNode.GetMessagesAfterTimestamp(topicID, firstConsensusTimestamp-1, lastMessage.ChunkInfo.Total)

		if err != nil {
			errMsg := fmt.Sprintf("Failed to fetch messages after first consensus timestamp - [%d]. Err: [%s]", firstConsensusTimestamp, err)
			return nil, errors.New(errMsg)
		}

		return allChunks, nil
	}

	consensusTimestamp := firstConsensusTimestamp - 1
	countOfRequestsWithMaxLimit := int(lastMessage.ChunkInfo.Total / svc.queryMaxLimit)
	leftOverChunks := lastMessage.ChunkInfo.Total % svc.queryMaxLimit
	allChunks := make([]mirrorNodeMsg.Message, 0)

	for i := 0; i < countOfRequestsWithMaxLimit; i++ {
		currMsgs, err := svc.mirrorNode.GetMessagesAfterTimestamp(topicID, consensusTimestamp, svc.queryMaxLimit)

		if err != nil {
			errMsg := fmt.Sprintf("Failed to fetch messages after first consensus timestamp - [%d]. Err: [%s]", consensusTimestamp, err)
			return nil, errors.New(errMsg)
		}

		allChunks = append(allChunks, currMsgs...)
		consensusTimestamp, _ = timestamp.FromString(currMsgs[len(currMsgs)-1].ConsensusTimestamp)
	}

	if leftOverChunks > 0 {
		currMsgs, err := svc.mirrorNode.GetMessagesAfterTimestamp(topicID, consensusTimestamp, leftOverChunks)

		if err != nil {
			errMsg := fmt.Sprintf("Failed to fetch messages after first consensus timestamp - [%d]. Err: [%s]", consensusTimestamp, err)
			return nil, errors.New(errMsg)
		}

		allChunks = append(allChunks, currMsgs...)
	}

	return allChunks, nil
}

func (svc *Service) waitForAllChunks(topicID hedera.TopicID, lastMessage mirrorNodeMsg.Message) (mirrorNodeMsg.Message, error) {
	var err error
	for lastMessage.ChunkInfo.Number < lastMessage.ChunkInfo.Total {
		time.Sleep(waitSleepTime * time.Second)
		var latestMessages []mirrorNodeMsg.Message
		latestMessages, err = svc.mirrorNode.GetLatestMessages(topicID, 1)

		if err != nil {
			return lastMessage, err
		}

		lastMessage = latestMessages[0]
	}

	return lastMessage, err
}

func (svc *Service) processAllMessages(allMessages []mirrorNodeMsg.Message) (*parser.FeePolicy, error) {
	chunksProcessor := new(chunkInfosProcessor)

	for _, msg := range allMessages {
		allChunksProcessed, err := svc.processMessage(msg, chunksProcessor)

		if err != nil {
			errMsg := fmt.Sprintf("Failed to process chunk info. Err: [%s]", err)
			return nil, errors.New(errMsg)
		} else {
			if allChunksProcessed {
				// Returning immediately after current config file is fully processed
				return svc.processFullMsgContent(chunksProcessor.content, msg.ConsensusTimestamp)
			}
		}
	}

	return nil, nil
}

func (svc *Service) processFullMsgContent(decodedMsgContent []byte, consensusTimestamp string) (*parser.FeePolicy, error) {
	parsedFeePolicy, err := svc.parseFullMsgContent(decodedMsgContent)
	if err != nil {
		return nil, err
	}
	svc.milestoneTimestamp, _ = timestamp.FromString(consensusTimestamp)
	svc.logger.Infof("Successfully processed latest fee policy config!")

	svc.parsedFeePolicyConfig = parsedFeePolicy

	svc.feePolicyConfig.StoreAddresses = svc.parsePolicyInterfaces(parsedFeePolicy)

	return parsedFeePolicy, nil
}

func (svc *Service) processMessage(msg mirrorNodeMsg.Message, chunksProcessor *chunkInfosProcessor) (bool, error) {
	decodedMsgContent, err := svc.decodeMsgContent(msg)

	if err != nil {
		return false, err
	}

	if chunksProcessor.total == 0 {
		chunksProcessor.total = msg.ChunkInfo.Total
		chunksProcessor.processed = msg.ChunkInfo.Number
		chunksProcessor.content = decodedMsgContent
	} else {
		if !chunksProcessor.allProcessed() {
			err := chunksProcessor.processChunk(msg.ChunkInfo.Number, decodedMsgContent)
			if err != nil {
				return false, err
			}
		}
	}

	if chunksProcessor.allProcessed() {
		return true, nil
	}

	return false, nil
}

func (svc *Service) parseFullMsgContent(content []byte) (*parser.FeePolicy, error) {
	configParser := &parser.FeePolicy{}
	err := yaml.Unmarshal(content, configParser)

	if err != nil {
		svc.logger.Errorf("Failed to parse fee policy config. Err: [%s]", err)
		return nil, err
	}

	svc.parsedFeePolicyConfig = configParser

	svc.feePolicyConfig = &config.FeePolicy{}
	svc.feePolicyConfig.StoreAddresses = svc.parsePolicyInterfaces(configParser)

	return configParser, nil
}

func (svc *Service) decodeMsgContent(msg mirrorNodeMsg.Message) ([]byte, error) {
	decodedMsgContent, err := base64.StdEncoding.DecodeString(msg.Contents)

	if err != nil {
		svc.logger.Errorf("Failed to decode message content from base64 format: [%s]. Err: [%s]", msg.Contents, err)
		return nil, err
	}

	return decodedMsgContent, nil
}

func (svc *Service) parsePolicyInterfaces(configParser *parser.FeePolicy) map[string]fee_policy.FeePolicyInterface {
	result := make(map[string]fee_policy.FeePolicyInterface)

	for _, itemLegalEntity := range configParser.LegalEntities {
		for _, itemAddress := range itemLegalEntity.Addresses {
			switch itemLegalEntity.PolicyInfo.FeeType {
			case constants.FeePolicyTypeFlat:
				result[itemAddress] = fee_policy.ParseNewFlatFeePolicy(itemLegalEntity.PolicyInfo.Networks, itemLegalEntity.PolicyInfo.Value)
			case constants.FeePolicyTypePercentage:
				result[itemAddress] = fee_policy.ParseNewPercentageFeePolicy(itemLegalEntity.PolicyInfo.Networks, itemLegalEntity.PolicyInfo.Value)
			case constants.FeePolicyTypeFlatPerToken:
				result[itemAddress] = fee_policy.ParseNewFlatFeePerTokenPolicy(itemLegalEntity.PolicyInfo.Networks, itemLegalEntity.PolicyInfo.Value)
			default:
				svc.logger.Errorf("Unrecognized fee policy type [%s]", itemLegalEntity.PolicyInfo.FeeType)
			}
		}
	}

	return result
}

type chunkInfosProcessor struct {
	total     int64
	processed int64
	content   []byte
}

func (svc *chunkInfosProcessor) allProcessed() bool {
	return svc.total == svc.processed
}

func (svc *chunkInfosProcessor) processChunk(number int64, content []byte) error {
	var err error
	if number-svc.processed > 1 {
		msg := fmt.Sprintf("missing chunk of fee policy config. latest processed number: %d, and current chunk number: %d", svc.processed, number)
		err = errors.New(msg)
	}

	svc.content = append(svc.content, content...)
	svc.processed = number
	return err
}
