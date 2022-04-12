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

package config

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var (
	////////////
	// Common //
	////////////

	ethereumNetworkId = uint64(1)
	feePercentage     = int64(10000)
	minFeeAmountInUsd = decimal.NewFromFloat(1)
	topicId           = "0.0.1234567"

	////////////////////////
	// Network 0 (Hedera) //
	////////////////////////

	// Native Tokens //

	// Fungible

	hbarCoinGeckoId                  = "hedera-hashgraph"
	hbarCoinMarketCapId              = "4642"
	networkHederaFungibleNativeToken = constants.Hbar

	// Non-Fungible

	networkHederaNFTNativeToken = "0.0.111122"

	networkHederaFungibleNativeTokenFungibleAssetInfo = asset.FungibleAssetInfo{
		networkHederaFungibleNativeToken,
		networkHederaFungibleNativeToken,
		constants.HederaDefaultDecimals,
		true,
	}

	//////////////////////
	// Ethereum Network //
	//////////////////////

	// Native Tokens //

	// Fungible

	ethereumCoinGeckoId                = "ethereum"
	ethereumCoinMarketCapId            = "1027"
	networkEthereumFungibleNativeToken = "0xb083879B1e10C8476802016CB12cd2F25a896691"
	networkEthereumFungibleNativeAsset = &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           ethereumNetworkId,
		Asset:             networkEthereumFungibleNativeToken,
		FeePercentage:     feePercentage,
	}

	// Wrapped Tokens //

	networkEthereumFungibleWrappedTokenForNetworkHedera                  = "0x0000000000000000000000000000000000000555"
	networkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = asset.FungibleAssetInfo{
		networkEthereumFungibleWrappedTokenForNetworkHedera,
		networkEthereumFungibleWrappedTokenForNetworkHedera,
		constants.EvmDefaultDecimals,
		false,
	}

	networks = map[uint64]*parser.Network{
		constants.HederaNetworkId: {
			Name:          "Hedera",
			BridgeAccount: "0.0.476139",
			PayerAccount:  "0.0.476139",
			Members:       []string{"0.0.123", "0.0.321", "0.0.231"},
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					networkHederaFungibleNativeToken: {
						Networks: map[uint64]string{
							ethereumNetworkId: networkEthereumFungibleWrappedTokenForNetworkHedera,
						},
						CoinGeckoId:       hbarCoinGeckoId,
						CoinMarketCapId:   hbarCoinMarketCapId,
						MinFeeAmountInUsd: minFeeAmountInUsd.String(),
						MinAmount:         big.NewInt(1000000),
					},
				},
				Nft: map[string]parser.Token{
					networkHederaNFTNativeToken: {
						Networks: map[uint64]string{},
						Fee:      feePercentage,
					},
				},
			},
		},
		ethereumNetworkId: {
			Name: "Ethereum",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					networkEthereumFungibleNativeToken: {
						Networks:          map[uint64]string{},
						CoinGeckoId:       ethereumCoinGeckoId,
						CoinMarketCapId:   ethereumCoinMarketCapId,
						MinFeeAmountInUsd: minFeeAmountInUsd.String(),
					},
				},
				Nft: nil,
			},
		},
	}
	parserBridge = parser.Bridge{
		TopicId:           topicId,
		Networks:          networks,
		MonitoredAccounts: make(map[string]string),
	}
)

func Test_NewBridge(t *testing.T) {
	bridge := NewBridge(parserBridge)

	assert.NotNil(t, bridge)
}

func Test_LoadStaticMinAmountsForWrappedFungibleTokens(t *testing.T) {
	mocks.Setup()
	bridge := NewBridge(parserBridge)
	mocks.MAssetsService.On("FungibleAssetInfo", constants.HederaNetworkId, networkHederaFungibleNativeToken).Return(networkHederaFungibleNativeTokenFungibleAssetInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", ethereumNetworkId, networkEthereumFungibleNativeToken).Return(networkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", ethereumNetworkId, networkEthereumFungibleWrappedTokenForNetworkHedera).Return(networkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo, true)

	bridge.LoadStaticMinAmountsForWrappedFungibleTokens(parserBridge, mocks.MAssetsService)

	assert.NotNil(t, bridge)
	mocks.MAssetsService.AssertCalled(t, "FungibleAssetInfo", constants.HederaNetworkId, networkHederaFungibleNativeToken)
	mocks.MAssetsService.AssertCalled(t, "FungibleAssetInfo", ethereumNetworkId, networkEthereumFungibleNativeToken)
	mocks.MAssetsService.AssertCalled(t, "FungibleAssetInfo", ethereumNetworkId, networkEthereumFungibleWrappedTokenForNetworkHedera)
}
