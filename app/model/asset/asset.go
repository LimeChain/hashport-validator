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

package asset

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"
	"math/big"

	"github.com/shopspring/decimal"
)

type NativeAsset struct {
	MinFeeAmountInUsd *decimal.Decimal
	ChainId           uint64
	Asset             string
	FeePercentage     int64
}

type FungibleAssetInfo struct {
	Name          string   `json:"name"`
	Symbol        string   `json:"symbol"`
	Decimals      uint8    `json:"decimals"`
	IsNative      bool     `json:"isNative"`
	ReserveAmount *big.Int `json:"-"`
}
type CustomFeeTotalAmounts struct {
	FallbackFeeAmountInHbar     int64            `json:"-"`
	FallbackFeeAmountsByTokenId map[string]int64 `json:"-"`
}

type NonFungibleAssetInfo struct {
	Name                  string                `json:"name"`
	Symbol                string                `json:"symbol"`
	IsNative              bool                  `json:"isNative"`
	ReserveAmount         *big.Int              `json:"-"`
	CustomFees            CustomFees            `json:"customFees"`
	CustomFeeTotalAmounts CustomFeeTotalAmounts `json:"-"`
	TreasuryAccountId     string                `json:"-"`
}

type CustomFees struct {
	CreatedTimestamp string       `json:"createdTimestamp"`
	RoyaltyFees      []RoyaltyFee `json:"royaltyFees"`
}

func (c *CustomFees) InitFromResponse(customFees token.CustomFees) {
	c.CreatedTimestamp = customFees.CreatedTimestamp
	c.RoyaltyFees = make([]RoyaltyFee, len(customFees.RoyaltyFees))
	for i, fee := range customFees.RoyaltyFees {
		c.RoyaltyFees[i].InitFromResponse(fee)
	}
}

type FixedFee struct {
	Amount              int64   `json:"amount"`
	DenominatingTokenId *string `json:"denominatingTokenId"`
}

func (f *FixedFee) InitFromResponse(fee token.FixedFee) {
	f.Amount = fee.Amount
	f.DenominatingTokenId = fee.DenominatingTokenId
}

type RoyaltyFee struct {
	Amount             token.Fraction `json:"amount"`
	FallbackFee        FixedFee       `json:"fallbackFee"`
	CollectorAccountID string         `json:"collectorAccountId"`
}

func (r *RoyaltyFee) InitFromResponse(fee token.RoyaltyFee) {
	r.Amount = fee.Amount
	r.FallbackFee.InitFromResponse(fee.FallbackFee)
	r.CollectorAccountID = fee.CollectorAccountID
}
