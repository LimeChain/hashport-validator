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

package token

type (
	TokenResponse struct {
		TokenID           string     `json:"token_id"`
		Name              string     `json:"name"`
		Symbol            string     `json:"symbol"`
		TotalSupply       string     `json:"total_supply"`
		Decimals          string     `json:"decimals"`
		CustomFees        CustomFees `json:"custom_fees"`
		TreasuryAccountId string     `json:"treasury_account_id"`
	}

	NetworkSupplyResponse struct {
		ReleasedSupply string `json:"released_supply"`
		Timestamp      string `json:"timestamp"`
		TotalSupply    string `json:"total_supply"`
	}
)

type CustomFees struct {
	CreatedTimestamp string       `json:"created_timestamp"`
	RoyaltyFees      []RoyaltyFee `json:"royalty_fees"`
}

type FixedFee struct {
	Amount              int64   `json:"amount"`
	DenominatingTokenId *string `json:"denominating_token_id"`
}

type RoyaltyFee struct {
	Amount             Fraction `json:"amount"`
	FallbackFee        FixedFee `json:"fallback_fee"`
	CollectorAccountID string   `json:"collector_account_id"`
}

type Fraction struct {
	Numerator   int64 `json:"numerator"`
	Denominator int64 `json:"denominator"`
}
