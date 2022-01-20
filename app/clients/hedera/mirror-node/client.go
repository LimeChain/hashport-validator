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

package mirror_node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	mirrorAPIAddress string
	httpClient       client.HttpClient
	pollingInterval  time.Duration
	logger           *log.Entry
}

func NewClient(mirrorNodeAPIAddress string, pollingInterval time.Duration) *Client {
	httpC := &http.Client{}
	return &Client{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		pollingInterval:  pollingInterval,
		httpClient:       httpC,
		logger:           config.GetLoggerFor("Mirror Node Client"),
	}
}

func (c Client) GetAccountTokenMintTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&scheduled=true&type=credit&timestamp=gt:%s&order=asc&transactiontype=tokenmint",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountTokenMintTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*model.Response, error) {
	return c.GetAccountTokenMintTransactionsAfterTimestampString(accountId, timestampHelper.String(from))
}

func (c Client) GetAccountTokenBurnTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&scheduled=true&timestamp=gt:%s&order=asc&transactiontype=tokenburn",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountTokenBurnTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*model.Response, error) {
	return c.GetAccountTokenBurnTransactionsAfterTimestampString(accountId, timestampHelper.String(from))
}

func (c Client) GetAccountDebitTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=debit&timestamp=gt:%s&order=asc&transactiontype=cryptotransfer",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountCreditTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc&transactiontype=cryptotransfer",
		accountId.String(),
		from)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c Client) GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*model.Response, error) {
	return c.GetAccountCreditTransactionsAfterTimestampString(accountId, timestampHelper.String(from))
}

// GetAccountCreditTransactionsBetween returns all incoming Transfers for the specified account between timestamp `from` and `to` excluded
func (c Client) GetAccountCreditTransactionsBetween(accountId hedera.AccountID, from, to int64) ([]model.Transaction, error) {
	transactions, err := c.GetAccountCreditTransactionsAfterTimestamp(accountId, from)
	if err != nil {
		return nil, err
	}

	var res []model.Transaction
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
func (c Client) GetMessagesAfterTimestamp(topicId hedera.TopicID, from int64) ([]model.Message, error) {
	messagesQuery := fmt.Sprintf("/%s/messages?timestamp=gt:%s",
		topicId.String(),
		timestampHelper.String(from))

	return c.getTopicMessagesByQuery(messagesQuery)
}

// GetMessagesForTopicBetween returns all Topic messages for the specified topic between timestamp `from` and `to` excluded
func (c Client) GetMessagesForTopicBetween(topicId hedera.TopicID, from, to int64) ([]model.Message, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s/messages?timestamp=gt:%s",
		topicId.String(),
		timestampHelper.String(from))
	msgs, err := c.getTopicMessagesByQuery(transactionsDownloadQuery)
	if err != nil {
		return nil, err
	}

	// TODO refactor into 1 function (reuse code above)
	var res []model.Message
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

func (c Client) GetTransaction(transactionID string) (*model.Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s",
		transactionID)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

// GetScheduledTransaction gets the Scheduled transaction of an executed transaction
func (c Client) GetScheduledTransaction(transactionID string) (*model.Response, error) {
	return c.GetTransaction(fmt.Sprintf("%s?scheduled=false", transactionID))
}

// GetSchedule retrieves a schedule entity by its id
func (c Client) GetSchedule(scheduleID string) (*model.Schedule, error) {
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

	var response *model.Schedule
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

func (c Client) AccountExists(accountID hedera.AccountID) bool {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "accounts")
	accountQuery := fmt.Sprintf("%s/%s",
		mirrorNodeApiTransactionAddress,
		accountID.String())
	response, e := c.httpClient.Get(accountQuery)
	if e != nil {
		return false
	}

	if response.StatusCode != 200 {
		return false
	}

	return true
}

// GetAccount retrieves an account entity by its id
func (c Client) GetAccount(accountID string) (*model.AccountsResponse, error) {
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

	var response *model.AccountsResponse
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

// GetToken retrieves a token entity by its id
func (c Client) GetToken(tokenID string) (*model.TokenResponse, error) {
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

	var response *model.TokenResponse
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

// GetNetworkSupply retrieves the Hedera network supply of HBAR
func (c Client) GetNetworkSupply() (*model.NetworkSupplyResponse, error) {
	query := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "network/supply")

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

	var response *model.NetworkSupplyResponse
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}

	return response, nil
}

func (c Client) TopicExists(topicID hedera.TopicID) bool {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "topics")
	accountQuery := fmt.Sprintf("%s/%s/messages",
		mirrorNodeApiTransactionAddress,
		topicID.String())
	response, e := c.httpClient.Get(accountQuery)
	if e != nil {
		return false
	}

	if response.StatusCode != 200 {
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

func (c Client) getTransactionsByQuery(query string) (*model.Response, error) {
	transactionsQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions", query)

	return c.getAndParse(transactionsQuery)
}

func (c Client) getAndParse(query string) (*model.Response, error) {
	httpResponse, e := c.get(query)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *model.Response
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return response, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, response.Status.String()))
	}

	return response, nil
}

func (c Client) getTopicMessagesByQuery(query string) ([]model.Message, error) {
	messagesQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "topics", query)
	response, e := c.get(messagesQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(response)
	if e != nil {
		return nil, e
	}

	var messages *model.Messages
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
