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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"io/ioutil"
	"net/http"
)

type MirrorNode struct {
	mirrorAPIAddress string
	httpClient       *http.Client
}

func NewMirrorNodeClient(mirrorNodeAPIAddress string) *MirrorNode {
	return &MirrorNode{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		httpClient:       &http.Client{},
	}
}

func (c MirrorNode) GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*Transactions, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		accountId.String(),
		timestampHelper.ToString(from))
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

// GetMessagesForTopicBetween returns all Topic messages for the specified topic between timestamp `from` and `to` excluded
func (c MirrorNode) GetAccountCreditTransactionsBetween(accountId hedera.AccountID, from, to int64) ([]Transaction, error) {
	transactions, err := c.GetAccountCreditTransactionsAfterTimestamp(accountId, from)
	if err != nil {
		return nil, err
	}

	var res []Transaction
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

// GetMessagesForTopicBetween returns all Topic messages for the specified topic between timestamp `from` and `to` excluded
func (c MirrorNode) GetMessagesForTopicBetween(topicId hedera.TopicID, from, to int64) ([]Message, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s/messages?timestamp=gt:%s",
		topicId.String(),
		timestampHelper.ToString(from))
	msgs, err := c.getTopicMessagesByQuery(transactionsDownloadQuery)
	if err != nil {
		return nil, err
	}

	// TODO refactor into 1 function (reuse code above)
	var res []Message
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

func (c MirrorNode) GetAccountTransaction(transactionID string) (*Transactions, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s",
		transactionID)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c MirrorNode) GetStateProof(transactionID string) ([]byte, error) {
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

func (c MirrorNode) get(query string) (*http.Response, error) {
	return c.httpClient.Get(query)
}

func (c MirrorNode) getTransactionsByQuery(query string) (*Transactions, error) {
	transactionsQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions", query)
	response, e := c.get(transactionsQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(response)
	if e != nil {
		return nil, e
	}

	var transactions *Transactions
	e = json.Unmarshal(bodyBytes, &transactions)
	if e != nil {
		return nil, e
	}

	return transactions, nil
}

func (c MirrorNode) getTopicMessagesByQuery(query string) ([]Message, error) {
	messagesQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "topics", query)
	response, e := c.get(messagesQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(response)
	if e != nil {
		return nil, e
	}

	var messages *Messages
	e = json.Unmarshal(bodyBytes, &messages)
	if e != nil {
		return nil, e
	}
	return messages.Messages, nil
}

func (c MirrorNode) AccountExists(accountID hedera.AccountID) bool {
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

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
