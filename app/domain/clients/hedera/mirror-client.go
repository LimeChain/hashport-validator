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

package clients

import (
	"net/http"

	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
)

type HederaMirrorClient interface {
	GetSuccessfulAccountCreditTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp int64) (*transaction.HederaTransactions, error)
	GetAccountTransaction(transactionID string) (*transaction.HederaTransactions, error)
	GetStateProof(transactionID string) ([]byte, error)
	Get(query string) (*http.Response, error)
	GetTransactionsByQuery(query string) (*transaction.HederaTransactions, error)
	AccountExists(accountID hedera.AccountID) bool
}
