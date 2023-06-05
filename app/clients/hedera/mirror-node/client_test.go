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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	mirrorAPIAddress = "some-api-address"
	pollingInterval  = 5 * time.Second
	mirrorNodeCfg    = config.MirrorNode{
		ApiAddress:      mirrorAPIAddress,
		PollingInterval: pollingInterval,
	}
	queryDefaultLimit = int64(1)
	queryMaxLimit     = int64(2)
	sequenceNumber    = int64(3)
	logger            = config.GetLoggerFor("Mirror Node Client")

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

	tokenId = hedera.TokenID{
		Shard: 0,
		Realm: 0,
		Token: 4,
	}

	c *Client
)

func setup() {
	mocks.Setup()
	c = &Client{
		mirrorAPIAddress:  mirrorAPIAddress,
		httpClient:        mocks.MHTTPClient,
		pollingInterval:   5 * time.Second,
		queryDefaultLimit: queryDefaultLimit,
		queryMaxLimit:     queryMaxLimit,
		logger:            logger,
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
	response, err := c.GetMessagesAfterTimestamp(topicId, time.Now().UnixNano(), queryDefaultLimit)
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

func Test_GetAccountCreditTransactionsBetween_HttpErr(t *testing.T) {
	setup()
	now := time.Now()
	then := now.Add(time.Hour * 2)
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountCreditTransactionsBetween(accountId, now.UnixNano(), then.UnixNano())
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountCreditTransactionsBetween(t *testing.T) {
	setup()

	expected := transaction.Response{
		Transactions: []transaction.Transaction{
			{
				ConsensusTimestamp: "1631092491.483966000",
				EntityId:           "1",
				Result:             hedera.StatusSuccess.String(),
			},
		},
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)

	response, err := c.GetAccountCreditTransactionsBetween(accountId, 0, time.Now().UnixNano())
	assert.Nil(t, err)
	assert.Equal(t, expected.Transactions[0].ConsensusTimestamp, response[0].ConsensusTimestamp)
}

func Test_QueryDefaultLimit(t *testing.T) {
	setup()

	actual := c.QueryDefaultLimit()

	assert.Equal(t, queryDefaultLimit, actual)
}

func Test_QueryMaxLimit(t *testing.T) {
	setup()

	actual := c.QueryMaxLimit()

	assert.Equal(t, queryMaxLimit, actual)
}

func Test_GetLatestMessages(t *testing.T) {
	setup()
	expectedMsg := message.Message{
		ConsensusTimestamp: "1",
		TopicId:            "1",
		Contents:           "1",
		RunningHash:        "1",
		SequenceNumber:     1,
		ChunkInfo:          nil,
	}
	content := message.Messages{
		Messages: []message.Message{expectedMsg},
	}
	encodedContent, err := httpHelper.EncodeBodyContent(content)
	if err != nil {
		t.Fatal(err)
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)

	response, err := c.GetLatestMessages(topicId, queryDefaultLimit)

	assert.Equal(t, expectedMsg, response[0])
	assert.Len(t, response, 1)
	assert.Nil(t, err)
}

func Test_GetMessageBySequenceNumber(t *testing.T) {
	setup()

	expectedMsg := &message.Message{
		ConsensusTimestamp: "1",
		TopicId:            "1",
		Contents:           "1",
		RunningHash:        "1",
		SequenceNumber:     sequenceNumber,
		ChunkInfo:          nil,
	}
	encodedContent, err := httpHelper.EncodeBodyContent(expectedMsg)
	if err != nil {
		t.Fatal(err)
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)

	response, err := c.GetMessageBySequenceNumber(topicId, sequenceNumber)

	assert.Equal(t, expectedMsg, response)
	assert.Nil(t, err)
}

func Test_GetMessageBySequenceNumber_HttpErr(t *testing.T) {
	setup()
	expectedErr := errors.New("some error")
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, expectedErr)

	response, err := c.GetMessageBySequenceNumber(topicId, sequenceNumber)

	assert.Error(t, err, expectedErr)
	assert.Nil(t, response)
}

func Test_GetMessageBySequenceNumber_JsonErr(t *testing.T) {
	setup()
	expectedErr := json.UnmarshalTypeError{
		Value: "string",
	}
	encodedContent, err := httpHelper.EncodeBodyContent("invalid-content")
	if err != nil {
		t.Fatal(err)
	}
	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)

	response, err := c.GetMessageBySequenceNumber(topicId, sequenceNumber)

	assert.Error(t, err, expectedErr)
	assert.Nil(t, response)
}

func Test_GetHBARUsdPrice(t *testing.T) {
	setup()
	encodedResponseReaderCloser, encodeErr := httpHelper.EncodeBodyContent(testConstants.ParsedTransactionResponse)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	response := &http.Response{
		StatusCode: 200,
		Body:       encodedResponseReaderCloser,
	}

	mocks.MHTTPClient.On("Do", mock.Anything).Return(response, error(nil))

	price, err := c.GetHBARUsdPrice()

	assert.Equal(t, testConstants.ParsedTransactionResponseCurrentRate, price)
	assert.Nil(t, err)
}

func Test_GetTransactionsAfterTimestamp_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetTransactionsAfterTimestamp(accountId, time.Now().UnixNano(), "CryptoTransfer")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountTokenMintTransactionsAfterTimestampString_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountTokenMintTransactionsAfterTimestampString(accountId, "42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountTokenMintTransactionsAfterTimestamp_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountTokenMintTransactionsAfterTimestamp(accountId, int64(42))
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountTokenBurnTransactionsAfterTimestampString_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountTokenBurnTransactionsAfterTimestampString(accountId, "42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccountCreditTransactionsAfterTimestampString_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccountCreditTransactionsAfterTimestampString(accountId, "42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetMessagesForTopicBetween_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetMessagesForTopicBetween(topicId, 0, 0)
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetMessagesForTopicBetween(t *testing.T) {
	setup()
	expectedMsg := message.Message{
		ConsensusTimestamp: "1631092491.483966000",
		TopicId:            "1",
		Contents:           "1",
		RunningHash:        "1",
		SequenceNumber:     1,
		ChunkInfo:          nil,
	}

	content := message.Messages{
		Messages: []message.Message{expectedMsg},
	}

	encodedContent, err := httpHelper.EncodeBodyContent(content)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)

	response, err := c.GetMessagesForTopicBetween(topicId, 0, time.Now().UnixNano())
	assert.Equal(t, expectedMsg, response[0])
	assert.Len(t, response, 1)
	assert.Nil(t, err)
}

func Test_GetNftTransactions_HttpErr(t *testing.T) {
	setup()
	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetNftTransactions("0.0.42", sequenceNumber)
	assert.Error(t, errors.New("some-error"), err)
	assert.Equal(t, transaction.NftTransactionsResponse{}, response)
}

func Test_GetNftTransactions(t *testing.T) {
	setup()

	expected := transaction.NftTransactionsResponse{
		Transactions: []transaction.NftTransaction{
			{
				TransactionID:     "TransactionID",
				Type:              "Type",
				SenderAccountID:   "SenderAccountID",
				ReceiverAccountID: "ReceiverAccountID",
			},
		},
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.GetNftTransactions("0.0.42", sequenceNumber)
	assert.Nil(t, err)
	assert.Equal(t, expected, response)
}

func Test_GetSuccessfulTransaction_HttpErr(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetSuccessfulTransaction("0.0.42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Equal(t, transaction.Transaction{}, response)
}

func Test_GetSuccessfulTransaction(t *testing.T) {
	setup()

	expected := transaction.Response{
		Transactions: []transaction.Transaction{
			{
				ConsensusTimestamp: "1",
				EntityId:           "1",
				Result:             hedera.StatusSuccess.String(),
			},
		},
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.GetSuccessfulTransaction("0.0.42")
	assert.Nil(t, err)
	assert.Equal(t, hedera.StatusSuccess.String(), response.Result)
}

func Test_GetSchedule_HttpErr(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetSchedule("0.0.42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetSchedule(t *testing.T) {
	setup()

	expected := transaction.Schedule{
		ConsensusTimestamp: "1",
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.GetSchedule("0.0.42")
	assert.Nil(t, err)
	assert.Equal(t, expected.ConsensusTimestamp, response.ConsensusTimestamp)
}

func Test_GetNft_HttpErr(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetNft("0.0.42", 42)
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetNft(t *testing.T) {
	setup()

	expected := transaction.Nft{
		CreatedTimestamp: "1",
		SerialNumber:     42,
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.GetNft("0.0.42", 42)
	assert.Nil(t, err)
	assert.Equal(t, expected.SerialNumber, response.SerialNumber)
}

func Test_AccountExists_ShouldNotExists(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response := c.AccountExists(accountId)
	assert.False(t, response)
}

func Test_GetAccount_HttpErr(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetAccount("0.0.42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetAccount(t *testing.T) {
	setup()

	expected := account.AccountsResponse{
		Account: "0.0.42",
		Balance: account.Balance{
			Balance: 42,
		},
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.GetAccount("0.0.42")
	assert.Nil(t, err)
	assert.Equal(t, expected.Account, response.Account)
	assert.Equal(t, expected.Balance.Balance, response.Balance.Balance)
}

func Test_GetToken_HttpErr(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response, err := c.GetToken("0.0.42")
	assert.Error(t, errors.New("some-error"), err)
	assert.Nil(t, response)
}

func Test_GetToken(t *testing.T) {
	setup()

	expected := token.TokenResponse{
		TokenID: "0.0.42",
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.GetToken("0.0.42")
	assert.Nil(t, err)
	assert.Equal(t, expected.TokenID, response.TokenID)
}

func Test_TopicExists_ShouldNotExists(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Get", mock.Anything).Return(nil, errors.New("some-error"))
	response := c.TopicExists(topicId)
	assert.False(t, response)
}

func Test_GetAndParse(t *testing.T) {
	setup()

	expected := transaction.Response{
		Transactions: []transaction.Transaction{
			{
				ConsensusTimestamp: "1",
				EntityId:           "1",
				Result:             hedera.StatusSuccess.String(),
			},
		},
	}

	encodedContent, err := httpHelper.EncodeBodyContent(expected)
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.getAndParse("example.com")
	assert.Nil(t, err)
	assert.Equal(t, hedera.StatusSuccess.String(), response.Transactions[0].Result)
}

func Test_GetAndParseOnJsonError(t *testing.T) {
	setup()

	encodedContent, err := httpHelper.EncodeBodyContent("not-json")
	if err != nil {
		t.Fatal(err)
	}

	mocks.MHTTPClient.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedContent}, nil)
	response, err := c.getAndParse("example.com")
	assert.Error(t, err)
	assert.Nil(t, response)

}
