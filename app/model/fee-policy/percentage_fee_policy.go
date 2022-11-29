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

package fee_policy

import (
	"errors"

	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

type PercentageFeePolicy struct {
	Networks []uint64
	Value    int64
}

func ParseNewPercentageFeePolicy(networks []uint64, parsingValue interface{}) (*PercentageFeePolicy, error) {
	value, ok := parsingValue.(int64)

	if !ok {
		return nil, errors.New("value is not integer")
	}

	return &PercentageFeePolicy{
		Networks: networks,
		Value:    int64(value),
	}, nil
}

func (policy *PercentageFeePolicy) FeeAmountFor(networkId uint64, _ string, amount int64) (feeAmount int64, exist bool) {
	found := networkAllowed(policy.Networks, networkId)

	if found {
		feeAmount = amount * policy.Value / constants.FeeMaxPercentage
		return feeAmount, true
	}

	return 0, false
}
