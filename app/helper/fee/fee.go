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
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"strconv"
)

// GetTotalFeeFromTransfers sums the positive amounts of transfers, excluding the receiver transfer
// Returns the sum and whether the receiver transfer has been found
func GetTotalFeeFromTransfers(transfers []model.Hedera, receiver hedera.AccountID) (totalFee string, hasReceiver bool) {
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
