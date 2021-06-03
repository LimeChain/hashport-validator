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

package memo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	validEthAddressBase64    = "MHgzOEU4OTM3YjVBN2I5ZjM3OWIxNzBiOTlGNWJEZUIyYjM2NGRmQjY1"
	validEthAddress          = "0x38E8937b5A7b9f379b170b99F5bDeB2b364dfB65"
	nonValidEthAddress       = "0"
	nonValidEthAddressBase64 = "MHgzOEU4OTM3YjVBN2I5ZjM3OWIxNzBiOTlGNWJEZUIyYjM2NGRmQjYq"
)

func Test_Validate(t *testing.T) {
	decodedAddress, err := Validate(validEthAddressBase64)
	assert.Equal(t, decodedAddress, validEthAddress)
	assert.Nil(t, err)
}

func Test_WrongAddress(t *testing.T) {
	_, err := Validate(nonValidEthAddressBase64)
	expectedError := "Memo is invalid or has invalid encoding format: [0x38E8937b5A7b9f379b170b99F5bDeB2b364dfB6*]"
	assert.EqualError(t, err, expectedError)
}

func Test_InvalidAddress(t *testing.T) {
	_, err := Validate(nonValidEthAddress)
	expectedError := "Invalid base64 string provided: [illegal base64 data at input byte 0]"
	assert.EqualError(t, err, expectedError)
}
