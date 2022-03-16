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
	minFeeAmountInUsd = decimal.NewFromFloat(0.0)

	////////////////////////
	// Network 0 (Hedera) //
	////////////////////////

	// Native Tokens //

	Network0FungibleNativeToken = constants.Hbar
	Network0FungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           0,
		Asset:             Network0FungibleNativeToken,
		FeePercentage:     0,
	}

	Network0FungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		Network0FungibleNativeToken,
		Network0FungibleNativeToken,
		constants.HederaDefaultDecimals,
	}

	// Wrapped Tokens //

	Network0FungibleWrappedTokenForNetwork2                  = "0.0.000002"
	Network0FungibleWrappedTokenForNetwork2FungibleAssetInfo = asset.FungibleAssetInfo{
		Network0FungibleWrappedTokenForNetwork2,
		Network0FungibleWrappedTokenForNetwork2,
		constants.HederaDefaultDecimals,
	}

	Network0FungibleWrappedTokenForNetwork3                  = "0.0.000003"
	Network0FungibleWrappedTokenForNetwork3FungibleAssetInfo = asset.FungibleAssetInfo{
		Network0FungibleWrappedTokenForNetwork3,
		Network0FungibleWrappedTokenForNetwork3,
		constants.HederaDefaultDecimals,
	}

	Network0FungibleWrappedTokenForNetwork32                  = "0.0.000032"
	Network0FungibleWrappedTokenForNetwork32FungibleAssetInfo = asset.FungibleAssetInfo{
		Network0FungibleWrappedTokenForNetwork32,
		Network0FungibleWrappedTokenForNetwork32,
		constants.HederaDefaultDecimals,
	}

	Network0FungibleWrappedTokenForNetwork33                  = "0.0.000033"
	Network0FungibleWrappedTokenForNetwork33FungibleAssetInfo = asset.FungibleAssetInfo{
		Network0FungibleWrappedTokenForNetwork33,
		Network0FungibleWrappedTokenForNetwork33,
		constants.HederaDefaultDecimals,
	}

	////////////////
	// Network 1 //
	////////////////

	// Native Tokens //

	Network1FungibleNativeToken = "0xb083879B1e10C8476802016CB12cd2F25a896691"
	Network1FungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           1,
		Asset:             Network1FungibleNativeToken,
		FeePercentage:     0,
	}

	Network1FungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		Network1FungibleNativeToken,
		Network1FungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	// Wrapped Tokens //

	Network1FungibleWrappedTokenForNetwork33                  = "0x0000000000000000000000000000000000000133"
	Network1FungibleWrappedTokenForNetwork33FungibleAssetInfo = asset.FungibleAssetInfo{
		Network1FungibleWrappedTokenForNetwork33,
		Network1FungibleWrappedTokenForNetwork33,
		constants.EvmDefaultDecimals,
	}

	////////////////
	// Network 2 //
	////////////////

	// Native Tokens //

	Network2FungibleNativeToken = "0x0000000000000000000000000000000000000002"
	Network2FungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           2,
		Asset:             Network2FungibleNativeToken,
		FeePercentage:     0,
	}
	Network2FungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		Network2FungibleNativeToken,
		Network2FungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	////////////////
	// Network 3 //
	////////////////

	// Native Tokens //

	Network3FungibleNativeToken = "0x0000000000000000000000000000000000000003"
	Network3FungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           3,
		Asset:             Network3FungibleNativeToken,
		FeePercentage:     0,
	}
	Network3FungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		Network3FungibleNativeToken,
		Network3FungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	////////////////
	// Network 32 //
	////////////////

	// Native Tokens //

	Network32FungibleNativeToken = "0x0000000000000000000000000000000000000032"
	Network32FungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           32,
		Asset:             Network32FungibleNativeToken,
		FeePercentage:     0,
	}
	Network32FungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		Network32FungibleNativeToken,
		Network32FungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	////////////////
	// Network 33 //
	////////////////

	// Native Tokens //

	Network33FungibleNativeToken = "0x0000000000000000000000000000000000000033"
	Network33FungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           33,
		Asset:             Network33FungibleNativeToken,
		FeePercentage:     0,
	}
	Network33FungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		Network33FungibleNativeToken,
		Network33FungibleNativeToken,
		constants.EvmDefaultDecimals,
	}

	// Wrapped Tokens //

	Network33FungibleWrappedTokenForNetwork0                  = "0x0000000000000000000000000000000000000001"
	Network33FungibleWrappedTokenForNetwork0FungibleAssetInfo = asset.FungibleAssetInfo{
		Network33FungibleWrappedTokenForNetwork0,
		Network33FungibleWrappedTokenForNetwork0,
		constants.EvmDefaultDecimals,
	}
	Network33FungibleWrappedTokenForNetwork1                  = "0x0000000000000000000000000000000000000123"
	Network33FungibleWrappedTokenForNetwork1FungibleAssetInfo = asset.FungibleAssetInfo{
		Network33FungibleWrappedTokenForNetwork1,
		Network33FungibleWrappedTokenForNetwork1,
		constants.EvmDefaultDecimals,
	}

	Networks = map[uint64]*parser.Network{
		0: {
			Name: "Hedera",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					Network0FungibleNativeToken: {
						Networks: map[uint64]string{
							33: Network33FungibleWrappedTokenForNetwork0,
						},
					},
				},
				Nft: nil,
			},
		},
		1: {
			Name: "Network1",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					Network1FungibleNativeToken: {
						Networks: map[uint64]string{
							33: Network33FungibleWrappedTokenForNetwork1,
						},
					},
				},
				Nft: nil,
			},
		},
		2: {
			Name: "Network2",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					Network2FungibleNativeToken: {
						Networks: map[uint64]string{
							0: Network0FungibleWrappedTokenForNetwork2,
						},
					},
				},
				Nft: nil,
			},
		},
		3: {
			Name: "Network3",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					Network3FungibleNativeToken: {
						Networks: map[uint64]string{
							0: Network0FungibleWrappedTokenForNetwork3,
						},
					},
				},
				Nft: nil,
			},
		},
		32: {
			Name: "Network32",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					Network32FungibleNativeToken: {
						Networks: map[uint64]string{
							0: Network0FungibleWrappedTokenForNetwork32,
						},
					},
				},
				Nft: nil,
			},
		},
		33: {
			Name: "Network33",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					Network33FungibleNativeToken: {
						Networks: map[uint64]string{
							0: Network0FungibleWrappedTokenForNetwork33,
							1: Network1FungibleWrappedTokenForNetwork33,
						},
					},
				},
				Nft: nil,
			},
		},
	}

	NativeToWrapped = map[uint64]map[string]map[uint64]string{
		0: {
			Network0FungibleNativeToken: {
				33: Network33FungibleWrappedTokenForNetwork0,
			},
		},
		1: {
			Network1FungibleNativeToken: {
				33: Network33FungibleWrappedTokenForNetwork1,
			},
		},
		2: {
			Network2FungibleNativeToken: {
				0: Network0FungibleWrappedTokenForNetwork2,
			},
		},
		3: {
			Network3FungibleNativeToken: {
				0: Network0FungibleWrappedTokenForNetwork3,
			},
		},
		32: {
			Network32FungibleNativeToken: {
				0: Network0FungibleWrappedTokenForNetwork32,
			},
		},

		33: {
			Network33FungibleNativeToken: {
				0: Network0FungibleWrappedTokenForNetwork33,
				1: Network1FungibleWrappedTokenForNetwork33,
			},
		},
	}

	WrappedToNative = map[uint64]map[string]*asset.NativeAsset{
		0: {
			Network0FungibleWrappedTokenForNetwork3:  Network3FungibleNativeAsset,
			Network0FungibleWrappedTokenForNetwork2:  Network2FungibleNativeAsset,
			Network0FungibleWrappedTokenForNetwork33: Network33FungibleNativeAsset,
			Network0FungibleWrappedTokenForNetwork32: Network32FungibleNativeAsset,
		},
		1: {
			Network1FungibleWrappedTokenForNetwork33: Network33FungibleNativeAsset,
		},
		33: {
			Network33FungibleWrappedTokenForNetwork0: Network0FungibleNativeAsset,
			Network33FungibleWrappedTokenForNetwork1: Network1FungibleNativeAsset,
		},
	}

	FungibleNetworkAssets = map[uint64][]string{
		0:  {Network0FungibleNativeToken, Network0FungibleWrappedTokenForNetwork2, Network0FungibleWrappedTokenForNetwork3, Network0FungibleWrappedTokenForNetwork32, Network0FungibleWrappedTokenForNetwork33},
		1:  {Network1FungibleNativeToken, Network1FungibleWrappedTokenForNetwork33},
		2:  {Network2FungibleNativeToken},
		3:  {Network3FungibleNativeToken},
		32: {Network32FungibleNativeToken},
		33: {Network33FungibleNativeToken, Network33FungibleWrappedTokenForNetwork1, Network33FungibleWrappedTokenForNetwork0},
	}

	FungibleNativeAssets = map[uint64]map[string]*asset.NativeAsset{
		0: {
			Network0FungibleNativeToken: Network0FungibleNativeAsset,
		},
		1: {
			Network1FungibleNativeToken: Network1FungibleNativeAsset,
		},
		2: {
			Network2FungibleNativeToken: Network2FungibleNativeAsset,
		},
		3: {
			Network3FungibleNativeToken: Network3FungibleNativeAsset,
		},
		32: {
			Network32FungibleNativeToken: Network32FungibleNativeAsset,
		},
		33: {
			Network33FungibleNativeToken: Network33FungibleNativeAsset,
		},
	}

	FungibleAssetInfos = map[uint64]map[string]asset.FungibleAssetInfo{
		0: {
			Network0FungibleNativeToken:              Network0FungibleNativeTokenFungibleAssetInfo,
			Network0FungibleWrappedTokenForNetwork2:  Network0FungibleWrappedTokenForNetwork2FungibleAssetInfo,
			Network0FungibleWrappedTokenForNetwork3:  Network0FungibleWrappedTokenForNetwork3FungibleAssetInfo,
			Network0FungibleWrappedTokenForNetwork32: Network0FungibleWrappedTokenForNetwork32FungibleAssetInfo,
			Network0FungibleWrappedTokenForNetwork33: Network0FungibleWrappedTokenForNetwork33FungibleAssetInfo,
		},
		1: {
			Network1FungibleNativeToken:              Network1FungibleNativeTokenFungibleAssetInfo,
			Network1FungibleWrappedTokenForNetwork33: Network1FungibleWrappedTokenForNetwork33FungibleAssetInfo,
		},
		2: {
			Network2FungibleNativeToken: Network2FungibleNativeTokenFungibleAssetInfo,
		},
		3: {
			Network3FungibleNativeToken: Network3FungibleNativeTokenFungibleAssetInfo,
		},
		32: {
			Network32FungibleNativeToken: Network32FungibleNativeTokenFungibleAssetInfo,
		},
		33: {
			Network33FungibleNativeToken:             Network33FungibleNativeTokenFungibleAssetInfo,
			Network33FungibleWrappedTokenForNetwork0: Network33FungibleWrappedTokenForNetwork0FungibleAssetInfo,
			Network33FungibleWrappedTokenForNetwork1: Network33FungibleWrappedTokenForNetwork1FungibleAssetInfo,
		},
	}
)
