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
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"io/ioutil"
	"net/http"
)

type HederaMirrorClient struct {
	mirrorAPIAddress string
	httpClient       *http.Client
}

func NewHederaMirrorClient(mirrorNodeAPIAddress string) *HederaMirrorClient {
	return &HederaMirrorClient{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		httpClient:       &http.Client{},
	}
}

func (c HederaMirrorClient) GetSuccessfulAccountCreditTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp int64) (*transaction.HederaTransactions, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		accountId.String(),
		timestampHelper.ToString(milestoneTimestamp))
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c HederaMirrorClient) GetAccountTransaction(transactionID string) (*transaction.HederaTransactions, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s",
		transactionID)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c HederaMirrorClient) GetStateProof(transactionID string) ([]byte, error) {
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

func (c HederaMirrorClient) get(query string) (*http.Response, error) {
	return c.httpClient.Get(query)
}

func (c HederaMirrorClient) getTransactionsByQuery(query string) (*transaction.HederaTransactions, error) {
	transactionsQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions", query)

	response, e := c.get(transactionsQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(response)
	if e != nil {
		return nil, e
	}

	var transactions *transaction.HederaTransactions
	e = json.Unmarshal(bodyBytes, &transactions)
	if e != nil {
		return nil, e
	}

	return transactions, nil
}

func (c HederaMirrorClient) AccountExists(accountID hedera.AccountID) bool {
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
