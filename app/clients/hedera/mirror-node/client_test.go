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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var (
	mirrorNodeAPIAddress       = "https://testnet.mirrornode.hedera.com/api/v1/"
	pollingInterval, _         = time.ParseDuration("1m3s")
	client                     = NewClient(mirrorNodeAPIAddress, pollingInterval)
	accountId, _               = hedera.AccountIDFromString("0.0.1547516")
	topicId, _                 = hedera.TopicIDFromString("0.0.1870757")
	transactionId              = "0.0.263546-1622538877-145421413"
	from                 int64 = 121593059375739000
	to                   int64 = 2622614200778494000
	onExecution                = "Pending"
	wg                   sync.WaitGroup
)

func TestNewClient(t *testing.T) {
	assert.Equal(t, mirrorNodeAPIAddress, client.mirrorAPIAddress)
	assert.Equal(t, pollingInterval, client.pollingInterval)
}

func TestGetAccountCreditTransactionsAfterTimestamp(t *testing.T) {
	res, err := client.GetAccountCreditTransactionsAfterTimestamp(accountId, from)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	expectedTransactions := 45
	assert.Equal(t, expectedTransactions, len(res.Transactions))
}

func TestGetAccountCreditTransactionsBetween(t *testing.T) {
	res, err := client.GetAccountCreditTransactionsBetween(accountId, from, to)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestGetMessagesAfterTimestamp(t *testing.T) {
	res, err := client.GetMessagesAfterTimestamp(topicId, from)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestGetMessagesForTopicBetween(t *testing.T) {
	res, err := client.GetMessagesForTopicBetween(topicId, from, to)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestGetTransaction(t *testing.T) {
	res, err := client.GetTransaction(transactionId)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestGetStateProof(t *testing.T) {
	res, err := client.GetStateProof(transactionId)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestAccountExists(t *testing.T) {
	exists := client.AccountExists(accountId)
	assert.True(t, exists)
}

func TestTopicExists(t *testing.T) {
	exists := client.TopicExists(topicId)
	assert.True(t, exists)
}

func TestWaitForTransaction(t *testing.T) {
	wg.Add(1)
	client.WaitForTransaction(transactionId, onSuccess, onFailure)
	wg.Wait()
	assert.Equal(t, "Success", onExecution)
}

func TestWaitForScheduledTransferTransaction(t *testing.T) {
	wg.Add(1)
	client.WaitForScheduledTransferTransaction("0.0.1547516-1622187519-324550837", onSuccess, onFailure)
	wg.Wait()
	assert.Equal(t, "Success", onExecution)
}

func onSuccess() {
	onExecution = "Success"
	defer wg.Done()
}

func onFailure() {
	onExecution = "Failure"
	defer wg.Done()
}
