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

package ethereum

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	signature = "e1e013272549101580c5fe79f8958c0ca4c559df93a162f7c3b5049308150aef1e38ffae06c603d5aecdd0d07202d9a14c4b33beb9c0e66606054844da62ce5d1c"
	invalidSignature = "0xb722ebf2400c2b67f73dbed19cf113e61b6a09320936ae9665c3ab125be8d6a945b040cdd8e2b82cc63c8a0bdf3ecaa2c675e48bf3fefb2c8de61defb6b1960d1b"
	signatureBytes = []byte{225, 224, 19, 39, 37, 73, 16, 21, 128, 197, 254, 121, 248, 149, 140, 12, 164, 197, 89, 223, 147, 161, 98, 247, 195, 181, 4, 147, 8, 21, 10, 239, 30, 56, 255, 174, 6, 198, 3, 213, 174, 205, 208, 208, 114, 2, 217, 161, 76, 75, 51, 190, 185, 192, 230, 102, 6, 5, 72, 68, 218, 98, 206, 93, 1}
	hashedData = []byte{251, 0, 78, 78, 113, 46, 99, 87, 61, 58, 142, 57, 214, 119, 55, 51, 200, 234, 182, 103, 168, 35, 143, 137, 207, 39, 48, 247, 171, 4, 159, 26}
	invalidHashedData = []byte{0, 78, 78, 113, 46, 99, 87, 61, 58, 142, 57, 214, 119, 55, 51, 200, 234, 182, 103, 168, 35, 143, 137, 207, 39, 48, 247, 171, 4, 159, 26}
)

func Test_DecodeSignature(t *testing.T) {
	decodedSignature, ethSignature, err := DecodeSignature(signature)
	assert.Equal(t, signatureBytes, decodedSignature)
	assert.Equal(t, signature, ethSignature)
	assert.Nil(t, err)
}

func Test_DecodeSignatureError(t *testing.T) {
	_, _, err := DecodeSignature(invalidSignature)
	assert.NotNil(t, err)
}

func Test_RecoverSignerFromBytes(t *testing.T) {
	signerAddress, err := RecoverSignerFromBytes(hashedData, signatureBytes)
	fmt.Println(signerAddress)
	assert.Nil(t, err)
}

func Test_RecoverSignerFromBytesError(t *testing.T) {
	_, err := RecoverSignerFromBytes(invalidHashedData, signatureBytes)
	assert.NotNil(t, err)
}

func Test_RecoverSignerFromString(t *testing.T) {
	signerAddress, signatureHex, err := RecoverSignerFromStr(signature, hashedData)
	fmt.Println(signerAddress)
	assert.Equal(t, signature, signatureHex)
	assert.Nil(t, err)
}

func Test_RecoverSignerFromStringWrongMsg(t *testing.T) {
	_, _, err := RecoverSignerFromStr(signature, invalidHashedData)
	assert.NotNil(t, err)
}

func Test_RecoverSignerFromStringWrongSig(t *testing.T) {
	_, _, err := RecoverSignerFromStr(invalidSignature, hashedData)
	assert.NotNil(t, err)
}
