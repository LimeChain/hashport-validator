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
	"strconv"
)

// TotalFeeFromTransfers sums the positive amounts of transfers, excluding the receiver transfer
// Returns the sum and whether the receiver transfer has been found
func TotalFeeFromTransfers(transfers []model.Hedera, receiver hedera.AccountID) (totalFee string, hasReceiver bool) {
	result := int64(0)
	for _, transfer := range transfers {
		if transfer.Amount < 0 {
			continue
		}
		if transfer.AccountID == receiver {
			hasReceiver = true
			continue
		}
		result += transfer.Amount
	}

	return strconv.FormatInt(result, 10), hasReceiver
}

// SumFallbackFeeAmounts sums fallback and fixed fees in HBAR and by token ID
// Returns the sum of the fallback and fixed fees in HBAR and by token ID
func SumFallbackFeeAmounts(customFees asset.CustomFees) asset.CustomFeeTotalAmounts {
	customFeeAmounts := new(asset.CustomFeeTotalAmounts)
	sumFallbackFeeAmounts(customFees.RoyaltyFees, customFeeAmounts)
	sumFixedFeeAmounts(customFees.FixedFees, customFeeAmounts)

	return *customFeeAmounts
}

func sumFallbackFeeAmounts(royaltyFees []asset.RoyaltyFee, customFeeAmounts *asset.CustomFeeTotalAmounts) {
	for _, royaltyFee := range royaltyFees {
		if royaltyFee.FallbackFee.DenominatingTokenId == nil {
			customFeeAmounts.FallbackFeeAmountInHbar += royaltyFee.FallbackFee.Amount
			customFeeAmounts.TotalFeeAmountsInHbar += royaltyFee.FallbackFee.Amount
		} else {
			tokenId := *royaltyFee.FallbackFee.DenominatingTokenId
			if _, ok := customFeeAmounts.FallbackFeeAmountsByTokenId[tokenId]; !ok {
				customFeeAmounts.FallbackFeeAmountsByTokenId[tokenId] = royaltyFee.FallbackFee.Amount
				customFeeAmounts.TotalAmountsByTokenId[tokenId] = royaltyFee.FallbackFee.Amount
			} else {
				customFeeAmounts.FallbackFeeAmountsByTokenId[tokenId] += royaltyFee.FallbackFee.Amount
				customFeeAmounts.TotalAmountsByTokenId[tokenId] += royaltyFee.FallbackFee.Amount
			}
		}
	}
}

func sumFixedFeeAmounts(fixedFees []asset.FixedFee, customFeeAmounts *asset.CustomFeeTotalAmounts) {
	for _, fixedFee := range fixedFees {
		if fixedFee.DenominatingTokenId == nil {
			customFeeAmounts.FixedFeeAmountInHbar += fixedFee.Amount
			customFeeAmounts.TotalFeeAmountsInHbar += fixedFee.Amount
		} else {
			tokenId := *fixedFee.DenominatingTokenId
			if _, ok := customFeeAmounts.FixedFeeAmountsByTokenId[tokenId]; !ok {
				customFeeAmounts.FixedFeeAmountsByTokenId[tokenId] = fixedFee.Amount
				customFeeAmounts.TotalAmountsByTokenId[tokenId] = fixedFee.Amount
			} else {
				customFeeAmounts.FixedFeeAmountsByTokenId[tokenId] += fixedFee.Amount
				customFeeAmounts.TotalAmountsByTokenId[tokenId] += fixedFee.Amount
			}
		}
	}
}
