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

package expected

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
)

func ReceiverAndFeeAmounts(feeCalc service.Fee, distributor service.Distributor, token string, amount int64) (receiverAmount, valdiatorsFee, treasuryFee int64) {
	fee, remainder := feeCalc.CalculateFee(token, amount)
	validTreasuryFee, validValidatorsFee := distributor.ValidAmounts(fee)
	validFee := validTreasuryFee + validValidatorsFee
	if validFee != fee {
		remainder += fee - validFee
	}

	return remainder, validValidatorsFee, validTreasuryFee
}

func EvmAmoundAndFee(router *router.Router, token string, amount int64, t *testing.T) (*big.Int, *big.Int) {
	t.Helper()
	amountBn := big.NewInt(amount)

	feeData, err := router.TokenFeeData(nil, common.HexToAddress(token))
	if err != nil {
		t.Fatal(err)
	}

	precision, err := router.ServiceFeePrecision(nil)
	if err != nil {
		t.Fatal(err)
	}

	multiplied := new(big.Int).Mul(amountBn, feeData.ServiceFeePercentage)
	serviceFee := new(big.Int).Div(multiplied, precision)

	return new(big.Int).Sub(amountBn, serviceFee), serviceFee
}
