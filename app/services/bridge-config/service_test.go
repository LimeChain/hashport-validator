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

package bridge_config

import (
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	test_config "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	serviceInstance       *Service
	queryDefaultLimit     = int64(1)
	queryMaxLimit         = int64(1)
	consensusTimestampStr = "1652341810.085288647"
	consensusTimestamp, _ = timestamp.FromString(consensusTimestampStr)
	configTopicId         = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 2,
	}
	topicId = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 1,
	}
	initialTransactionId = message.InitialTransactionId{
		AccountId:             "0.0.111111",
		Nonce:                 0,
		Scheduled:             false,
		TransactionValidStart: "",
	}
	expectedParsedBridge = &parser.Bridge{
		UseLocalConfig:  false,
		ConfigTopicId:   configTopicId.String(),
		PollingInterval: time.Duration(10),
		TopicId:         topicId.String(),
		Networks: parser.Networks{
			Hedera: map[uint64]*parser.HederaNetwork{
				0: {
					Name:          "Hedera",
					BridgeAccount: "0.0.111111111",
					PayerAccount:  "0.0.111111111",
					Members:       []string{"0.0.111111111"},
				},
			},
		},
		RegularTokens: map[string]*parser.RegularToken{
			"HBAR": {
				NativeChain:     0,
				CoinGeckoId:     "hedera-hashgraph",
				CoinMarketCapId: "4642",
				FeePercentage:   10000,
				AddressesPerNetwork: map[uint64]string{
					3:     "0xb083879B1e10C8476802016CB12cd2F22a896571",
					80001: "0xb083879B1e10C14761010161B12cd2F25a896691",
				},
			},
		},
		MonitoredAccounts: nil,
	}

	encodedOneChunkConfig = "YnJpZGdlOgogIHVzZV9sb2NhbF9jb25maWc6IGZhbHNlCiAgY29uZmlnX3RvcGljX2lkOiAwLjAuMgogIHBvbGxpbmdfaW50ZXJ2YWw6IDEwCiAgdG9waWNfaWQ6IDAuMC4xCiAgbmV0d29ya3M6CiAgICBoZWRlcmE6CiAgICAgIDA6CiAgICAgICAgbmFtZTogSGVkZXJhCiAgICAgICAgYnJpZGdlX2FjY291bnQ6IDAuMC4xMTExMTExMTEKICAgICAgICBwYXllcl9hY2NvdW50OiAwLjAuMTExMTExMTExCiAgICAgICAgbWVtYmVyczoKICAgICAgICAgIC0gMC4wLjExMTExMTExMQogIHJlZ3VsYXJfdG9rZW5zOgogICAgIkhCQVIiOgogICAgICBuYXRpdmVfY2hhaW46IDAKICAgICAgY29pbl9nZWNrb19pZDogImhlZGVyYS1oYXNoZ3JhcGgiCiAgICAgIGNvaW5fbWFya2V0X2NhcF9pZDogIjQ2NDIiCiAgICAgIGZlZV9wZXJjZW50YWdlOiAxMDAwMAogICAgICBhZGRyZXNzZXNfcGVyX25ldHdvcms6CiAgICAgICAgMzogIjB4YjA4Mzg3OUIxZTEwQzg0NzY4MDIwMTZDQjEyY2QyRjIyYTg5NjU3MSIKICAgICAgICA4MDAwMTogIjB4YjA4Mzg3OUIxZTEwQzE0NzYxMDEwMTYxQjEyY2QyRjI1YTg5NjY5MSIK"
	encodedTwoChunkConfig = []string{
		"YnJpZGdlOgogIHVzZV9sb2NhbF9jb25maWc6IGZhbHNlCiAgY29uZmlnX3RvcGljX2lkOiAwLjAuMgogIHBvbGxpbmdfaW50ZXJ2YWw6IDEwCiAgdG9waWNfaWQ6IDAuMC4xCiAgbmV0d29ya3M6CiAgICBoZWRlcmE6CiAgICAgIDA6CiAgICAgICAgbmFtZTogSGVkZXJh",
		"CiAgICAgICAgYnJpZGdlX2FjY291bnQ6IDAuMC4xMTExMTExMTEKICAgICAgICBwYXllcl9hY2NvdW50OiAwLjAuMTExMTExMTExCiAgICAgICAgbWVtYmVyczoKICAgICAgICAgIC0gMC4wLjExMTExMTExMQogIHJlZ3VsYXJfdG9rZW5zOgogICAgIkhCQVIiOgogICAgICBuYXRpdmVfY2hhaW46IDAKICAgICAgY29pbl9nZWNrb19pZDogImhlZGVyYS1oYXNoZ3JhcGgiCiAgICAgIGNvaW5fbWFya2V0X2NhcF9pZDogIjQ2NDIiCiAgICAgIGZlZV9wZXJjZW50YWdlOiAxMDAwMAogICAgICBhZGRyZXNzZXNfcGVyX25ldHdvcms6CiAgICAgICAgMzogIjB4YjA4Mzg3OUIxZTEwQzg0NzY4MDIwMTZDQjEyY2QyRjIyYTg5NjU3MSIKICAgICAgICA4MDAwMTogIjB4YjA4Mzg3OUIxZTEwQzE0NzYxMDEwMTYxQjEyY2QyRjI1YTg5NjY5MSIK",
	}
	twoMsgs                 = helper.MakeMessagePerChunk(encodedTwoChunkConfig, consensusTimestampStr, topicId.String())
	encodedThreeChunkConfig = []string{
		"YnJpZGdlOgogIHVzZV9sb2NhbF9jb25maWc6IGZhbHNlCiAgY29uZmlnX3RvcGljX2lkOiAwLjAuMgogIHBvbGxpbmdfaW50ZXJ2YWw6IDEwCiAgdG9waWNfaWQ6IDAuMC4xCiAgbmV0d29ya3M6CiAgICBoZWRlcmE6CiAgICAgIDA6CiAgICAgICAgbmFtZTogSGVkZXJh",
		"CiAgICAgICAgYnJpZGdlX2FjY291bnQ6IDAuMC4xMTExMTExMTEKICAgICAgICBwYXllcl9hY2NvdW50OiAwLjAuMTExMTExMTExCiAgICAgICAgbWVtYmVyczoKICAgICAgICAgIC0gMC4wLjExMTExMTExMQogIHJlZ3VsYXJfdG9rZW5zOgogICAgIkhCQVIiOgogICAg",
		"ICBuYXRpdmVfY2hhaW46IDAKICAgICAgY29pbl9nZWNrb19pZDogImhlZGVyYS1oYXNoZ3JhcGgiCiAgICAgIGNvaW5fbWFya2V0X2NhcF9pZDogIjQ2NDIiCiAgICAgIGZlZV9wZXJjZW50YWdlOiAxMDAwMAogICAgICBhZGRyZXNzZXNfcGVyX25ldHdvcms6CiAgICAgICAgMzogIjB4YjA4Mzg3OUIxZTEwQzg0NzY4MDIwMTZDQjEyY2QyRjIyYTg5NjU3MSIKICAgICAgICA4MDAwMTogIjB4YjA4Mzg3OUIxZTEwQzE0NzYxMDEwMTYxQjEyY2QyRjI1YTg5NjY5MSIK",
	}
	threeMsgs = helper.MakeMessagePerChunk(encodedThreeChunkConfig, consensusTimestampStr, topicId.String())
	nilMsg    *message.Message
	nilMsgs   []message.Message
	returnErr = errors.New("some-error")
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService(&test_config.TestConfig, &testConstants.ParserBridge, mocks.MHederaMirrorClient)

	assert.Equal(t, serviceInstance, actualService)
}

func Test_ProcessLatestConfig_OneChunk(t *testing.T) {
	setup()
	messageFromTopic := helper.NewMessage(
		consensusTimestampStr,
		configTopicId.String(),
		encodedOneChunkConfig,
		1,
		1,
		1)
	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{messageFromTopic}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Equal(t, *expectedParsedBridge, *parsedBridge)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_TwoChunks(t *testing.T) {
	setup()

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&twoMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, int64(1)).Return([]message.Message{twoMsgs[0]}, nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Equal(t, *expectedParsedBridge, *parsedBridge)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_TwoChunksWithBiggerMaxLimit(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 3

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&twoMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, int64(len(twoMsgs))).Return([]message.Message{twoMsgs[0], twoMsgs[1]}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)
	serviceInstance.queryMaxLimit = queryMaxLimit

	assert.Equal(t, *expectedParsedBridge, *parsedBridge)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_WaitingChunks(t *testing.T) {
	setup()
	waitSleepTime = 0 // Changing to 0 to avoid sleep while testing

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{twoMsgs[0]}, nil).Once()
	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&twoMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, int64(1)).Return([]message.Message{twoMsgs[0]}, nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp, int64(1)).Return([]message.Message{twoMsgs[1]}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Equal(t, *expectedParsedBridge, *parsedBridge)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_ThreeChunksWithLeftOver(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 2

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return([]message.Message{threeMsgs[0], threeMsgs[1]}, nil).Once()
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp+1, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)
	serviceInstance.queryMaxLimit = queryMaxLimit

	assert.Equal(t, *expectedParsedBridge, *parsedBridge)
	assert.Nil(t, err)
}

func Test_ProcessLatestConfig_ErrFirstReqLastMsg(t *testing.T) {
	setup()
	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return(nilMsgs, returnErr)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedBridge)
}

func Test_ProcessLatestConfig_NoNewMessages(t *testing.T) {
	setup()
	serviceInstance.milestoneTimestamp = consensusTimestamp
	messageFromTopic := helper.NewMessage(
		consensusTimestampStr,
		configTopicId.String(),
		encodedOneChunkConfig,
		1,
		1,
		1)
	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{messageFromTopic}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)
	serviceInstance.milestoneTimestamp = 0

	assert.Nil(t, err)
	assert.Nil(t, parsedBridge)
}

func Test_ProcessLatestConfig_ErrInvalidContent(t *testing.T) {
	setup()
	messageFromTopic := helper.NewMessage(
		consensusTimestampStr,
		configTopicId.String(),
		"___",
		1,
		1,
		1)
	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{messageFromTopic}, nil)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedBridge)
}

func Test_ProcessLatestConfig_ErrOneMissingFromThreeChunks(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 3

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return([]message.Message{threeMsgs[0], threeMsgs[2]}, nil).Once()

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)
	serviceInstance.queryMaxLimit = queryMaxLimit

	assert.Error(t, err)
	assert.Nil(t, parsedBridge)
}

func Test_ProcessLatestConfig_ErrOnFetchingBySequenceNumber(t *testing.T) {
	setup()

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(nilMsg, returnErr)

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedBridge)
}

func Test_ProcessLatestConfig_ErrOnFetchingAllChunks(t *testing.T) {
	setup()

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return(nilMsgs, returnErr).Once()

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedBridge)
}

func Test_ProcessLatestConfig_ErrOnFetchingLeftOverChunks(t *testing.T) {
	setup()
	serviceInstance.queryMaxLimit = 2

	mocks.MHederaMirrorClient.On("GetLatestMessages", configTopicId, int64(1)).Return([]message.Message{threeMsgs[2]}, nil)
	mocks.MHederaMirrorClient.On("GetMessageBySequenceNumber", configTopicId, int64(1)).Return(&threeMsgs[0], nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp-1, serviceInstance.queryMaxLimit).Return([]message.Message{threeMsgs[0], threeMsgs[1]}, nil).Once()
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", configTopicId, consensusTimestamp+1, int64(1)).Return(nilMsgs, returnErr).Once()

	parsedBridge, err := serviceInstance.ProcessLatestConfig(configTopicId)

	assert.Error(t, err)
	assert.Nil(t, parsedBridge)
}

func setup() {
	mocks.Setup()
	helper.SetupNetworks()
	mocks.MHederaMirrorClient.On("QueryDefaultLimit").Return(queryDefaultLimit)
	mocks.MHederaMirrorClient.On("QueryMaxLimit").Return(queryMaxLimit)

	serviceInstance = &Service{
		mirrorNode:        mocks.MHederaMirrorClient,
		config:            &test_config.TestConfig,
		parsedBridgeCfg:   &testConstants.ParserBridge,
		queryDefaultLimit: queryDefaultLimit,
		queryMaxLimit:     queryMaxLimit,
		logger:            config.GetLoggerFor("Bridge Config Service"),
	}
}
