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
	"errors"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	test_config "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	"github.com/stretchr/testify/assert"
)

var (
	serviceInstance       *Service
	queryDefaultLimit     = int64(1)
	queryMaxLimit         = int64(1)
	consensusTimestampStr = "1652341810.085288647"
	consensusTimestamp, _ = timestamp.FromString(consensusTimestampStr)
	feePolicyTopicId      = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 1,
	}
	expectedFeePolicy = &parser.FeePolicy{
		LegalEntities: map[string]*parser.LegalEntity{
			"Some LTD": &parser.LegalEntity{
				Addresses: []string{"0.0.101", "0.0.102", "0.0.103"},
				PolicyInfo: parser.PolicyInfo{
					FeeType:  constants.FeePolicyTypeFlat,
					Networks: []uint64{8001},
					Value:    3000,
				},
			},
		},
	}

	encodedOneChunkConfig = "cG9saWNpZXM6CiAgIlNvbWUgTFREIjoKICAgIGFkZHJlc3NlczoKICAgICAgLSAiMC4wLjEwMSIKICAgICAgLSAiMC4wLjEwMiIKICAgICAgLSAiMC4wLjEwMyIKICAgIHBvbGljeToKICAgICAgZmVlX3R5cGU6ICJmbGF0IgogICAgICBuZXR3b3JrczoKICAgICAgICAtIDgwMDEKICAgICAgdmFsdWU6IDMwMDA="
	encodedTwoChunkConfig = []string{
		"cG9saWNpZXM6CiAgIlNvbWUgTFREIjoKICAgIGFkZHJlc3NlczoKICAgICAgLSAiMC4wLjEwMSIKICAgICAgLSAiMC4wLjEwMiIKICAgICAgLSAiMC4wLjEwMyI=",
		"ICAgIHBvbGljeToKICAgICAgZmVlX3R5cGU6ICJmbGF0IgogICAgICBuZXR3b3JrczoKICAgICAgICAtIDgwMDEKICAgICAgdmFsdWU6IDIwMDA=",
	}
	twoMsgs                 = helper.MakeMessagePerChunk(encodedTwoChunkConfig, consensusTimestampStr, feePolicyTopicId.String())
	encodedThreeChunkConfig = []string{
		"cG9saWNpZXM6CiAgIlNvbWUgTFREIjoKICAgIGFkZHJlc3NlczoKICAgICAgLSAiMC4wLjEwMSIKICAgICAgLSAiMC4wLjEwMiIKICAgICAgLSAiMC4wLjEwMyI=",
		"ICAgIHBvbGljeToKICAgICAgZmVlX3R5cGU6ICJmbGF0Ig==",
		"ICAgICAgbmV0d29ya3M6CiAgICAgICAgLSA4MDAxCiAgICAgIHZhbHVlOiAyMDAw",
	}
	threeMsgs = helper.MakeMessagePerChunk(encodedThreeChunkConfig, consensusTimestampStr, feePolicyTopicId.String())
	nilMsg    *message.Message
	nilMsgs   []message.Message
	returnErr = errors.New("some-error")
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService(&test_config.TestConfig, mocks.MHederaMirrorClient)

	assert.Equal(t, serviceInstance, actualService)
}

func Test_ProcessLatestConfig_OneChunk(t *testing.T) {
	setup()
	messageFromTopic := helper.NewMessage(
		consensusTimestampStr,
		feePolicyTopicId.String(),
		encodedOneChunkConfig,
		1,
		1,
		1)
	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{messageFromTopic}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Equal(t, *expectedFeePolicy, *parsedFeePolicy)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_TwoChunks(t *testing.T) {
	setup()

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&twoMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, int64(1)).Return([]message.Message{twoMsgs[0]}, nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Equal(t, *expectedFeePolicy, *parsedFeePolicy)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_TwoChunksWithBiggerMaxLimit(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 3

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&twoMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, int64(len(twoMsgs))).Return([]message.Message{twoMsgs[0], twoMsgs[1]}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)
	serviceInstance.queryMaxLimit = queryMaxLimit

	assert.Equal(t, *expectedFeePolicy, *parsedFeePolicy)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_WaitingChunks(t *testing.T) {
	setup()
	waitSleepTime = 0 // Changing to 0 to avoid sleep while testing

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{twoMsgs[0]}, nil).Once()
	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&twoMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, int64(1)).Return([]message.Message{twoMsgs[0]}, nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Equal(t, *expectedFeePolicy, *parsedFeePolicy)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_ThreeChunksWithLeftOver(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 2

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return([]message.Message{threeMsgs[0], threeMsgs[1]}, nil).Once()
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp+1, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)
	serviceInstance.queryMaxLimit = queryMaxLimit

	assert.Equal(t, *expectedFeePolicy, *parsedFeePolicy)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_ErrFirstReqLastMsg(t *testing.T) {
	setup()
	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return(nilMsgs, returnErr)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func Test_ProcessLatestConfig_NoNewMessages(t *testing.T) {
	setup()
	serviceInstance.milestoneTimestamp = consensusTimestamp
	messageFromTopic := helper.NewMessage(
		consensusTimestampStr,
		feePolicyTopicId.String(),
		encodedOneChunkConfig,
		1,
		1,
		1)
	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{messageFromTopic}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)
	serviceInstance.milestoneTimestamp = 0

	assert.Nil(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func Test_ProcessLatestConfig_ErrInvalidContent(t *testing.T) {
	setup()
	messageFromTopic := helper.NewMessage(
		consensusTimestampStr,
		feePolicyTopicId.String(),
		"___",
		1,
		1,
		1)
	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{messageFromTopic}, nil)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func Test_ProcessLatestConfig_ErrOneMissingFromThreeChunks(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 3

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return([]message.Message{threeMsgs[0], threeMsgs[2]}, nil).Once()

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)
	serviceInstance.queryMaxLimit = queryMaxLimit

	assert.Error(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func Test_ProcessLatestConfig_ErrOnFetchingBySequenceNumber(t *testing.T) {
	setup()

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(nilMsg, returnErr)

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func Test_ProcessLatestConfig_ErrOnFetchingAllChunks(t *testing.T) {
	setup()

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return(nilMsgs, returnErr).Once()

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func Test_ProcessLatestConfig_ErrOnFetchingLeftOverChunks(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 2

	mocks.MHederaMirrorClient.On("GetLatestMessages", feePolicyTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", feePolicyTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return([]message.Message{threeMsgs[0], threeMsgs[1]}, nil).Once()
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", feePolicyTopicId, consensusTimestamp+1, int64(1)).Return(nilMsgs, returnErr).Once()

	parsedFeePolicy, err := serviceInstance.ProcessLatestFeePolicyConfig(feePolicyTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedFeePolicy)
}

func setup() {
	mocks.Setup()
	helper.SetupNetworks()
	mocks.MHederaMirrorClient.On("QueryDefaultLimit").Return(queryDefaultLimit)
	mocks.MHederaMirrorClient.On("QueryMaxLimit").Return(queryMaxLimit)

	serviceInstance = &Service{
		mirrorNode:            mocks.MHederaMirrorClient,
		config:                &test_config.TestConfig,
		parsedFeePolicyConfig: &testConstants.ParsedFeePolicyConfig,
		feePolicyConfig:       &testConstants.FeePolicyConfig,
		queryDefaultLimit:     queryDefaultLimit,
		queryMaxLimit:         queryMaxLimit,
		logger:                config.GetLoggerFor("Fee Policy Config Service"),
	}
}
