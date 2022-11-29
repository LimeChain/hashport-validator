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
	testingFlatFeePolicyPolicy = FlatFeePolicy{
		Networks: []uint64{10, 20, 30, 40, 50},
		Value:    10,
	}
)

func Test_ParseNewFlatFeePolicy_Works(t *testing.T) {
	policy, err := ParseNewFlatFeePolicy(nil, int64(10))

	assert.Nil(t, err)
	assert.NotNil(t, policy)
}

func Test_FlatFeePolicy_FeeAmountFor_ShouldReturnFlatFee(t *testing.T) {
	feeAmount, exist := testingFlatFeePolicyPolicy.FeeAmountFor(10, "", 1000)

	assert.Equal(t, true, exist)
	assert.Equal(t, int64(10), feeAmount)
}
