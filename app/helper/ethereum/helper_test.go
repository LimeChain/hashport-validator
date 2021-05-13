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
	expectedAddress       = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	signature             = "69b6640d0a0b32538def8cfeb0bc9aa21a7b83fac2486b99d7e438c1020494394aa6c9a59c9f55450c4a6496cc8cd216fedcb38b6e82248ff9ba718b8ae4cd3b1b"
	invalidSignature      = "0xb722ebf2400c2b67f73dbed19cf113e61b6a09320936ae9665c3ab125be8d6a945b040cdd8e2b82cc63c8a0bdf3ecaa2c675e48bf3fefb2c8de61defb6b1960d1b"
	signatureBytes        = []byte{105, 182, 100, 13, 10, 11, 50, 83, 141, 239, 140, 254, 176, 188, 154, 162, 26, 123, 131, 250, 194, 72, 107, 153, 215, 228, 56, 193, 2, 4, 148, 57, 74, 166, 201, 165, 156, 159, 85, 69, 12, 74, 100, 150, 204, 140, 210, 22, 254, 220, 179, 139, 110, 130, 36, 143, 249, 186, 113, 139, 138, 228, 205, 59, 0}
	invalidSignatureBytes = []byte{105, 182, 100, 13, 10, 11, 50, 83, 141, 239, 140, 254, 176, 188, 154, 162, 26, 123, 131, 250, 194, 72, 107, 153, 215, 228, 56, 193, 2, 4, 148, 57, 74, 166, 201, 165, 156, 159, 85, 69, 12, 74, 100, 150, 204, 140, 210, 22, 254, 220, 179, 139, 110, 130, 36, 143, 249, 186, 113, 139, 138, 228, 205, 59, 59, 0}
	hashedData            = []byte{251, 0, 78, 78, 113, 46, 99, 87, 61, 58, 142, 57, 214, 119, 55, 51, 200, 234, 182, 103, 168, 35, 143, 137, 207, 39, 48, 247, 171, 4, 159, 26}
	invalidHashedData     = []byte{0, 78, 78, 113, 46, 99, 87, 61, 58, 142, 57, 214, 119, 55, 51, 200, 234, 182, 103, 168, 35, 143, 137, 207, 39, 48, 247, 171, 4, 159, 26}
)

func Test_DecodeSignature(t *testing.T) {
	decodedSignature, ethSignature, err := DecodeSignature(signature)
	assert.Equal(t, signatureBytes, decodedSignature)
	assert.Equal(t, signature, ethSignature)
	assert.Nil(t, err)
}

func Test_DecodeSignatureError(t *testing.T) {
	_, _, err := DecodeSignature(invalidSignature)
	assert.Error(t, err)
}

func Test_RecoverSignerFromBytes(t *testing.T) {
	signerAddress, err := RecoverSignerFromBytes(hashedData, signatureBytes)
	fmt.Println(signerAddress)
	assert.Equal(t, expectedAddress, signerAddress)
	assert.Nil(t, err)
}

func Test_RecoverSignerFromBytesError(t *testing.T) {
	_, err := RecoverSignerFromBytes(invalidHashedData, signatureBytes)
	assert.Error(t, err)
}

func Test_RecoverSignerFromString(t *testing.T) {
	signerAddress, signatureHex, err := RecoverSignerFromStr(signature, hashedData)
	assert.Equal(t, expectedAddress, signerAddress)
	assert.Equal(t, signature, signatureHex)
	assert.Nil(t, err)
}

func Test_RecoverSignerFromStringWrongMsg(t *testing.T) {
	_, _, err := RecoverSignerFromStr(signature, invalidHashedData)
	assert.Error(t, err)
}

func Test_RecoverSignerFromStringWrongSig(t *testing.T) {
	_, _, err := RecoverSignerFromStr(invalidSignature, hashedData)
	assert.Error(t, err)
}

func Test_switchSignatureValueVLengthErr(t *testing.T) {
	_, _, err := switchSignatureValueV(invalidSignatureBytes)
	assert.Error(t, err)
}

func Test_switchSignatureValueV(t *testing.T) {
	_, _, err := switchSignatureValueV(signatureBytes)
	assert.Nil(t, err)
}
