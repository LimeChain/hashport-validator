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
	TopicId           = "0.0.1234567"

	////////////////////////
	// Network 0 (Hedera) //
	////////////////////////

	// Native Tokens //

	// Fungible

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
		true,
	}

	// Non-Fungible

	NetworkHederaNonFungibleNativeToken = "0.0.111122"
	NetworkHederaNonFungibleNativeAsset = &asset.NativeAsset{
		ChainId: constants.HederaNetworkId,
		Asset:   NetworkHederaNonFungibleNativeToken,
	}
	NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo = asset.NonFungibleAssetInfo{
		NetworkHederaNonFungibleNativeToken,
		NetworkHederaNonFungibleNativeToken,
		true,
	}

	// Wrapped Tokens //

	NetworkHederaFungibleWrappedTokenForNetworkPolygon                  = "0.0.000033"
	NetworkHederaFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkHederaFungibleWrappedTokenForNetworkPolygon,
		NetworkHederaFungibleWrappedTokenForNetworkPolygon,
		constants.HederaDefaultDecimals,
		false,
	}

	//////////////////////
	// Ethereum Network //
	//////////////////////

	// Native Tokens //

	// Fungible

	NetworkEthereumFungibleNativeToken = "0xb083879B1e10C8476802016CB12cd2F25a896691"
	NetworkEthereumFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &MinFeeAmountInUsd,
		ChainId:           EthereumNetworkId,
		Asset:             NetworkEthereumFungibleNativeToken,
		FeePercentage:     FeePercentage,
	}

	// Non-Fungible

	NetworkEthereumNFTWrappedTokenForNetworkHedera = "0x0000000000000000000000000000000000009999"

	NetworkEthereumFungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkEthereumFungibleNativeToken,
		NetworkEthereumFungibleNativeToken,
		constants.EvmDefaultDecimals,
		true,
	}

	// Wrapped Tokens //

	NetworkEthereumFungibleWrappedTokenForNetworkPolygon                  = "0x0000000000000000000000000000000000000133"
	NetworkEthereumFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
		NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
		constants.EvmDefaultDecimals,
		false,
	}

	NetworkEthereumFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000555"
	NetworkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkEthereumFungibleWrappedTokenForNetworkHedera,
		NetworkEthereumFungibleWrappedTokenForNetworkHedera,
		constants.EvmDefaultDecimals,
		false,
	}

	/////////////////////
	// Polygon Network //
	/////////////////////

	// Native Tokens //

	// Fungible

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
		true,
	}

	// Wrapped Tokens //

	// Fungible

	NetworkPolygonFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000001"
	NetworkPolygonFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkPolygonFungibleWrappedTokenForNetworkHedera,
		NetworkPolygonFungibleWrappedTokenForNetworkHedera,
		constants.EvmDefaultDecimals,
		false,
	}
	NetworkPolygonFungibleWrappedTokenForNetworkEthereum                  = "0x0000000000000000000000000000000000000123"
	NetworkPolygonFungibleWrappedTokenForNetworkEthereumFungibleAssetInfo = asset.FungibleAssetInfo{
		NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
		NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
		constants.EvmDefaultDecimals,
		false,
	}

	// Non-Fungible

	NetworkPolygonWrappedNonFungibleTokenForHedera                     = "0x0000000000000000000000000000000011111122"
	NetworkPolygonWrappedNonFungibleTokenForHederaNonFungibleAssetInfo = asset.NonFungibleAssetInfo{
		NetworkPolygonWrappedNonFungibleTokenForHedera,
		NetworkPolygonWrappedNonFungibleTokenForHedera,
		false,
	}

	Networks = map[uint64]*parser.Network{
		constants.HederaNetworkId: {
			Name:          "Hedera",
			BridgeAccount: "0.0.476139",
			PayerAccount:  "0.0.476139",
			Members:       []string{"0.0.123", "0.0.321", "0.0.231"},
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					NetworkHederaFungibleNativeToken: {
						Networks: map[uint64]string{
							PolygonNetworkId:  NetworkPolygonFungibleWrappedTokenForNetworkHedera,
							EthereumNetworkId: NetworkEthereumFungibleWrappedTokenForNetworkHedera,
						},
						CoinGeckoId:       HbarCoinGeckoId,
						CoinMarketCapId:   HbarCoinMarketCapId,
						MinFeeAmountInUsd: MinFeeAmountInUsd.String(),
					},
				},
				Nft: map[string]parser.Token{
					NetworkHederaNonFungibleNativeToken: {
						Networks: map[uint64]string{PolygonNetworkId: NetworkPolygonWrappedNonFungibleTokenForHedera},
					},
				},
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
						CoinGeckoId:       EthereumCoinGeckoId,
						CoinMarketCapId:   EthereumCoinMarketCapId,
						MinFeeAmountInUsd: MinFeeAmountInUsd.String(),
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
						MinFeeAmountInUsd: MinFeeAmountInUsd.String(),
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
			NetworkHederaNonFungibleNativeToken: {
				PolygonNetworkId: NetworkPolygonWrappedNonFungibleTokenForHedera,
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
			NetworkPolygonWrappedNonFungibleTokenForHedera:       NetworkHederaNonFungibleNativeAsset,
		},
	}

	// Fungible Assets //

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

	// Non-Fungible Assets //

	NonFungibleNetworkAssets = map[uint64][]string{
		constants.HederaNetworkId: {NetworkHederaNonFungibleNativeToken},
		PolygonNetworkId:          {NetworkPolygonWrappedNonFungibleTokenForHedera},
	}

	NonFungibleAssetInfos = map[uint64]map[string]asset.NonFungibleAssetInfo{
		constants.HederaNetworkId: {
			NetworkHederaNonFungibleNativeToken: NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo,
		},
		PolygonNetworkId: {
			NetworkPolygonWrappedNonFungibleTokenForHedera: NetworkPolygonWrappedNonFungibleTokenForHederaNonFungibleAssetInfo,
		},
	}

	ParserBridge = parser.Bridge{
		TopicId:           TopicId,
		Networks:          Networks,
		MonitoredAccounts: make(map[string]string),
	}

	HederaNftFees = map[string]int64{
		NetworkHederaNonFungibleNativeToken: 1000,
	}
)
