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
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testingFlatFeePerTokenPolicyPolicy = FlatFeePerTokenPolicy{
		Networks: []uint64{10, 20, 30, 40, 50},
		TokenFees: map[string]int64{
			"token1": 100,
			"token2": 200,
			"token3": 300,
			"token4": 400,
		},
	}
)

func Test_ParseNewFlatFeePerTokenPolicy_Works(t *testing.T) {
	token1 := map[interface{}]interface{}{"token": "token1", "value": int64(100)}
	token2 := map[interface{}]interface{}{"token": "token2", "value": int64(200)}

	tokens := []interface{}{token1, token2}

	policy, err := ParseNewFlatFeePerTokenPolicy(nil, tokens)

	assert.Nil(t, err)
	assert.NotNil(t, policy)
}

func Test_FlatFeePerTokenPolicy_FeeAmountFor_ShouldReturnFlatFee(t *testing.T) {
	feeAmount, exist := testingFlatFeePerTokenPolicyPolicy.FeeAmountFor(10, "token1", 1000)

	assert.Equal(t, true, exist)
	assert.Equal(t, int64(100), feeAmount)
}

func Test_FlatFeePerTokenPolicy_FeeAmountFor_ShouldReturnNotFound(t *testing.T) {
	_, exist := testingFlatFeePerTokenPolicyPolicy.FeeAmountFor(10, "token100", 1000)

	assert.Equal(t, false, exist)
}
