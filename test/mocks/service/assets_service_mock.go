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
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockAssetsService struct {
	mock.Mock
}

// FungibleNetworkAssets Gets all Fungible Assets by Network ID
func (mas *MockAssetsService) FungibleNetworkAssets() map[uint64][]string {
	args := mas.Called()
	result := args.Get(0).(map[uint64][]string)
	return result
}

// NonFungibleNetworkAssets Gets all Non-Fungible Assets by Network ID
func (mas *MockAssetsService) NonFungibleNetworkAssets() map[uint64][]string {
	args := mas.Called()
	result := args.Get(0).(map[uint64][]string)
	return result
}

// NativeToWrappedAssets Gets all Native assets with their Wrapped assets by Network ID
func (mas *MockAssetsService) NativeToWrappedAssets() map[uint64]map[string]map[uint64]string {
	args := mas.Called()
	result := args.Get(0).(map[uint64]map[string]map[uint64]string)
	return result
}

// WrappedFromNative Gets All Wrapped Assets for passed Native Asset's address
func (mas *MockAssetsService) WrappedFromNative(nativeChainId uint64, nativeAssetAddress string) map[uint64]string {
	args := mas.Called(nativeChainId, nativeAssetAddress)
	result := args.Get(0).(map[uint64]string)
	return result
}

// NativeToWrapped Gets Wrapped Asset for passed Native Asset's address and Target Chain ID
func (mas *MockAssetsService) NativeToWrapped(nativeAssetAddress string, nativeChainId, targetChainId uint64) string {
	args := mas.Called(nativeAssetAddress, nativeChainId, targetChainId)
	result := args.Get(0).(string)
	return result
}

// WrappedToNative Gets Native Asset from passed Wrapped Asset's address and Wrapped Chain ID
func (mas *MockAssetsService) WrappedToNative(wrappedAssetAddress string, wrappedChainId uint64) *assetModel.NativeAsset {
	args := mas.Called(wrappedAssetAddress, wrappedChainId)
	result := args.Get(0).(*assetModel.NativeAsset)
	return result
}

// FungibleNetworkAssetsByChainId Gets all Fungible assets for passed Chain ID
func (mas *MockAssetsService) FungibleNetworkAssetsByChainId(chainId uint64) []string {
	args := mas.Called(chainId)
	result := args.Get(0).([]string)
	return result
}

// FungibleNativeAsset Gets NativeAsset for passed chainId and assetAddress
func (mas *MockAssetsService) FungibleNativeAsset(chainId uint64, nativeAssetAddress string) *assetModel.NativeAsset {
	args := mas.Called(chainId, nativeAssetAddress)
	result := args.Get(0).(*assetModel.NativeAsset)
	return result
}

// IsNative Returns flag showing if the passed Asset's Address is Native for the passed Chain ID
func (mas *MockAssetsService) IsNative(chainId uint64, assetAddress string) bool {
	args := mas.Called(chainId, assetAddress)
	result := args.Get(0).(bool)
	return result
}

// OppositeAsset Gets Opposite asset for passed chain IDs and assetAddress
func (mas *MockAssetsService) OppositeAsset(sourceChainId uint64, targetChainId uint64, assetAddress string) string {
	args := mas.Called(sourceChainId, targetChainId, assetAddress)
	result := args.Get(0).(string)
	return result
}

// FungibleAssetInfo Gets FungibleAssetInfo
func (mas *MockAssetsService) FungibleAssetInfo(networkId uint64, assetAddress string) (assetInfo *assetModel.FungibleAssetInfo, exist bool) {
	args := mas.Called(networkId, assetAddress)
	assetInfo = args.Get(0).(*assetModel.FungibleAssetInfo)
	exist = args.Get(1).(bool)

	return assetInfo, exist
}

// NonFungibleAssetInfo Gets NonFungibleAssetInfo
func (mas *MockAssetsService) NonFungibleAssetInfo(networkId uint64, assetAddress string) (assetInfo *assetModel.NonFungibleAssetInfo, exist bool) {
	args := mas.Called(networkId, assetAddress)
	assetInfo = args.Get(0).(*assetModel.NonFungibleAssetInfo)
	exist = args.Get(1).(bool)

	return assetInfo, exist
}

// FetchHederaTokenReserveAmount Gets Hedera's Token Reserve Amount
func (mas *MockAssetsService) FetchHederaTokenReserveAmount(assetId string, mirrorNode client.MirrorNode, isNative bool, hederaTokenBalances map[string]int) (reserveAmount *big.Int, err error) {
	args := mas.Called(assetId, mirrorNode, isNative, hederaTokenBalances)
	return args.Get(0).(*big.Int), args.Error(1)
}

// FetchEvmFungibleReserveAmount Gets EVM's Fungible Token Reserve Amount
func (mas *MockAssetsService) FetchEvmFungibleReserveAmount(networkId uint64, assetAddress string, isNative bool, evmTokenClient client.EvmFungibleToken, routerContractAddress string) (inLowestDenomination *big.Int, err error) {
	args := mas.Called(networkId, assetAddress, isNative, evmTokenClient, routerContractAddress)
	return args.Get(0).(*big.Int), args.Error(1)
}

// FetchEvmNonFungibleReserveAmount Gets EVM's Non-Fungible Token Reserve Amount
func (mas *MockAssetsService) FetchEvmNonFungibleReserveAmount(networkId uint64, assetAddress string, isNative bool, evmTokenClient client.EvmNft, routerContractAddress string) (inLowestDenomination *big.Int, err error) {
	args := mas.Called(networkId, assetAddress, isNative, evmTokenClient, routerContractAddress)
	return args.Get(0).(*big.Int), args.Error(1)
}
