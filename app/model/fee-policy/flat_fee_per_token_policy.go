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

type FlatFeePerTokenPolicy struct {
	Networks  []uint64
	TokenFees map[string]int64 // { token: "0.0.4564", value: 7 }
}

func ParseNewFlatFeePerTokenPolicy(networks []uint64, parsingValue interface{}) *FlatFeePerTokenPolicy {
	var result = &FlatFeePerTokenPolicy{
		Networks: networks,
	}
	result.TokenFees = make(map[string]int64)

	pairs := parsingValue.([]interface{})

	for _, item := range pairs {
		token, foundToken := getInterfaceValue(item, "token")
		value, foundValue := getInterfaceValue(item, "value")

		if foundToken && foundValue {
			result.TokenFees[token.(string)] = (int64(value.(int)))
		}
	}

	return result
}

func (policy *FlatFeePerTokenPolicy) FeeAmountFor(networkId uint64, token string, amount int64) (feeAmount int64, exist bool) {
	var found bool = networkFound(policy.Networks, networkId)

	if found {
		value, ok := policy.TokenFees[token]

		return value, ok
	}

	return 0, false
}
