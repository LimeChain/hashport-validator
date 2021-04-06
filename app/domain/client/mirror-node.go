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

package client

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
)

type MirrorNode interface {
	GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*mirror_node.Response, error)
	// GetAccountCreditTransactionsBetween returns all incoming Transfers for the specified account between timestamp `from` and `to` excluded
	GetAccountCreditTransactionsBetween(accountId hedera.AccountID, from, to int64) ([]mirror_node.Transaction, error)
	GetMessagesAfterTimestamp(topicId hedera.TopicID, from int64) ([]mirror_node.Message, error)
	// GetMessagesForTopicBetween returns all topic messages for a given topic between timestamp `from` included and `to` excluded
	GetMessagesForTopicBetween(topicId hedera.TopicID, from, to int64) ([]mirror_node.Message, error)
	GetAccountTransaction(transactionID string) (*mirror_node.Response, error)
	GetStateProof(transactionID string) ([]byte, error)
	AccountExists(accountID hedera.AccountID) bool
	// WaitForTransaction Polls the transaction at intervals. Depending on the
	// result, the corresponding `onSuccess` and `onFailure` functions are called
	WaitForTransaction(txId string, onSuccess, onFailure func())
}
