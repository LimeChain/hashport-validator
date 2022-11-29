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
	networks = []uint64{10, 20, 30, 40, 50}
)

func Test_getInterfaceValue_Works(t *testing.T) {

	input := make(map[interface{}]interface{})
	input["token"] = 10

	val, valFound := getInterfaceValue(input, "token")

	assert.Equal(t, true, valFound)
	assert.Equal(t, 10, val)
}

func Test_networkAllowed_ShouldNotBeAllowedWithArrayOfNetworks(t *testing.T) {
	ok := networkAllowed(networks, 10000)

	assert.Equal(t, false, ok)
}

func Test_networkAllowed_ShouldBeAllowedWithArrayOfNetworks(t *testing.T) {
	ok := networkAllowed(networks, 10)

	assert.Equal(t, true, ok)
}

func Test_networkAllowed_ShouldBeAllowedWithEmtyNetworks(t *testing.T) {
	ok := networkAllowed(nil, 10)

	assert.Equal(t, true, ok)
}
