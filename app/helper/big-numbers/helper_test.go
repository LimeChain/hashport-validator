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

package big_numbers

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

const (
	validNumber            = "54321"
	notValidNumber   = "0xsomerouteraddress"

)

func Test_StringToBigInt(t *testing.T) {

	value, err := ToBigInt(validNumber)
	assert.IsType(t, big.Int{}, *value)
	assert.Nil(t, err)
}

func Test_ToBigIntError(t *testing.T) {
	_, err := ToBigInt(notValidNumber)
	assert.NotNil(t, err)
}
