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
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func DecodeSignature(signature string) (decodedSignature []byte, ethSignature string, err error) {
	decodedSig, err := hex.DecodeString(signature)
	if err != nil {
		return nil, "", err
	}

	return switchSignatureValueV(decodedSig)
}

func RecoverSignerFromBytes(hash []byte, signature []byte) (string, error) {
	key, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		return "", err
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(*pubKey).String(), nil
}

func RecoverSignerFromStr(signature string, authMsgBytes []byte) (signerAddress string, signatureHex string, err error) {
	signatureBytes, signatureHex, err := DecodeSignature(signature)
	if err != nil {
		return "", "", err
	}

	signerAddress, err = RecoverSignerFromBytes(authMsgBytes, signatureBytes)
	if err != nil {
		return "", "", err
	}
	return signerAddress, signatureHex, nil
}

func switchSignatureValueV(decodedSig []byte) (decodedSignature []byte, ethSignature string, err error) {
	if len(decodedSig) != 65 {
		return nil, "", errors.New("invalid signature length")
	}

	// note: https://github.com/ethereum/go-ethereum/issues/19751
	evmSig := make([]byte, len(decodedSig))
	copy(evmSig, decodedSig)

	if decodedSig[64] == 0 || decodedSig[64] == 1 {
		evmSig[64] += 27
	} else if decodedSig[64] == 27 || decodedSig[64] == 28 {
		decodedSig[64] -= 27
	}

	return decodedSig, hex.EncodeToString(evmSig), nil
}

func OriginatorFromTx(tx *types.Transaction) (string, error) {
	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()), nil)
	if err != nil {
		return "", err
	}

	return msg.From().String(), nil
}
