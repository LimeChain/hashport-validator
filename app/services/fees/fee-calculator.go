/*
 * Copyright 2021 LimeChain Ltd.
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

package fees

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"math/big"
)

func getFee() (*big.Int, error) {
	return new(big.Int), nil
}

func ValidateExecutionFee(strTransferFee string) (bool, error) {
	transferFee, err := helper.ToBigInt(strTransferFee)
	if err != nil {
		return false, err
	}

	estimatedFee, err := getFee()
	if err != nil {
		return false, err
	}

	if transferFee.Cmp(estimatedFee) >= 0 {
		return true, nil
	}

	return false, nil
}
