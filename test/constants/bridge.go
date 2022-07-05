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
	"math/big"
	"strconv"

	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"

	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	constants "github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
)

var (

	////////////
	// Common //
	////////////

	FeePercentage                 = int64(10000)
	MinFeeAmountInUsd             = decimal.NewFromFloat(1)
	TopicId                       = "0.0.1234567"
	BridgeAccountId               = "0.0.476139"
	ReserveAmount                 = int64(100)
	ReserveAmountStr              = strconv.FormatInt(ReserveAmount, 10)
	ReserveAmountBigInt           = big.NewInt(ReserveAmount)
	ReserveAmountWrappedNFTBigInt = big.NewInt(0)

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
	NetworkHederaFungibleNativeTokenFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkHederaFungibleNativeToken,
		Symbol:        NetworkHederaFungibleNativeToken,
		Decimals:      constants.HederaDefaultDecimals,
		IsNative:      true,
		ReserveAmount: ReserveAmountBigInt,
	}

	// Non-Fungible

	NetworkHederaNonFungibleNativeToken = "0.0.111122"
	NetworkHederaNonFungibleNativeAsset = &asset.NativeAsset{
		ChainId: constants.HederaNetworkId,
		Asset:   NetworkHederaNonFungibleNativeToken,
	}
	NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo = &asset.NonFungibleAssetInfo{
		Name:          NetworkHederaNonFungibleNativeToken,
		Symbol:        NetworkHederaNonFungibleNativeToken,
		IsNative:      true,
		ReserveAmount: ReserveAmountBigInt,
		CustomFees: asset.CustomFees{
			CreatedTimestamp: "",
			RoyaltyFees:      make([]asset.RoyaltyFee, 0),
		},
	}

	// Wrapped Tokens //

	NetworkHederaFungibleWrappedTokenForNetworkPolygon                  = "0.0.000033"
	NetworkHederaFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkHederaFungibleWrappedTokenForNetworkPolygon,
		Symbol:        NetworkHederaFungibleWrappedTokenForNetworkPolygon,
		Decimals:      constants.HederaDefaultDecimals,
		ReserveAmount: ReserveAmountBigInt,
	}

	//////////////////////
	// Ethereum Network //
	//////////////////////

	EthereumNetworkId             = uint64(1)
	EthereumRouterContractAddress = "0xb083879B1e10C8476802016CB12cd2F25a000000"

	// Native Tokens //

	// Fungible

	NetworkEthereumFungibleNativeToken = "0xb083879B1e10C8476802016CB12cd2F25a896691"
	NetworkEthereumFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &MinFeeAmountInUsd,
		ChainId:           EthereumNetworkId,
		Asset:             NetworkEthereumFungibleNativeToken,
		FeePercentage:     FeePercentage,
	}

	NetworkEthereumFungibleNativeTokenFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkEthereumFungibleNativeToken,
		Symbol:        NetworkEthereumFungibleNativeToken,
		Decimals:      constants.EvmDefaultDecimals,
		IsNative:      true,
		ReserveAmount: ReserveAmountBigInt,
	}

	// Non-Fungible

	NetworkEthereumNFTWrappedTokenForNetworkHedera = "0x0000000000000000000000000000000000009999"

	// Wrapped Tokens //

	NetworkEthereumFungibleWrappedTokenForNetworkPolygon                  = "0x0000000000000000000000000000000000000133"
	NetworkEthereumFungibleWrappedTokenForNetworkPolygonFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
		Symbol:        NetworkEthereumFungibleWrappedTokenForNetworkPolygon,
		Decimals:      constants.EvmDefaultDecimals,
		ReserveAmount: ReserveAmountBigInt,
	}

	NetworkEthereumFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000555"
	NetworkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkEthereumFungibleWrappedTokenForNetworkHedera,
		Symbol:        NetworkEthereumFungibleWrappedTokenForNetworkHedera,
		Decimals:      constants.EvmDefaultDecimals,
		ReserveAmount: ReserveAmountBigInt,
	}

	/////////////////////
	// Polygon Network //
	/////////////////////

	PolygonNetworkId             = uint64(137)
	PolygonRouterContractAddress = "0xb083879B1e10C8476802016CB12cd2F25a000001"

	// Native Tokens //

	// Fungible

	NetworkPolygonFungibleNativeToken = "0x0000000000000000000000000000000000000033"
	NetworkPolygonFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &MinFeeAmountInUsd,
		ChainId:           PolygonNetworkId,
		Asset:             NetworkPolygonFungibleNativeToken,
		FeePercentage:     FeePercentage,
	}
	NetworkPolygonFungibleNativeTokenFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkPolygonFungibleNativeToken,
		Symbol:        NetworkPolygonFungibleNativeToken,
		Decimals:      constants.EvmDefaultDecimals,
		IsNative:      true,
		ReserveAmount: ReserveAmountBigInt,
	}

	// Wrapped Tokens //

	// Fungible

	NetworkPolygonFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000001"
	NetworkPolygonFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkPolygonFungibleWrappedTokenForNetworkHedera,
		Symbol:        NetworkPolygonFungibleWrappedTokenForNetworkHedera,
		Decimals:      constants.EvmDefaultDecimals,
		ReserveAmount: ReserveAmountBigInt,
	}

	NetworkPolygonFungibleWrappedTokenForNetworkEthereum                  = "0x0000000000000000000000000000000000000123"
	NetworkPolygonFungibleWrappedTokenForNetworkEthereumFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
		Symbol:        NetworkPolygonFungibleWrappedTokenForNetworkEthereum,
		Decimals:      constants.EvmDefaultDecimals,
		ReserveAmount: ReserveAmountBigInt,
	}

	// Non-Fungible

	NetworkPolygonWrappedNonFungibleTokenForHedera                     = "0x0000000000000000000000000000000011111122"
	NetworkPolygonWrappedNonFungibleTokenForHederaNonFungibleAssetInfo = &asset.NonFungibleAssetInfo{
		Name:          NetworkPolygonWrappedNonFungibleTokenForHedera,
		Symbol:        NetworkPolygonWrappedNonFungibleTokenForHedera,
		ReserveAmount: ReserveAmountWrappedNFTBigInt,
	}

	Networks = map[uint64]*parser.Network{
		constants.HederaNetworkId: {
			Name:          "Hedera",
			BridgeAccount: BridgeAccountId,
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
			Name:                  "Ethereum",
			RouterContractAddress: EthereumRouterContractAddress,
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
			Name:                  "Polygon",
			RouterContractAddress: PolygonRouterContractAddress,
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

	FungibleAssetInfos = map[uint64]map[string]*asset.FungibleAssetInfo{
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

	NonFungibleAssetInfos = map[uint64]map[string]*asset.NonFungibleAssetInfo{
		constants.HederaNetworkId: {
			NetworkHederaNonFungibleNativeToken: NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo,
		},
		PolygonNetworkId: {
			NetworkPolygonWrappedNonFungibleTokenForHedera: NetworkPolygonWrappedNonFungibleTokenForHederaNonFungibleAssetInfo,
		},
	}

	ParserBridge = parser.Bridge{
		UseLocalConfig:    true,
		TopicId:           TopicId,
		Networks:          Networks,
		MonitoredAccounts: make(map[string]string),
	}

	HederaNftFees = map[string]int64{
		NetworkHederaNonFungibleNativeToken: 1000,
	}

	PaymentTokens = map[uint64]string{
		PolygonNetworkId: "0x0000000000000000000000000000000000006655",
	}

	NftFeesForApi = map[uint64]map[string]pricing.NonFungibleFee{
		constants.HederaNetworkId: {
			NetworkHederaNonFungibleNativeToken: pricing.NonFungibleFee{
				IsNative:     true,
				PaymentToken: constants.Hbar,
				Fee:          decimal.NewFromInt(HederaNftFees[NetworkHederaNonFungibleNativeToken]),
			},
		},
		PolygonNetworkId: {
			NetworkPolygonWrappedNonFungibleTokenForHedera: pricing.NonFungibleFee{
				IsNative:     false,
				PaymentToken: PaymentTokens[PolygonNetworkId],
				Fee:          decimal.NewFromBigInt(big.NewInt(10000000000), 0),
			},
		},
	}
)
