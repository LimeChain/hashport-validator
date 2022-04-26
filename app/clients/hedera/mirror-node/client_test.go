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

package mirror_node

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	mirrorAPIAddress = "some-api-address"
	pollingInterval  = 5 * time.Second
	mirrorNodeCfg    = config.MirrorNode{
		ApiAddress:      mirrorAPIAddress,
		PollingInterval: pollingInterval,
	}
	logger = config.GetLoggerFor("Mirror Node Client")

	accountId = hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 1,
	}

	topicId = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 2,
	}

	scheduleId = hedera.ScheduleID{
		Shard:    0,
		Realm:    0,
		Schedule: 3,
	}

	c *Client
)

func setup() {
	mocks.Setup()
	c = &Client{
		mirrorAPIAddress: mirrorAPIAddress,
		httpClient:       mocks.MHTTPClient,
		pollingInterval:  5 * time.Second,
		logger:           logger,
	}
}

func Test_NewClient(t *testing.T) {
	setup()
	newClient := NewClient(mirrorNodeCfg)
	assert.Equal(t, c.mirrorAPIAddress, newClient.mirrorAPIAddress)
	assert.Equal(t, c.pollingInterval, newClient.pollingInterval)
	assert.Equal(t, c.logger, newClient.logger)
}

func Test_GetAccountTokenMintTransactionsAfterTimestamp_ThrowsError(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountTokenMintTransactionsAfterTimestamp(accountId, time.Now().UnixNano())
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetSchedule_Status400(t *testing.T) {
	setup()
	response := &http.Response{
		StatusCode: 400,
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(response, nil)
	schedule, err := c.GetSchedule("0.0.2")
	assert.Nil(t, schedule)
	assert.NotNil(t, err)
}

func Test_GetSchedule_Fails(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	schedule, err := c.GetSchedule("0.0.2")
	assert.Nil(t, schedule)
	assert.NotNil(t, err)
}

func Test_AccountExists_Status400(t *testing.T) {
	setup()
	stringReader := strings.NewReader("error")
	stringReadCloser := ioutil.NopCloser(stringReader)
	response := &http.Response{
		StatusCode: 400,
		Body:       stringReadCloser,
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(response, nil)
	exists := c.AccountExists(accountId)
	assert.False(t, exists)
}

func Test_GetAccount_Status400(t *testing.T) {
	setup()
	response := &http.Response{
		StatusCode: 400,
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(response, nil)
	schedule, err := c.GetAccount("0.0.2")
	assert.Nil(t, schedule)
	assert.NotNil(t, err)
}

func Test_GetAccount_Fails(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	schedule, err := c.GetAccount("0.0.2")
	assert.Nil(t, schedule)
	assert.NotNil(t, err)
}

func Test_GetToken_Status400(t *testing.T) {
	setup()
	response := &http.Response{
		StatusCode: 400,
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(response, nil)
	schedule, err := c.GetToken("0.0.2")
	assert.Nil(t, schedule)
	assert.NotNil(t, err)
}

func Test_GetToken_Fails(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	schedule, err := c.GetToken("0.0.2")
	assert.Nil(t, schedule)
	assert.NotNil(t, err)
}

func Test_TopicExists_Status400(t *testing.T) {
	setup()
	stringReader := strings.NewReader("error")
	stringReadCloser := ioutil.NopCloser(stringReader)
	response := &http.Response{
		StatusCode: 400,
		Body:       stringReadCloser,
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(response, nil)
	exists := c.TopicExists(topicId)
	assert.False(t, exists)
}

func Test_GetAccountTokenBurnTransactionsAfterTimestamp(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountTokenBurnTransactionsAfterTimestamp(accountId, time.Now().UnixNano())
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetMessagesAfterTimestamp(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetMessagesAfterTimestamp(topicId, time.Now().UnixNano())
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetTransaction(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetTransaction("txid")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetScheduledTransaction(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetScheduledTransaction("txid")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetStateProof_Status400(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 400,
	}, nil)
	response, err := c.GetStateProof("txid")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetStateProof_Fails(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetStateProof("txid")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountCreditTransactionsAfterTimestamp(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountCreditTransactionsAfterTimestamp(accountId, time.Now().UnixNano())
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountDebitTransactionsAfterTimestampString(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountDebitTransactionsAfterTimestampString(accountId, fmt.Sprintf("%v", time.Now().UnixNano()))
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountCreditTransactionsBetween(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	readCloser := constructTransactionsHttpResponseBody(timestamp.String(then.UnixNano()))

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{Body: readCloser}, nil)
	response, err := c.GetAccountCreditTransactionsBetween(accountId, now.UnixNano(), then.UnixNano())

	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func Test_GetAccountCreditTransactionsBetween_HttpError(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	expectedErr := errors.New("some-error")

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, expectedErr)
	response, err := c.GetAccountCreditTransactionsBetween(accountId, now.UnixNano(), then.UnixNano())

	assert.Error(t, expectedErr, err)
	assert.Nil(t, response)
}

func Test_GetAccountCreditTransactionsBetween_TimestampError(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	expectedErr := errors.New("invalid timestamp provided")
	readCloser := constructTransactionsHttpResponseBody("invalid-timestamp")

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{Body: readCloser}, nil)
	response, err := c.GetAccountCreditTransactionsBetween(accountId, now.UnixNano(), then.UnixNano())

	assert.Error(t, expectedErr, err)
	assert.Nil(t, response)
}

func Test_GetMessagesForTopicBetween(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	readCloser := constructMessagesHttpResponseBody(timestamp.String(now.UnixNano()), topicId.String())

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{Body: readCloser}, nil)
	response, err := c.GetMessagesForTopicBetween(topicId, now.UnixNano(), then.UnixNano())

	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func Test_GetMessagesForTopicBetween_HttpErr(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	expectedErr := errors.New("some-error")

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, expectedErr)
	response, err := c.GetMessagesForTopicBetween(topicId, now.UnixNano(), then.UnixNano())

	assert.Error(t, expectedErr, err)
	assert.Nil(t, response)
}

func Test_GetMessagesForTopicBetween_TimestampErr(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	expectedErr := errors.New("invalid timestamp provided")
	readCloser := constructMessagesHttpResponseBody("invalid-timestamp", topicId.String())

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{Body: readCloser}, nil)
	response, err := c.GetMessagesForTopicBetween(topicId, now.UnixNano(), then.UnixNano())

	assert.Error(t, expectedErr, err)
	assert.Nil(t, response)
}

func Test_GetHBARUsdPrice(t *testing.T) {
	setup()

	encodedResponseBuffer := new(bytes.Buffer)
	encodeErr := json.NewEncoder(encodedResponseBuffer).Encode(testConstants.ParsedTransactionResponse)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	encodedResponseReader := bytes.NewReader(encodedResponseBuffer.Bytes())
	encodedResponseReaderCloser := ioutil.NopCloser(encodedResponseReader)
	response := &http.Response{
		StatusCode: 200,
		Body:       encodedResponseReaderCloser,
	}

	mocks.MHTTPClient.On("Do", mock.Anything).Return(response, error(nil))

	price, err := c.GetHBARUsdPrice()

	assert.Equal(t, testConstants.ParsedTransactionResponseCurrentRate, price)
	assert.Nil(t, err)
}

func constructMessagesHttpResponseBody(consensusTimestamp, topicId string) io.ReadCloser {
	httpResponseBody := &message.Messages{
		Messages: []message.Message{
			{
				ConsensusTimestamp: consensusTimestamp,
				TopicId:            topicId,
			},
		},
	}
	serialized, _ := json.Marshal(httpResponseBody)
	reader := bytes.NewReader(serialized)
	readCloser := ioutil.NopCloser(reader)

	return readCloser
}

func constructTransactionsHttpResponseBody(consensusTimestamp string) io.ReadCloser {
	httpResponseBody := &transaction.Response{
		Transactions: []transaction.Transaction{
			{
				ConsensusTimestamp: consensusTimestamp,
			},
		},
	}
	serialized, _ := json.Marshal(httpResponseBody)
	reader := bytes.NewReader(serialized)
	readCloser := ioutil.NopCloser(reader)

	return readCloser
}
