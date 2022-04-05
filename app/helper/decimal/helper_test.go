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
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"testing"
)

var (
	amountString      = "10"
	amount, _         = new(big.Int).SetString(amountString, 10)
	decimalAmount     = decimal.NewFromInt(amount.Int64())
	zeroDecimalAmount = decimal.NewFromFloat(0.0)
	decimals          = uint8(8)
	invalidAmount     = "asd"
	sourceDecimals    = uint8(9)
	targetDecimals    = uint8(18)
)

func Test_ToLowestDenomination(t *testing.T) {
	expectedAmount := new(big.Int).Mul(amount, big.NewInt(int64(math.Pow10(int(decimals)))))

	result := ToLowestDenomination(decimalAmount, decimals)

	assert.Equal(t, expectedAmount, result)
}

func Test_ParseAmount(t *testing.T) {
	result, err := ParseAmount(amountString)

	assert.Nil(t, err)
	assert.Equal(t, &decimalAmount, result)
}

func Test_ParseAmount_Err(t *testing.T) {
	result, err := ParseAmount(invalidAmount)

	assert.NotNil(t, err)
	assert.Equal(t, &decimal.Decimal{}, result)
}

func Test_ParseAmount_EmptyString(t *testing.T) {
	result, err := ParseAmount("")

	assert.Nil(t, err)
	assert.Equal(t, &zeroDecimalAmount, result)
}

func Test_ToTargetAmount_SourceMoreThanTargetDecimals(t *testing.T) {
	amount := big.NewInt(1_000_000_000)
	divider := big.NewInt(int64(math.Pow10(int(targetDecimals - sourceDecimals))))
	expectedAmount := new(big.Int).Div(amount, divider)

	result := ToTargetAmount(targetDecimals, sourceDecimals, amount)

	assert.Equal(t, expectedAmount, result)
}

func Test_ToTargetAmount_SourceMoreThanTargetDecimals_EqualsZero(t *testing.T) {
	result := ToTargetAmount(targetDecimals, sourceDecimals, amount)

	assert.Equal(t, big.NewInt(0), result)
}

func Test_ToTargetAmount_EqualDecimals(t *testing.T) {
	result := ToTargetAmount(targetDecimals, targetDecimals, amount)

	assert.Equal(t, amount, result)
}

func Test_ToTargetAmount_SourceLessThanTargetDecimals(t *testing.T) {
	multiplier := big.NewInt(int64(math.Pow10(int(targetDecimals - sourceDecimals))))
	expected := new(big.Int).Mul(amount, multiplier)

	result := ToTargetAmount(sourceDecimals, targetDecimals, amount)

	assert.Equal(t, expected, result)
}
