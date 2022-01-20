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

package model

type (
	// AccountsResponse struct used by the Hedera Mirror node REST API to return information
	// regarding a given Account
	AccountsResponse struct {
		Account string  `json:"account"`
		Balance Balance `json:"balance"`
	}
	// Balance struct used by the Hedera Mirror node REST API to return information
	// regarding a given Account
	Balance struct {
		Balance   int     `json:"balance"`
		Timestamp string  `json:"timestamp"`
		Tokens    []Token `json:"tokens"`
	}
	// Tokens struct used by the Hedera Mirror node REST API to return information
	// regarding a given Account's tokens
	Token struct {
		TokenID string `json:"token_id"`
		Balance int    `json:"balance"`
	}
)
