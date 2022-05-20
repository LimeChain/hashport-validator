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
		Networks: map[uint64]*parser.Network{
			0: {
				Name:                  "Hedera",
				BridgeAccount:         "0.0.111111111",
				PayerAccount:          "0.0.111111111",
				RouterContractAddress: "",
				Members:               []string{"0.0.111111111"},
				Tokens: parser.Tokens{
					Fungible: map[string]parser.Token{
						"HBAR": {
							Fee:               0,
							FeePercentage:     10000,
							MinFeeAmountInUsd: "0.001",
							MinAmount:         nil,
							Networks: map[uint64]string{
								3:     "0xb083879B1e10C8476802016CB12cd2F22a896571",
								80001: "0xb083879B1e10C14761010161B12cd2F25a896691",
							},
							CoinGeckoId:     "hedera-hashgraph",
							CoinMarketCapId: "4642",
						},
					},
					Nft: nil,
				},
			},
		},
		MonitoredAccounts: nil,
	}

	encodedOneChunkConfig = "IyBUaGlzIGZpbGUgY29udGFpbnMgYXBwbGljYXRpb24gZGVmYXVsdHMKIyBDaGVjayBkb2NzL2NvbmZpZ3VyYXRpb24ubWQgZm9yIG1vcmUgaW5mb3JtYXRpb24KYnJpZGdlOgogIHVzZV9sb2NhbF9jb25maWc6IGZhbHNlCiAgY29uZmlnX3RvcGljX2lkOiAwLjAuMgogIHBvbGxpbmdfaW50ZXJ2YWw6IDEwICMgaW4gc2Vjb25kcwogIHRvcGljX2lkOiAwLjAuMQogIG5ldHdvcmtzOgogICAgMDogIyBIZWRlcmEKICAgICAgbmFtZTogSGVkZXJhCiAgICAgIGJyaWRnZV9hY2NvdW50OiAwLjAuMTExMTExMTExCiAgICAgIHBheWVyX2FjY291bnQ6IDAuMC4xMTExMTExMTEKICAgICAgbWVtYmVyczoKICAgICAgICAtIDAuMC4xMTExMTExMTEKICAgICAgdG9rZW5zOgogICAgICAgIGZ1bmdpYmxlOgogICAgICAgICAgIkhCQVIiOgogICAgICAgICAgICBmZWVfcGVyY2VudGFnZTogMTAwMDAgIyAxMC4wMDAlCiAgICAgICAgICAgIGNvaW5fZ2Vja29faWQ6ICJoZWRlcmEtaGFzaGdyYXBoIgogICAgICAgICAgICBjb2luX21hcmtldF9jYXBfaWQ6ICI0NjQyIgogICAgICAgICAgICBtaW5fZmVlX2Ftb3VudF9pbl91c2Q6IDAuMDAxICMgVVNECiAgICAgICAgICAgIG5ldHdvcmtzOgogICAgICAgICAgICAgIDM6ICIweGIwODM4NzlCMWUxMEM4NDc2ODAyMDE2Q0IxMmNkMkYyMmE4OTY1NzEiCiAgICAgICAgICAgICAgODAwMDE6ICIweGIwODM4NzlCMWUxMEMxNDc2MTAxMDE2MUIxMmNkMkYyNWE4OTY2OTEi"
	encodedTwoChunkConfig = []string{
		"IyBUaGlzIGZpbGUgY29udGFpbnMgYXBwbGljYXRpb24gZGVmYXVsdHMKIyBDaGVjayBkb2NzL2NvbmZpZ3VyYXRpb24ubWQgZm9yIG1vcmUgaW5mb3JtYXRpb24KYnJpZGdlOgogIHVzZV9sb2NhbF9jb25maWc6IGZhbHNlCiAgY29uZmlnX3RvcGljX2lkOiAwLjAuMgogIHBvbGxpbmdfaW50ZXJ2YWw6IDEwICMgaW4gc2Vjb25kcwogIHRvcGljX2lkOiAwLjAuMQogIG5ldHdvcmtzOgo=",
		"ICAgIDA6ICMgSGVkZXJhCiAgICAgIG5hbWU6IEhlZGVyYQogICAgICBicmlkZ2VfYWNjb3VudDogMC4wLjExMTExMTExMQogICAgICBwYXllcl9hY2NvdW50OiAwLjAuMTExMTExMTExCiAgICAgIG1lbWJlcnM6CiAgICAgICAgLSAwLjAuMTExMTExMTExCiAgICAgIHRva2VuczoKICAgICAgICBmdW5naWJsZToKICAgICAgICAgICJIQkFSIjoKICAgICAgICAgICAgZmVlX3BlcmNlbnRhZ2U6IDEwMDAwICMgMTAuMDAwJQogICAgICAgICAgICBjb2luX2dlY2tvX2lkOiAiaGVkZXJhLWhhc2hncmFwaCIKICAgICAgICAgICAgY29pbl9tYXJrZXRfY2FwX2lkOiAiNDY0MiIKICAgICAgICAgICAgbWluX2ZlZV9hbW91bnRfaW5fdXNkOiAwLjAwMSAjIFVTRAogICAgICAgICAgICBuZXR3b3JrczoKICAgICAgICAgICAgICAzOiAiMHhiMDgzODc5QjFlMTBDODQ3NjgwMjAxNkNCMTJjZDJGMjJhODk2NTcxIgogICAgICAgICAgICAgIDgwMDAxOiAiMHhiMDgzODc5QjFlMTBDMTQ3NjEwMTAxNjFCMTJjZDJGMjVhODk2NjkxIg==",
	}
	twoMsgs                 = helper.MakeMessagePerChunk(encodedTwoChunkConfig, consensusTimestampStr, topicId.String())
	encodedThreeChunkConfig = []string{
		"IyBUaGlzIGZpbGUgY29udGFpbnMgYXBwbGljYXRpb24gZGVmYXVsdHMKIyBDaGVjayBkb2NzL2NvbmZpZ3VyYXRpb24ubWQgZm9yIG1vcmUgaW5mb3JtYXRpb24KYnJpZGdlOgogIHVzZV9sb2NhbF9jb25maWc6IGZhbHNlCiAgY29uZmlnX3RvcGljX2lkOiAwLjAuMgogIHBvbGxpbmdfaW50ZXJ2YWw6IDEwICMgaW4gc2Vjb25kcwo=",
		"ICB0b3BpY19pZDogMC4wLjEKICBuZXR3b3JrczoKICAgIDA6ICMgSGVkZXJhCiAgICAgIG5hbWU6IEhlZGVyYQogICAgICBicmlkZ2VfYWNjb3VudDogMC4wLjExMTExMTExMQogICAgICBwYXllcl9hY2NvdW50OiAwLjAuMTExMTExMTExCiAgICAgIG1lbWJlcnM6CiAgICAgICAgLSAwLjAuMTExMTExMTExCg==",
		"ICAgICAgdG9rZW5zOgogICAgICAgIGZ1bmdpYmxlOgogICAgICAgICAgIkhCQVIiOgogICAgICAgICAgICBmZWVfcGVyY2VudGFnZTogMTAwMDAgIyAxMC4wMDAlCiAgICAgICAgICAgIGNvaW5fZ2Vja29faWQ6ICJoZWRlcmEtaGFzaGdyYXBoIgogICAgICAgICAgICBjb2luX21hcmtldF9jYXBfaWQ6ICI0NjQyIgogICAgICAgICAgICBtaW5fZmVlX2Ftb3VudF9pbl91c2Q6IDAuMDAxICMgVVNECiAgICAgICAgICAgIG5ldHdvcmtzOgogICAgICAgICAgICAgIDM6ICIweGIwODM4NzlCMWUxMEM4NDc2ODAyMDE2Q0IxMmNkMkYyMmE4OTY1NzEiCiAgICAgICAgICAgICAgODAwMDE6ICIweGIwODM4NzlCMWUxMEMxNDc2MTAxMDE2MUIxMmNkMkYyNWE4OTY2OTEi",
	}
	threeMsgs = helper.MakeMessagePerChunk(encodedThreeChunkConfig, consensusTimestampStr, topicId.String())
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
		queryDefaultLimit: queryDefaultLimit,
		queryMaxLimit:     queryMaxLimit,
		logger:            config.GetLoggerFor("Bridge Config Service"),
	}
}
