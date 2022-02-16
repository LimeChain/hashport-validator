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

package auth_message

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
	"math/big"
)

// EncodeFungibleBytesFrom returns the array of bytes representing an
// authorisation ERC-20 Mint signature ready to be signed by EVM Private Key
func EncodeFungibleBytesFrom(sourceChainId, targetChainId uint64, txId, asset, receiverEthAddress, amount string) ([]byte, error) {
	args, err := generateFungibleArguments()
	if err != nil {
		return nil, err
	}
	amountBn, err := big_numbers.ToBigInt(amount)
	if err != nil {
		return nil, err
	}

	bytesToHash, err := args.Pack(
		new(big.Int).SetUint64(sourceChainId),
		new(big.Int).SetUint64(targetChainId),
		[]byte(txId),
		common.HexToAddress(asset),
		common.HexToAddress(receiverEthAddress),
		amountBn)
	if err != nil {
		return nil, err
	}
	return keccak(bytesToHash), nil
}

// EncodeNftBytesFrom returns the array of bytes representing an
// authorisation ERC-721 NFT signature for Mint ready to be signed by EVM Private Key
func EncodeNftBytesFrom(sourceChainId, targetChainId uint64, txId, asset string, serialNum int64, metadata, receiverEthAddress string) ([]byte, error) {
	args, err := generateNftArguments()
	if err != nil {
		return nil, err
	}

	bytesToHash, err := args.Pack(
		new(big.Int).SetUint64(sourceChainId),
		new(big.Int).SetUint64(targetChainId),
		[]byte(txId),
		common.HexToAddress(asset),
		big.NewInt(serialNum),
		metadata,
		common.HexToAddress(receiverEthAddress))
	if err != nil {
		return nil, err
	}
	return keccak(bytesToHash), nil
}

func generateNftArguments() (abi.Arguments, error) {
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, err
	}

	stringType, err := abi.NewType("string", "internalType", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		{
			Type: uint256Type,
		},
		{
			Type: uint256Type,
		},
		{
			Type: bytesType,
		},
		{
			Type: addressType,
		},
		{
			Type: uint256Type,
		},
		{
			Type: stringType,
		},
		{
			Type: addressType,
		},
	}, nil
}

func generateFungibleArguments() (abi.Arguments, error) {
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		{
			Type: uint256Type,
		},
		{
			Type: uint256Type,
		},
		{
			Type: bytesType,
		},
		{
			Type: addressType,
		},
		{
			Type: addressType,
		},
		{
			Type: uint256Type,
		}}, nil
}

func keccak(encodedData []byte) []byte {
	toEthSignedMsg := []byte("\x19Ethereum Signed Message:\n32")
	hash := crypto.Keccak256(encodedData)
	return crypto.Keccak256(toEthSignedMsg, hash)
}
