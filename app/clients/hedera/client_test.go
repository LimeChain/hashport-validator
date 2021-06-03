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

package hedera

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	tc "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var nodeClient = NewNodeClient(tc.TestConfig.Validator.Clients.Hedera)

func TestGetClientForTestnet(t *testing.T) {
	client := nodeClient.GetClient()
	assert.NotNil(t, client)
	clientNetwork := client.GetNetwork()
	for key, _ := range clientNetwork {
		assert.True(t, strings.Contains(key, "testnet"))
	}
}

func TestGetClientForPreviewnet(t *testing.T) {
	tc.TestConfig.Validator.Clients.Hedera.NetworkType = "previewnet"
	defer func() { tc.TestConfig.Validator.Clients.Hedera.NetworkType = "testnet" }()
	nodeClient := NewNodeClient(tc.TestConfig.Validator.Clients.Hedera)
	client := nodeClient.GetClient()
	assert.NotNil(t, client)
	clientNetwork := client.GetNetwork()
	for key, _ := range clientNetwork {
		assert.True(t, strings.Contains(key, "previewnet"))
	}
}

func TestGetClientForMainnet(t *testing.T) {
	tc.TestConfig.Validator.Clients.Hedera.NetworkType = "mainnet"
	defer func() { tc.TestConfig.Validator.Clients.Hedera.NetworkType = "testnet" }()
	nodeClient := NewNodeClient(tc.TestConfig.Validator.Clients.Hedera)
	client := nodeClient.GetClient()
	assert.NotNil(t, client)

	clientNetwork := client.GetNetwork()
	for key, _ := range clientNetwork {
		assert.False(t, strings.Contains(key, "testnet"))
		assert.False(t, strings.Contains(key, "previewnet"))
	}
}

func TestGetClientForNetworkError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false

	tc.TestConfig.Validator.Clients.Hedera.NetworkType = "nonExisted"
	defer func() { tc.TestConfig.Validator.Clients.Hedera.NetworkType = "testnet" }()

	assert.Panics(t, func() {
		NewNodeClient(tc.TestConfig.Validator.Clients.Hedera)
	})

	assert.True(t, fatal)
}

func TestGetClientAccIDError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false

	tc.TestConfig.Validator.Clients.Hedera.Operator.AccountId = "000"
	defer func() { tc.TestConfig.Validator.Clients.Hedera.Operator.AccountId = "0.0.478300" }()
	NewNodeClient(tc.TestConfig.Validator.Clients.Hedera)
	assert.True(t, fatal)
}

func TestGetClientPrivateKeyError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false

	tc.TestConfig.Validator.Clients.Hedera.Operator.PrivateKey = "000"
	defer func() {
		tc.TestConfig.Validator.Clients.Hedera.Operator.PrivateKey = "302e020100300506032b657004220420479934e1729d3a2a25f3cdec95862d247944635113b4f4a07ec44c5ff8ec0884"
	}()
	assert.Panics(t, func() {
		NewNodeClient(tc.TestConfig.Validator.Clients.Hedera)
	})
	assert.True(t, fatal)
}

func TestSubmitTopicConsensusMessage(t *testing.T) {
	msg := []byte{0, 1, 2, 3, 4}
	topicID, _ := hedera.TopicIDFromString("0.0.1870757")
	id, err := nodeClient.SubmitTopicConsensusMessage(topicID, msg)
	assert.Nil(t, err)
	assert.NotNil(t, id)
}

func TestSubmitTopicConsensusMessageError(t *testing.T) {
	msg := []byte{0}
	topicID, _ := hedera.TopicIDFromString("")
	response, err := nodeClient.SubmitTopicConsensusMessage(topicID, msg)
	assert.False(t, response.GetScheduled())
	assert.Error(t, err)
}

func TestSubmitScheduleSign(t *testing.T) {
	scheduleID, err := hedera.ScheduleIDFromString("0.0.0")

	txResponse, err := nodeClient.SubmitScheduleSign(scheduleID)
	assert.NotNil(t, txResponse)
	assert.Nil(t, err)
}

func TestSubmitScheduledTokenTransferTransaction(t *testing.T) {
	tokenID, err := hedera.TokenIDFromString("0.0.447200")

	if err != nil {
		panic(err)
	}

	var transfers []transfer.Hedera

	recipient, _ := hedera.AccountIDFromString("0.0.263546")

	transfer := transfer.Hedera{
		AccountID: recipient,
		Amount:    1,
	}

	transfers = append(transfers, transfer)
	payerAccountID, _ := hedera.AccountIDFromString(tc.TestConfig.Validator.Clients.Hedera.PayerAccount)
	memo := "this is memo"

	txResponse, err := nodeClient.SubmitScheduledTokenTransferTransaction(tokenID, transfers, payerAccountID, memo)
	assert.NotNil(t, txResponse)
	assert.Nil(t, err)
}

func TestSubmitScheduledHbarTransferTransaction(t *testing.T) {
	var transfers []transfer.Hedera

	recipient, _ := hedera.AccountIDFromString("0.0.263546")

	transfer := transfer.Hedera{
		AccountID: recipient,
		Amount:    1,
	}

	transfers = append(transfers, transfer)
	payerAccountID, _ := hedera.AccountIDFromString(tc.TestConfig.Validator.Clients.Hedera.PayerAccount)
	memo := "this is memo"

	txResponse, err := nodeClient.SubmitScheduledHbarTransferTransaction(transfers, payerAccountID, memo)
	assert.NotNil(t, txResponse)
	assert.Nil(t, err)
}
