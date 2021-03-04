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

package servicefee_test

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/bridge/servicefee"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSet(t *testing.T) {
	serviceFeeInstance := servicefee.Servicefee{}
	newServiceFee := big.NewInt(int64(5))
	serviceFeeInstance.Set(*newServiceFee)

	serviceFee := serviceFeeInstance.Get()
	assert.Equal(t, serviceFee, newServiceFee, "Service fee was not set correctly")
}
