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
	"github.com/shopspring/decimal"
	"math/big"
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

type NonFungibleAssetInfo struct {
	Name          string   `json:"name"`
	Symbol        string   `json:"symbol"`
	IsNative      bool     `json:"isNative"`
	ReserveAmount *big.Int `json:"-"`
}
