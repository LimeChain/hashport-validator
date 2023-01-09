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

package fee

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	transfers []model.Hedera
)

func setup() {
	acc1, acc2, _ := accountIds()
	transfers = []model.Hedera{
		{
			AccountID: acc1,
			Amount:    -5000,
		},
		{
			AccountID: acc2,
			Amount:    5000,
		},
	}
}

func accountIds() (hedera.AccountID, hedera.AccountID, hedera.AccountID) {
	accId1, _ := hedera.AccountIDFromString("0.0.1")
	accId2, _ := hedera.AccountIDFromString("0.0.2")
	accId3, _ := hedera.AccountIDFromString("0.0.3")
	return accId1, accId2, accId3
}

func Test_TotalFeeFromTransfers(t *testing.T) {
	setup()
	_, receiver, _ := accountIds()

	expectedReceiverFound := true
	expectedFee := "0"

	actualFee, actualReceiverFound := TotalFeeFromTransfers(transfers, receiver)
	assert.Equal(t, expectedFee, actualFee)
	assert.Equal(t, expectedReceiverFound, actualReceiverFound)
}

func Test_TotalFeeFromTransfers_ReceiverNotFound(t *testing.T) {
	setup()
	_, _, receiver2 := accountIds()

	expectedReceiverFound := false
	expectedFee := "5000"

	actualFee, actualReceiverFound := TotalFeeFromTransfers(transfers, receiver2)
	assert.Equal(t, expectedFee, actualFee)
	assert.Equal(t, expectedReceiverFound, actualReceiverFound)
}

func Test_SumFallbackFeeAmounts(t *testing.T) {
	hbarFee := asset.FixedFee{
		Amount:              50,
		DenominatingTokenId: nil, // HBAR
	}
	token1Name := "token1"
	token1Fee := asset.FixedFee{
		Amount:              150,
		DenominatingTokenId: &token1Name,
	}
	token2Name := "token2"
	token2Fee1 := asset.FixedFee{
		Amount:              30,
		DenominatingTokenId: &token2Name,
	}
	token2Fee2 := asset.FixedFee{
		Amount:              60,
		DenominatingTokenId: &token2Name,
	}
	customFees := asset.CustomFees{
		RoyaltyFees: []asset.RoyaltyFee{
			{
				FallbackFee: hbarFee,
			},
			{
				FallbackFee: token1Fee,
			},
			{
				FallbackFee: token2Fee1,
			},
			{
				FallbackFee: token2Fee2,
			},
		},
	}

	result := SumFallbackFeeAmounts(customFees)

	assert.Equal(t, result.FallbackFeeAmountInHbar, hbarFee.Amount)
	assert.Equal(t, result.FallbackFeeAmountsByTokenId[token1Name], token1Fee.Amount)

	token2TotalFee := token2Fee1.Amount + token2Fee2.Amount
	assert.Equal(t, result.FallbackFeeAmountsByTokenId[token2Name], token2TotalFee)
}
