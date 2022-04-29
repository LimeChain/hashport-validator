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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	assetModel "github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"math/big"
)

type Assets interface {
	// FungibleNetworkAssets Gets all Fungible Assets by Network ID
	FungibleNetworkAssets() map[uint64][]string
	// NonFungibleNetworkAssets Gets all Non-Fungible Assets by Network ID
	NonFungibleNetworkAssets() map[uint64][]string
	// NativeToWrappedAssets Gets all Native assets with their Wrapped assets by Network ID
	NativeToWrappedAssets() map[uint64]map[string]map[uint64]string
	// WrappedFromNative Gets All Wrapped Assets for passed Native Asset's address
	WrappedFromNative(nativeChainId uint64, nativeAssetAddress string) map[uint64]string
	// NativeToWrapped Gets Wrapped Asset for passed Native Asset's address and Target Chain ID
	NativeToWrapped(nativeAssetAddress string, nativeChainId, targetChainId uint64) string
	// WrappedToNative Gets Native Asset from passed Wrapped Asset's address and Wrapped Chain ID
	WrappedToNative(wrappedAssetAddress string, wrappedChainId uint64) *assetModel.NativeAsset
	// FungibleNetworkAssetsByChainId Gets all Fungible assets for passed Chain ID
	FungibleNetworkAssetsByChainId(chainId uint64) []string
	// FungibleNativeAsset Gets NativeAsset for passed chainId and assetAddress
	FungibleNativeAsset(chainId uint64, nativeAssetAddress string) *assetModel.NativeAsset
	// IsNative Returns flag showing if the passed Asset's Address is Native for the passed Chain ID
	IsNative(chainId uint64, assetAddress string) bool
	// OppositeAsset Gets Opposite asset for passed chain IDs and assetAddress
	OppositeAsset(sourceChainId uint64, targetChainId uint64, assetAddress string) string
	// FungibleAssetInfo Gets FungibleAssetInfo
	FungibleAssetInfo(networkId uint64, assetAddress string) (assetInfo *assetModel.FungibleAssetInfo, exist bool)
	// NonFungibleAssetInfo Gets NonFungibleAssetInfo
	NonFungibleAssetInfo(networkId uint64, assetAddressOrId string) (assetInfo *assetModel.NonFungibleAssetInfo, exist bool)
	// FetchHederaTokenReserveAmount Gets Hedera's Token Reserve Amount
	FetchHederaTokenReserveAmount(assetId string, mirrorNode client.MirrorNode, isNative bool, hederaTokenBalances map[string]int) (reserveAmount *big.Int, err error)
	// FetchEvmFungibleReserveAmount Gets EVM's Fungible Token Reserve Amount
	FetchEvmFungibleReserveAmount(networkId uint64, assetAddress string, isNative bool, evmTokenClient client.EvmFungibleToken, routerContractAddress string) (inLowestDenomination *big.Int, err error)
	// FetchEvmNonFungibleReserveAmount Gets EVM's Non-Fungible Token Reserve Amount
	FetchEvmNonFungibleReserveAmount(networkId uint64, assetAddress string, isNative bool, evmTokenClient client.EvmNft, routerContractAddress string) (inLowestDenomination *big.Int, err error)
}
