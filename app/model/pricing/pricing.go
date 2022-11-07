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

package pricing

import (
	"math/big"

	"github.com/shopspring/decimal"
)

type TokenPriceInfo struct {
	UsdPrice         decimal.Decimal
	MinAmountWithFee *big.Int
	DefaultMinAmount *big.Int
}

type NonFungibleFee struct {
	IsNative     bool            `json:"isNative"`
	PaymentToken string          `json:"paymentToken"`
	Fee          decimal.Decimal `json:"fee"`
}
