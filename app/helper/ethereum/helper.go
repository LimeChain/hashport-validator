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
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/params"

	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
)

const (
	MintFunctionParameterAmount        = "amount"
	MintFunctionParameterReceiver      = "receiver"
	MintFunctionParameterSignatures    = "signatures"
	MintFunctionParameterWrappedToken  = "wrappedToken"
	MintFunctionParameterTransactionId = "transactionId"
	MintFunctionParameterTxCost        = "txCost"
)

const (
	MintFunction                = "mintWithReimbursement"
	MintFunctionParametersCount = 6
)

var (
	ErrorInvalidMintFunctionParameters = errors.New("invalid mint function parameters length")
)

func DecodeSignature(signature string) (decodedSignature []byte, ethSignature string, err error) {
	decodedSig, err := hex.DecodeString(signature)
	if err != nil {
		return nil, "", err
	}

	return switchSignatureValueV(decodedSig)
}

func DecodeBridgeMintFunction(data []byte) (txId, ethAddress, erc20address, amount, fee string, signatures [][]byte, err error) {
	bridgeAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		return "", "", "", "", "", nil, err
	}

	// bytes transactionId, address receiver, address erc20address, uint256 amount, uint256 fee, uint256 gasCost, bytes[] signatures
	decodedParameters := make(map[string]interface{})
	err = bridgeAbi.Methods[MintFunction].Inputs.UnpackIntoMap(decodedParameters, data[4:]) // data[4:] <- slice function name
	if err != nil {
		return "", "", "", "", "", nil, err
	}

	if len(decodedParameters) != MintFunctionParametersCount {
		return "", "", "", "", "", nil, ErrorInvalidMintFunctionParameters
	}

	transactionId := decodedParameters[MintFunctionParameterTransactionId].([]byte)
	receiver := decodedParameters[MintFunctionParameterReceiver].(common.Address)
	amountBn := decodedParameters[MintFunctionParameterAmount].(*big.Int)
	txCost := decodedParameters[MintFunctionParameterTxCost].(*big.Int)
	signatures = decodedParameters[MintFunctionParameterSignatures].([][]byte)

	erc20 := decodedParameters[MintFunctionParameterWrappedToken].(common.Address)

	for _, sig := range signatures {
		_, _, err := switchSignatureValueV(sig)
		if err != nil {
			return "", "", "", "", "", nil, err
		}
	}

	return string(transactionId), receiver.String(), erc20.String(), amountBn.String(), txCost.String(), signatures, nil
}

func GweiToWei(gwei *big.Int) *big.Int {
	return new(big.Int).Mul(gwei, big.NewInt(params.GWei))
}

func WeiToGwei(wei *big.Int) *big.Int {
	return new(big.Int).Div(wei, big.NewInt(params.GWei))
}

func GetAddressBySignature(hash []byte, signature []byte) (string, error) {
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

func switchSignatureValueV(decodedSig []byte) (decodedSignature []byte, ethSignature string, err error) {
	if len(decodedSig) != 65 {
		return nil, "", errors.New("invalid signature length")
	}

	// note: https://github.com/ethereum/go-ethereum/issues/19751
	ethSig := make([]byte, len(decodedSig))
	copy(ethSig, decodedSig)

	if decodedSig[64] == 0 || decodedSig[64] == 1 {
		ethSig[64] += 27
	} else if decodedSig[64] == 27 || decodedSig[64] == 28 {
		decodedSig[64] -= 27
	}

	return decodedSig, hex.EncodeToString(ethSig), nil
}
