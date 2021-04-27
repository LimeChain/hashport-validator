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

package util

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"testing"
)

func GetHederaAccountBalance(client *hedera.Client, account hedera.AccountID, t *testing.T) hedera.AccountBalance {
	// Get bridge account hbar balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(account).
		Execute(client)
	if err != nil {
		t.Fatalf("Unable to query the balance of the account [%s], Error: [%s]", account.String(), err)
	}
	return receiverBalance
}

func GetMembersAccountBalances(client *hedera.Client, members []hedera.AccountID, t *testing.T) []hedera.AccountBalance {
	var balances []hedera.AccountBalance
	for _, member := range members {
		// Get bridge account hbar balance before transfer
		balance, err := hedera.NewAccountBalanceQuery().
			SetAccountID(member).
			Execute(client)
		if err != nil {
			t.Fatalf("Unable to query the balance of the account [%s], Error: [%s]", member, err)
		}

		balances = append(balances, balance)
	}

	return balances
}
