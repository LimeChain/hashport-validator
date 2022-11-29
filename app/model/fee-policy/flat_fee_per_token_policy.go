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
)

type FlatFeePerTokenPolicy struct {
	Networks  []uint64
	TokenFees map[string]int64
}

func ParseNewFlatFeePerTokenPolicy(networks []uint64, parsingValue interface{}) (*FlatFeePerTokenPolicy, error) {
	result := &FlatFeePerTokenPolicy{
		Networks: networks,
	}
	result.TokenFees = make(map[string]int64)

	pairs, ok := parsingValue.([]interface{})

	if !ok {
		return nil, errors.New("invalid map of token fee pairs")
	}

	for _, item := range pairs {
		token, foundToken := getInterfaceValue(item, "token")
		if !foundToken {
			return nil, errors.New("map does not contains key 'token'")
		}

		value, foundValue := getInterfaceValue(item, "value")
		if !foundValue {
			return nil, errors.New("map does not contains key 'value'")
		}

		if foundToken && foundValue {
			feeValue, ok := value.(int64)

			if !ok {
				return nil, errors.New("value is not integer")
			}
			result.TokenFees[token.(string)] = int64(feeValue)
		}
	}

	return result, nil
}

func (policy *FlatFeePerTokenPolicy) FeeAmountFor(networkId uint64, token string, _ int64) (feeAmount int64, exist bool) {
	found := networkAllowed(policy.Networks, networkId)

	if found {
		value, ok := policy.TokenFees[token]

		return value, ok
	}

	return 0, false
}
