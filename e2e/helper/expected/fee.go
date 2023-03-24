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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
)

func ReceiverAndFeeAmounts(
	feeCalc service.Fee,
	distributor service.Distributor,
	token string,
	amount int64,
	targetChainId uint64,
	sender string,
) (receiverAmount, fee int64) {
	fee, remainder := feeCalc.CalculateFee(targetChainId, sender, token, amount)
	validFee := distributor.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	return remainder, validFee
}
