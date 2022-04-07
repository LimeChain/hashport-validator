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

package client

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/shopspring/decimal"
)

type MirrorNode interface {
	// GetAccountTokenMintTransactionsAfterTimestampString queries the hedera mirror node for transactions on a certain account with type TokenMint
	GetAccountTokenMintTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error)
	// GetAccountTokenMintTransactionsAfterTimestamp queries the hedera mirror node for transactions on a certain account with type TokenMint
	GetAccountTokenMintTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*transaction.Response, error)
	// GetAccountTokenBurnTransactionsAfterTimestampString queries the hedera mirror node for transactions on a certain account with type TokenBurn
	GetAccountTokenBurnTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error)
	// GetAccountTokenBurnTransactionsAfterTimestamp queries the hedera mirror node for transactions on a certain account with type TokenBurn
	GetAccountTokenBurnTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*transaction.Response, error)
	// GetAccountDebitTransactionsAfterTimestampString queries the hedera mirror node for transactions that are debit and after a given timestamp
	GetAccountDebitTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error)
	// GetAccountCreditTransactionsAfterTimestampString returns all transaction after a given timestamp
	GetAccountCreditTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*transaction.Response, error)
	// GetAccountCreditTransactionsAfterTimestamp returns all transaction after a given timestamp
	GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*transaction.Response, error)
	// GetAccountCreditTransactionsBetween returns all incoming Transfers for the specified account between timestamp `from` and `to` excluded
	GetAccountCreditTransactionsBetween(accountId hedera.AccountID, from, to int64) ([]transaction.Transaction, error)
	// GetMessagesAfterTimestamp returns all topic messages after the given timestamp
	GetMessagesAfterTimestamp(topicId hedera.TopicID, from int64) ([]message.Message, error)
	// GetMessagesForTopicBetween returns all topic messages for a given topic between timestamp `from` included and `to` excluded
	GetMessagesForTopicBetween(topicId hedera.TopicID, from, to int64) ([]message.Message, error)
	// GetNftTransactions returns the nft transactions for tokenID and serialNum
	GetNftTransactions(tokenID string, serialNum int64) (transaction.NftTransactionsResponse, error)
	// GetScheduledTransaction gets the Scheduled transaction of an executed transaction
	GetScheduledTransaction(transactionID string) (*transaction.Response, error)
	// GetTransaction gets all data related to a specific transaction id or returns an error
	GetTransaction(transactionID string) (*transaction.Response, error)
	// GetSuccessfulTransaction gets the success transaction by transaction id or returns an error
	GetSuccessfulTransaction(transactionID string) (transaction.Transaction, error)
	// GetSchedule retrieves a schedule entity by its id
	GetSchedule(scheduleID string) (*transaction.Schedule, error)
	// GetStateProof sends a query to get the state proof. If the query is successful, the function returns the state.
	// If the query returns a status != 200, the function returns an error.
	GetStateProof(transactionID string) ([]byte, error)
	// GetNft retrieves an nft token entity by its id and serial number
	GetNft(tokenID string, serialNum int64) (*transaction.Nft, error)
	// AccountExists sends a query to check whether a specific account exists. If the query returns a status != 200, the function returns a false value
	AccountExists(accountID hedera.AccountID) bool
	// GetAccount gets the account data by ID.
	GetAccount(accountID string) (*account.AccountsResponse, error)
	// GetToken gets the token data by ID.
	GetToken(tokenID string) (*token.TokenResponse, error)
	// TopicExists sends a query to check whether a specific topic exists. If the query returns a status != 200, the function returns a false value
	TopicExists(topicID hedera.TopicID) bool
	// WaitForTransaction Polls the transaction at intervals. Depending on the
	// result, the corresponding `onSuccess` and `onFailure` functions are called
	WaitForTransaction(txId string, onSuccess, onFailure func())
	// WaitForScheduledTransaction Polls the transaction at intervals. Depending on the
	// result, the corresponding `onSuccess` and `onFailure` functions are called
	WaitForScheduledTransaction(txId string, onSuccess, onFailure func())
	// GetHBARUsdPrice Returns USD price for HBAR
	GetHBARUsdPrice() (price decimal.Decimal, err error)
}
