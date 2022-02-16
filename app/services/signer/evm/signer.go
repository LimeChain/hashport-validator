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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"math/big"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
}

func NewEVMSigner(privateKey string) *Signer {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("Invalid EVM Private Key provided: [%s]", privateKey)
	}
	return &Signer{privateKey: pk}
}

func (s *Signer) Sign(msg []byte) ([]byte, error) {
	signature, err := crypto.Sign(msg, s.privateKey)
	if err != nil {
		return nil, err
	}
	// note: https://github.com/ethereum/go-ethereum/issues/19751
	signature[64] += 27

	return signature, nil
}

func (s *Signer) NewKeyTransactor(chainId *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(s.privateKey, chainId)
}

func (s *Signer) Address() string {
	publicKey := s.privateKey.Public().(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKey).String()
}
