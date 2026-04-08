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
	bridgeAccountId   = "0.0.476139"
	reserveAmount     = big.NewInt(100)

	////////////////////////
	// Network 0 (Hedera) //
	////////////////////////

	// Native Tokens //

	// Fungible

	hbarCoinGeckoId                  = "hedera-hashgraph"
	hbarCoinMarketCapId              = "4642"
	networkHederaFungibleNativeToken = constants.Hbar

	networkHederaFungibleNativeTokenFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          networkHederaFungibleNativeToken,
		Symbol:        networkHederaFungibleNativeToken,
		Decimals:      constants.HederaDefaultDecimals,
		IsNative:      true,
		ReserveAmount: reserveAmount,
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
	networkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo = &asset.FungibleAssetInfo{
		Name:          networkEthereumFungibleWrappedTokenForNetworkHedera,
		Symbol:        networkEthereumFungibleWrappedTokenForNetworkHedera,
		Decimals:      constants.EvmDefaultDecimals,
		ReserveAmount: reserveAmount,
	}

	parserBridge = parser.Bridge{
		TopicId: topicId,
		Networks: parser.Networks{
			Hedera: map[uint64]*parser.HederaNetwork{
				constants.HederaNetworkId: {
					Name:          "Hedera",
					BridgeAccount: bridgeAccountId,
					PayerAccount:  "0.0.476139",
					Members:       []string{"0.0.123", "0.0.321", "0.0.231"},
				},
			},
			EVM: map[uint64]*parser.EVMNetwork{
				ethereumNetworkId: {
					Name: "Ethereum",
				},
			},
		},
		RegularTokens: map[string]*parser.RegularToken{
			networkHederaFungibleNativeToken: {
				NativeChain:     constants.HederaNetworkId,
				CoinGeckoId:     hbarCoinGeckoId,
				CoinMarketCapId: hbarCoinMarketCapId,
				FeePercentage:   feePercentage,
				AddressesPerNetwork: map[uint64]string{
					ethereumNetworkId: networkEthereumFungibleWrappedTokenForNetworkHedera,
				},
			},
			networkEthereumFungibleNativeToken: {
				NativeChain:         ethereumNetworkId,
				Address:             &networkEthereumFungibleNativeToken,
				CoinGeckoId:         ethereumCoinGeckoId,
				CoinMarketCapId:     ethereumCoinMarketCapId,
				FeePercentage:       feePercentage,
				AddressesPerNetwork: map[uint64]string{},
			},
		},
		MonitoredAccounts: make(map[string]string),
	}
)

func Test_NewBridge(t *testing.T) {
	bridge := NewBridge(parserBridge)

	assert.NotNil(t, bridge)
}

// Suppress unused variable warnings for variables used only in removed test
var (
	_ = networkHederaFungibleNativeTokenFungibleAssetInfo
	_ = networkEthereumFungibleNativeAsset
	_ = networkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo
)
