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

package decimal

import (
	"github.com/shopspring/decimal"
	"math"
	"math/big"
)

// ToLowestDenomination decimal amount to the lowest denomination
func ToLowestDenomination(amount decimal.Decimal, decimals uint8) *big.Int {

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	toSmallestDenomination := new(big.Int)
	toSmallestDenomination.SetString(result.String(), 10)

	return toSmallestDenomination
}

func ParseAmount(amount string) (result *decimal.Decimal, err error) {
	if amount == "" {
		zeroAmount := decimal.NewFromFloat(0.0)
		return &zeroAmount, nil
	}
	newResult, err := decimal.NewFromString(amount)

	return &newResult, err
}

// ToTargetAmount converts the provided amount to a target amount,
// depending on the decimals provided
// Examples: sourceDecimals 8, targetDecimals 8, amount 1 000 => result 1 000
// Examples: sourceDecimals 9, targetDecimals 8, amount 1 000 => result 100
// Examples: sourceDecimals 8, targetDecimals 9, amount 1 000 => result 10 000
func ToTargetAmount(sourceDecimals uint8, targetDecimals uint8, amount *big.Int) *big.Int {
	if sourceDecimals == targetDecimals {
		return amount
	} else if sourceDecimals > targetDecimals {
		power := int(sourceDecimals - targetDecimals)
		divider := big.NewInt(int64(math.Pow10(power)))

		return new(big.Int).Div(amount, divider)
	} else {
		power := int(targetDecimals - sourceDecimals)
		multiplier := big.NewInt(int64(math.Pow10(power)))

		return new(big.Int).Mul(amount, multiplier)
	}
}
