package service

import (
	assetModel "github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
)

type Assets interface {
	// GetFungibleNetworkAssets Gets all Fungible Assets by Network ID
	GetFungibleNetworkAssets() map[uint64][]string
	// GetNativeToWrappedAssets Gets all Native assets with their Wrapped assets by Network ID
	GetNativeToWrappedAssets() map[uint64]map[string]map[uint64]string
	// WrappedFromNative Gets All Wrapped Assets for passed Native Asset's address
	WrappedFromNative(nativeChainId uint64, nativeAssetAddress string) map[uint64]string
	// NativeToWrapped Gets Wrapped Asset for passed Native Asset's address and Target Chain ID
	NativeToWrapped(nativeAssetAddress string, nativeChainId, targetChainId uint64) string
	// WrappedToNative Gets Native Asset from passed Wrapped Asset's address and Wrapped Chain ID
	WrappedToNative(wrappedAssetAddress string, wrappedChainId uint64) *assetModel.NativeAsset
	// FungibleNetworkAssets Gets all Fungible assets for passed Chain ID
	FungibleNetworkAssets(chainId uint64) []string
	// FungibleNativeAsset Gets NativeAsset for passed chainId and assetAddress
	FungibleNativeAsset(chainId uint64, nativeAssetAddress string) *assetModel.NativeAsset
	// IsNative Returns flag showing if the passed Asset's Address is Native for the passed Chain ID
	IsNative(chainId uint64, assetAddress string) bool
	// GetOppositeAsset Gets Opposite asset for passed chain IDs and assetAddress
	GetOppositeAsset(sourceChainId uint64, targetChainId uint64, assetAddress string) string
	// GetFungibleAssetInfo Gets FungibleAssetInfo
	GetFungibleAssetInfo(networkId uint64, assetAddress string) (assetInfo assetModel.FungibleAssetInfo, exist bool)
}
