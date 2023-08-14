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

package account

import "github.com/limechain/hedera-eth-bridge-validator/constants"

type (
	// AccountsResponse struct used by the Hedera Mirror node REST API to return information
	// regarding a given Account
	AccountsResponse struct {
		Account string  `json:"account"`
		Balance Balance `json:"balance"`
	}

	// AccountsQueryResponse struct used by the Hedera Mirror node REST API to return information regarding a given Account
	AccountsQueryResponse struct {
		Accounts []struct {
			Account string `json:"account"`
		} `json:"accounts"`
	}

	// Balance struct used by the Hedera Mirror node REST API to return information
	// regarding a given Account
	Balance struct {
		Balance   int            `json:"balance"`
		Timestamp string         `json:"timestamp"`
		Tokens    []AccountToken `json:"tokens"`
	}
	// AccountToken struct used by the Hedera Mirror node REST API to return information
	// regarding a given Account's tokens
	AccountToken struct {
		TokenID string `json:"token_id"`
		Balance int    `json:"balance"`
	}
)

func (b *Balance) GetAccountTokenBalancesByAddress() map[string]int {
	hederaTokenBalancesByAddress := make(map[string]int)
	for _, token := range b.Tokens {
		hederaTokenBalancesByAddress[token.TokenID] = token.Balance
	}
	hederaTokenBalancesByAddress[constants.Hbar] = b.Balance

	return hederaTokenBalancesByAddress
}
