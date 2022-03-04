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
	assetModel "github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/stretchr/testify/mock"
)

type MockAssetsService struct {
	mock.Mock
}

// GetFungibleNetworkAssets Gets all Fungible Assets by Network ID
func (mas *MockAssetsService) GetFungibleNetworkAssets() map[uint64][]string {
	args := mas.Called()
	result := args.Get(0).(map[uint64][]string)
	return result
}

// GetNativeToWrappedAssets Gets all Native assets with their Wrapped assets by Network ID
func (mas *MockAssetsService) GetNativeToWrappedAssets() map[uint64]map[string]map[uint64]string {
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

// FungibleNetworkAssets Gets all Fungible assets for passed Chain ID
func (mas *MockAssetsService) FungibleNetworkAssets(chainId uint64) []string {
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

// GetOppositeAsset Gets Opposite asset for passed chain IDs and assetAddress
func (mas *MockAssetsService) GetOppositeAsset(sourceChainId uint64, targetChainId uint64, assetAddress string) string {
	args := mas.Called(sourceChainId, targetChainId, assetAddress)
	result := args.Get(0).(string)
	return result
}

// GetFungibleAssetInfo Gets FungibleAssetInfo
func (mas *MockAssetsService) GetFungibleAssetInfo(networkId uint64, assetAddress string) (assetInfo assetModel.FungibleAssetInfo, exist bool) {
	args := mas.Called(networkId, assetAddress)
	assetInfo = args.Get(0).(assetModel.FungibleAssetInfo)
	exist = args.Get(1).(bool)

	return assetInfo, exist
}
