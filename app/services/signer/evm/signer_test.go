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

package evm

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func mockSigner() (service.Signer, *ecdsa.PrivateKey) {
	pk, _ := crypto.GenerateKey()
	pkBytes := crypto.FromECDSA(pk)
	s := NewEVMSigner(hexutil.Encode(pkBytes)[2:])
	return s, pk
}

func Test_Sign(t *testing.T) {
	s, _ := mockSigner()

	msg := []byte("12345678123456781234567812345678")
	res, err := s.Sign(msg)
	assert.Empty(t, err)
	assert.NotEmpty(t, res)
}

func TestSigner_NewKeyTransactor(t *testing.T) {
	s, _ := mockSigner()

	res, err := s.NewKeyTransactor(big.NewInt(80001))
	assert.Empty(t, err)
	assert.NotEmpty(t, res)
}

func Test_Address(t *testing.T) {
	s, pk := mockSigner()

	addr := crypto.PubkeyToAddress(pk.PublicKey).String()
	assert.Equal(t, s.Address(), addr)
}
