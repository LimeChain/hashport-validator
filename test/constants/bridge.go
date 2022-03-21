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

package constants

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
)

var (

	////////////
	// Common //
	////////////

	EthereumNetworkId = uint64(1)
	PolygonNetworkId  = uint64(137)
	FeePercentage     = int64(10000)
	MinFeeAmountInUsd = decimal.NewFromFloat(1)

	////////////////////////
	// Network 0 (Hedera) //
	////////////////////////

	// Native Tokens //

	NetworkHederaFungibleNativeToken = constants.Hbar
	NetworkHederaFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &MinFeeAmountInUsd,
		ChainId:           constants.HederaNetworkId,
		Asset:             NetworkHederaFungibleNativeToken,
		FeePercentage:     FeePercentage,
	}

	NetworkHederaFungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkHederaFungibleNativeToken,
		NetworkHederaFungibleNativeToken,
		constants.HederaDefaultDecimals,
	}

	// Wrapped Tokens //

	NetworkHederaFungibleWrappedTokenForNetworkPolygon                  = "0.0.000033"
	NetworkHederaFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkHederaFungibleWrappedTokenForNetworkPolygon,
		NetworkHederaFungibleWrappedTokenForNetworkPolygon,
		constants.HederaDefaultDecimals,
	}

	//////////////////////
	// Ethereum Network //
	//////////////////////

	// Native Tokens //

	NetworkEthereumFungibleNativeToken = "0xb083879B1e10C8476802016CB12cd2F25a896691"
	NetworkEthereumFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &MinFeeAmountInUsd,
		ChainId:           EthereumNetworkId,
		Asset:             NetworkEthereumFungibleNativeToken,
		FeePercentage:     FeePercentage,
	}

	NetworkEthereumFungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkEthereumFungibleNativeToken,
		NetworkEthereumFungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	// Wrapped Tokens //

	NetworkEthereumFungibleWrappedTokenForNetworkPolygon                  = "0x0000000000000000000000000000000000000133"
	NetworkEthereumFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
		NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
		constants.EvmDefaultDecimals,
	}

	NetworkEthereumFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000555"
	NetworkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkEthereumFungibleWrappedTokenForNetworkHedera,
		NetworkEthereumFungibleWrappedTokenForNetworkHedera,
		constants.EvmDefaultDecimals,
	}

	/////////////////////
	// Polygon Network //
	/////////////////////

	// Native Tokens //

	NetworkPolygonFungibleNativeToken = "0x0000000000000000000000000000000000000033"
	NetworkPolygonFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &MinFeeAmountInUsd,
		ChainId:           PolygonNetworkId,
		Asset:             NetworkPolygonFungibleNativeToken,
		FeePercentage:     FeePercentage,
	}
	NetworkPolygonFungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkPolygonFungibleNativeToken,
		NetworkPolygonFungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	// Wrapped Tokens //

	NetworkPolygonFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000001"
	NetworkPolygonFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkPolygonFungibleWrappedTokenForNetworkHedera,
		NetworkPolygonFungibleWrappedTokenForNetworkHedera,
		constants.EvmDefaultDecimals,
	}
	NetworkPolygonFungibleWrappedTokenForNetworkEthereum                  = "0x0000000000000000000000000000000000000123"
	NetworkPolygonFungibleWrappedTokenForNetworkEthereumFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
		NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
		constants.EvmDefaultDecimals,
	}

	Networks = map[uint64]*parser.Network{
		constants.HederaNetworkId: {
			Name: "Hedera",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					NetworkHederaFungibleNativeToken: {
						Networks: map[uint64]string{
							PolygonNetworkId:  NetworkPolygonFungibleWrappedTokenForNetworkHedera,
							EthereumNetworkId: NetworkEthereumFungibleWrappedTokenForNetworkHedera,
						},
					},
				},
				Nft: nil,
			},
		},
		EthereumNetworkId: {
			Name: "Ethereum",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					NetworkEthereumFungibleNativeToken: {
						Networks: map[uint64]string{
							PolygonNetworkId: NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
						},
					},
				},
				Nft: nil,
			},
		},
		PolygonNetworkId: {
			Name: "Polygon",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					NetworkPolygonFungibleNativeToken: {
						Networks: map[uint64]string{
							constants.HederaNetworkId: NetworkHederaFungibleWrappedTokenForNetworkPolygon,
							EthereumNetworkId:         NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
						},
					},
				},
				Nft: nil,
			},
		},
	}

	NativeToWrapped = map[uint64]map[string]map[uint64]string{
		constants.HederaNetworkId: {
			NetworkHederaFungibleNativeToken: {
				PolygonNetworkId:  NetworkPolygonFungibleWrappedTokenForNetworkHedera,
				EthereumNetworkId: NetworkEthereumFungibleWrappedTokenForNetworkHedera,
			},
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken: {
				PolygonNetworkId: NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
			},
		},
		PolygonNetworkId: {
			NetworkPolygonFungibleNativeToken: {
				constants.HederaNetworkId: NetworkHederaFungibleWrappedTokenForNetworkPolygon,
				EthereumNetworkId:         NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
			},
		},
	}

	WrappedToNative = map[uint64]map[string]*asset.NativeAsset{
		constants.HederaNetworkId: {
			NetworkHederaFungibleWrappedTokenForNetworkPolygon: NetworkPolygonFungibleNativeAsset,
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleWrappedTokenForNetworkPolygon: NetworkPolygonFungibleNativeAsset,
			NetworkEthereumFungibleWrappedTokenForNetworkHedera:  NetworkHederaFungibleNativeAsset,
		},
		PolygonNetworkId: {
			NetworkPolygonFungibleWrappedTokenForNetworkHedera:   NetworkHederaFungibleNativeAsset,
			NetworkPolygonFungibleWrappedTokenForNetworkEthereum: NetworkEthereumFungibleNativeAsset,
		},
	}

	FungibleNetworkAssets = map[uint64][]string{
		constants.HederaNetworkId: {NetworkHederaFungibleNativeToken, NetworkHederaFungibleWrappedTokenForNetworkPolygon},
		EthereumNetworkId:         {NetworkEthereumFungibleNativeToken, NetworkEthereumFungibleWrappedTokenForNetworkPolygon, NetworkEthereumFungibleWrappedTokenForNetworkHedera},
		PolygonNetworkId:          {NetworkPolygonFungibleNativeToken, NetworkPolygonFungibleWrappedTokenForNetworkEthereum, NetworkPolygonFungibleWrappedTokenForNetworkHedera},
	}

	FungibleNativeAssets = map[uint64]map[string]*asset.NativeAsset{
		constants.HederaNetworkId: {
			NetworkHederaFungibleNativeToken: NetworkHederaFungibleNativeAsset,
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken: NetworkEthereumFungibleNativeAsset,
		},
		PolygonNetworkId: {
			NetworkPolygonFungibleNativeToken: NetworkPolygonFungibleNativeAsset,
		},
	}

	FungibleAssetInfos = map[uint64]map[string]asset.FungibleAssetInfo{
		constants.HederaNetworkId: {
			NetworkHederaFungibleNativeToken:                   NetworkHederaFungibleNativeTokenFungibleAssetInfo,
			NetworkHederaFungibleWrappedTokenForNetworkPolygon: NetworkHederaFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo,
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken:                   NetworkEthereumFungibleNativeTokenFungibleAssetInfo,
			NetworkEthereumFungibleWrappedTokenForNetworkPolygon: NetworkEthereumFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo,
			NetworkEthereumFungibleWrappedTokenForNetworkHedera:  NetworkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo,
		},
		PolygonNetworkId: {
			NetworkPolygonFungibleNativeToken:                    NetworkPolygonFungibleNativeTokenFungibleAssetInfo,
			NetworkPolygonFungibleWrappedTokenForNetworkHedera:   NetworkPolygonFungibleWrappedTokenForNetworkHederaFungibleAssetInfo,
			NetworkPolygonFungibleWrappedTokenForNetworkEthereum: NetworkPolygonFungibleWrappedTokenForNetworkEthereumFungibleAssetInfo,
		},
	}
)
