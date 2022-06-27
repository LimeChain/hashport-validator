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
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	mirrorNodeHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/mirror-node"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	mirrorNodeModel "github.com/limechain/hedera-eth-bridge-validator/app/model/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var (
	GetHbarPriceHeaders         = map[string]string{"Accepts": "application/json"}
	TransactionsGetHBARUsdPrice = "transactions?account.id=0.0.57&transactiontype=fileupdate&limit=1"
)

type Client struct {
	mirrorAPIAddress             string
	httpClient                   client.HttpClient
	pollingInterval              time.Duration
	queryMaxLimit                int64
	queryDefaultLimit            int64
	fullHederaGetHbarUsdPriceUrl string
	logger                       *log.Entry
}

func NewClient(mirrorNode config.MirrorNode) *Client {
	return &Client{
		mirrorAPIAddress:             mirrorNode.ApiAddress,
		pollingInterval:              mirrorNode.PollingInterval,
		queryMaxLimit:                mirrorNode.QueryMaxLimit,
		queryDefaultLimit:            mirrorNode.QueryDefaultLimit,
		fullHederaGetHbarUsdPriceUrl: strings.Join([]string{mirrorNode.ApiAddress, TransactionsGetHBARUsdPrice}, ""),
		httpClient:                   new(http.Client),
		logger:                       config.GetLoggerFor("Mirror Node Client"),
	}
}

func (c *Client) GetHBARUsdPrice() (price decimal.Decimal, err error) {
	var parsedResponse mirrorNodeModel.TransactionsResponse
	err = httpHelper.Get(c.httpClient, c.fullHederaGetHbarUsdPriceUrl, GetHbarPriceHeaders, &parsedResponse, c.logger)
	if err != nil {
		return decimal.Decimal{}, err
	}

	hederaFileRate, err := mirrorNodeHelper.GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedResponse, c.logger)
	if err == nil {
		price = hederaFileRate.CurrentRate
	}

	return price, err
}

func (c Client) GetAccountTokenMintTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&scheduled=true&type=credit&timestamp=gt:%s&order=asc&transactiontype=tokenmint",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountTokenMintTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*transaction.Response, error) {
	return c.GetAccountTokenMintTransactionsAfterTimestampString(accountId, timestampHelper.String(from))
}

func (c Client) GetAccountTokenBurnTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&scheduled=true&timestamp=gt:%s&order=asc&transactiontype=tokenburn",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountTokenBurnTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*transaction.Response, error) {
	return c.GetAccountTokenBurnTransactionsAfterTimestampString(accountId, timestampHelper.String(from))
}

func (c Client) GetAccountDebitTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=debit&timestamp=gt:%s&order=asc&transactiontype=cryptotransfer",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountCreditTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc&transactiontype=cryptotransfer",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*transaction.Response, error) {
	return c.GetAccountCreditTransactionsAfterTimestampString(accountId, timestampHelper.String(from))
}

// GetAccountCreditTransactionsBetween returns all incoming Transfers for the specified account between timestamp `from` and `to` excluded
func (c Client) GetAccountCreditTransactionsBetween(accountId hedera.AccountID, from, to int64) ([]transaction.Transaction, error) {
	transactions, err := c.GetAccountCreditTransactionsAfterTimestamp(accountId, from)
	if err != nil {
		return nil, err
	}

	var res []transaction.Transaction
	for _, t := range transactions.Transactions {
		ts, err := timestampHelper.FromString(t.ConsensusTimestamp)
		if err != nil {
			return nil, err
		}
		if ts < to {
			res = append(res, t)
		}
	}
	return res, nil
}

// GetMessagesAfterTimestamp returns all Topic messages after the given timestamp
func (c Client) GetMessagesAfterTimestamp(topicId hedera.TopicID, from int64, limit int64) ([]message.Message, error) {
	messagesQuery := fmt.Sprintf("/%s/messages?timestamp=gt:%s&limit=%d",
		topicId.String(),
		timestampHelper.String(from),
		limit)

	return c.getTopicMessagesByQuery(messagesQuery)
}

// GetMessageBySequenceNumber returns message from given topic with provided sequence number
func (c Client) GetMessageBySequenceNumber(topicId hedera.TopicID, sequenceNumber int64) (*message.Message, error) {
	messagesQuery := fmt.Sprintf("%s%s/%s/messages/%d",
		c.mirrorAPIAddress,
		"topics",
		topicId.String(),
		sequenceNumber)

	response, e := c.get(messagesQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(response)
	if e != nil {
		return nil, e
	}

	var message *message.Message
	e = json.Unmarshal(bodyBytes, &message)
	if e != nil {
		return nil, e
	}

	return message, nil
}

// GetLatestMessages returns latest Topic messages
func (c Client) GetLatestMessages(topicId hedera.TopicID, limit int64) ([]message.Message, error) {
	latestMessagesQuery := fmt.Sprintf("/%s/messages?order=desc&limit=%d", topicId.String(), limit)
	return c.getTopicMessagesByQuery(latestMessagesQuery)
}

// GetQueryDefaultLimit returns the default records limit per query
func (c Client) QueryDefaultLimit() int64 {
	return c.queryDefaultLimit
}

// GetQueryMaxLimit returns the maximum allowed limit per messages query
func (c Client) QueryMaxLimit() int64 {
	return c.queryMaxLimit
}

// GetMessagesForTopicBetween returns all Topic messages for the specified topic between timestamp `from` and `to` excluded
func (c Client) GetMessagesForTopicBetween(topicId hedera.TopicID, from, to int64) ([]message.Message, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s/messages?timestamp=gt:%s",
		topicId.String(),
		timestampHelper.String(from))
	msgs, err := c.getTopicMessagesByQuery(transactionsDownloadQuery)
	if err != nil {
		return nil, err
	}

	// TODO refactor into 1 function (reuse code above)
	var res []message.Message
	for _, m := range msgs {
		ts, err := timestampHelper.FromString(m.ConsensusTimestamp)
		if err != nil {
			return nil, err
		}
		if ts < to {
			res = append(res, m)
		}
	}
	return res, nil
}

// GetNftTransactions returns the nft transactions for tokenID and serialNum
func (c Client) GetNftTransactions(tokenID string, serialNum int64) (transaction.NftTransactionsResponse, error) {
	query := fmt.Sprintf("%stokens/%s/nfts/%d/transactions", c.mirrorAPIAddress, tokenID, serialNum)

	httpResponse, err := c.get(query)
	if err != nil {
		return transaction.NftTransactionsResponse{}, err
	}

	bodyBytes, err := readResponseBody(httpResponse)
	if err != nil {
		return transaction.NftTransactionsResponse{}, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return transaction.NftTransactionsResponse{}, errors.New(fmt.Sprintf("Mirror Node API [%s] ended with Status Code [%d]. Body bytes: [%s]", query, httpResponse.StatusCode, bodyBytes))
	}

	var response *transaction.NftTransactionsResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return transaction.NftTransactionsResponse{}, err
	}

	return *response, nil
}

func (c Client) GetTransaction(transactionID string) (*transaction.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s",
		transactionID)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetSuccessfulTransaction(transactionID string) (transaction.Transaction, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s",
		transactionID)
	response, err := c.getTransactionsByQuery(transactionsDownloadQuery)
	if err != nil {
		return transaction.Transaction{}, err
	}
	txs := response.Transactions
	for _, tx := range txs {
		if tx.Result == hedera.StatusSuccess.String() {
			return tx, nil
		}
	}

	return transaction.Transaction{}, errors.New(fmt.Sprintf("[%s] - No SUCCESS transaction found", transactionID))
}

// GetScheduledTransaction gets the Scheduled transaction of an executed transaction
func (c Client) GetScheduledTransaction(transactionID string) (*transaction.Response, error) {
	return c.GetTransaction(fmt.Sprintf("%s?scheduled=false", transactionID))
}

// GetSchedule retrieves a schedule entity by its id
func (c Client) GetSchedule(scheduleID string) (*transaction.Schedule, error) {
	query := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "schedules/", scheduleID)

	httpResponse, e := c.get(query)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, query))
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *transaction.Schedule
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

func (c Client) GetStateProof(transactionID string) ([]byte, error) {
	query := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions",
		fmt.Sprintf("/%s/stateproof", transactionID))

	response, e := c.get(query)
	if e != nil {
		return nil, e
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("State Proof HTTP GET for TransactionID [%s] ended with Status Code [%d].", transactionID, response.StatusCode))
	}

	return readResponseBody(response)
}

func (c Client) GetNft(tokenID string, serialNum int64) (*transaction.Nft, error) {
	nftQuery := fmt.Sprintf("%s%d", "/nfts/", serialNum)
	query := fmt.Sprintf("%s%s%s%s", c.mirrorAPIAddress, "tokens/", tokenID, nftQuery)

	httpResponse, e := c.get(query)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, query))
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *transaction.Nft
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

func (c Client) AccountExists(accountID hedera.AccountID) bool {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "accounts")
	accountQuery := fmt.Sprintf("%s/%s",
		mirrorNodeApiTransactionAddress,
		accountID.String())

	return c.query(accountQuery, accountID.String())
}

// GetAccount retrieves an account entity by its id
func (c Client) GetAccount(accountID string) (*account.AccountsResponse, error) {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "accounts")
	query := fmt.Sprintf("%s/%s",
		mirrorNodeApiTransactionAddress,
		accountID)

	httpResponse, e := c.get(query)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, query))
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *account.AccountsResponse
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

// GetToken retrieves a token entity by its id
func (c Client) GetToken(tokenID string) (*token.TokenResponse, error) {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "tokens")
	query := fmt.Sprintf("%s/%s",
		mirrorNodeApiTransactionAddress,
		tokenID)

	httpResponse, e := c.get(query)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, query))
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *token.TokenResponse
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

func (c Client) TopicExists(topicID hedera.TopicID) bool {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "topics")
	topicQuery := fmt.Sprintf("%s/%s/messages",
		mirrorNodeApiTransactionAddress,
		topicID.String())

	return c.query(topicQuery, topicID.String())
}

func (c *Client) GetTransactionsAfterTimestamp(accountId hedera.AccountID, startTimestamp int64, transactionType string) ([]transaction.Transaction, error) {
	query := fmt.Sprintf("?account_id=%s&transactionType=%s&timestamp=gte:%s&limit=%d",
		accountId,
		timestampHelper.String(startTimestamp),
		transactionType,
		c.queryDefaultLimit)

	resp, err := c.getTransactionsByQuery(query)
	if err != nil {
		return nil, err
	}

	return resp.Transactions, nil
}

func (c Client) query(query, entityID string) bool {
	response, err := c.httpClient.Get(query)
	if err != nil {
		c.logger.Errorf("[%s] - failed to query account. Error [%s].", entityID, err)
		return false
	}

	body, err := readResponseBody(response)
	if err != nil {
		c.logger.Errorf("[%s] - failed to read response body. Error [%s].", entityID, err)
		return false
	}

	if response.StatusCode != 200 {
		c.logger.Errorf("[%s] - query ended with [%d]. Response body: [%s]. ", entityID, response.StatusCode, body)
		return false
	}

	return true
}

// WaitForTransaction Polls the transaction at intervals. Depending on the
// result, the corresponding `onSuccess` and `onFailure` functions are called
func (c Client) WaitForTransaction(txId string, onSuccess, onFailure func()) {
	go func() {
		for {
			response, err := c.GetTransaction(txId)
			if response != nil && response.IsNotFound() {
				continue
			}
			if err != nil {
				c.logger.Errorf("[%s] Error while trying to get tx. Error: [%s].", txId, err.Error())
				return
			}

			if len(response.Transactions) > 0 {
				success := false
				for _, transaction := range response.Transactions {
					if transaction.Result == hedera.StatusSuccess.String() {
						success = true
						break
					}
				}

				if success {
					c.logger.Debugf("TX [%s] was successfully mined", txId)
					onSuccess()
				} else {
					c.logger.Debugf("TX [%s] has failed", txId)
					onFailure()
				}
				return
			}
			c.logger.Tracef("Pinged Mirror Node for TX [%s]. No update", txId)
			time.Sleep(c.pollingInterval * time.Second)
		}
	}()
	c.logger.Debugf("Added new TX [%s] for monitoring", txId)
}

// WaitForScheduledTransaction Polls the transaction at intervals. Depending on the
// result, the corresponding `onSuccess` and `onFailure` functions are called
func (c Client) WaitForScheduledTransaction(txId string, onSuccess, onFailure func()) {
	c.logger.Debugf("Added new Scheduled TX [%s] for monitoring", txId)
	for {
		response, err := c.GetTransaction(txId)
		if response != nil && response.IsNotFound() {
			continue
		}
		if err != nil {
			c.logger.Errorf("[%s] Error while trying to get tx. Error: [%s].", txId, err)
			return
		}

		if len(response.Transactions) > 1 {
			success := false
			for _, transaction := range response.Transactions {
				if transaction.Scheduled && transaction.Result == hedera.StatusSuccess.String() {
					success = true
					break
				}
			}

			if success {
				c.logger.Debugf("Scheduled TX [%s] was successfully mined", txId)
				onSuccess()
			} else {
				c.logger.Debugf("Scheduled TX [%s] has failed", txId)
				onFailure()
			}
			return
		}
		c.logger.Tracef("Pinged Mirror Node for Scheduled TX [%s]. No update", txId)
		time.Sleep(c.pollingInterval * time.Second)
	}
}

func (c Client) get(query string) (*http.Response, error) {
	return c.httpClient.Get(query)
}

func (c Client) getTransactionsByQuery(query string) (*transaction.Response, error) {
	transactionsQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions", query)

	return c.getAndParse(transactionsQuery)
}

func (c Client) getAndParse(query string) (*transaction.Response, error) {
	httpResponse, e := c.get(query)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *transaction.Response
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return response, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, response.Status.String()))
	}

	return response, nil
}

func (c Client) getTopicMessagesByQuery(query string) ([]message.Message, error) {
	messagesQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "topics", query)
	response, e := c.get(messagesQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(response)
	if e != nil {
		return nil, e
	}
	var messages *message.Messages
	e = json.Unmarshal(bodyBytes, &messages)
	if e != nil {
		return nil, e
	}
	return messages.Messages, nil
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
