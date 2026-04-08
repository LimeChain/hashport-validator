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
	"math/big"

	decimalHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

type Bridge struct {
	TopicId             string
	Hedera              *BridgeHedera
	EVMs                map[uint64]BridgeEvm
	CoinMarketCapIds    map[uint64]map[string]string
	CoinGeckoIds        map[uint64]map[string]string
	MinAmounts          map[uint64]map[string]*big.Int
	MonitoredAccounts   map[string]string
	BlacklistedAccounts []string
}

func (b *Bridge) Update(from *Bridge) {
	b.TopicId = from.TopicId
	b.Hedera = from.Hedera
	b.EVMs = from.EVMs
	b.CoinMarketCapIds = from.CoinMarketCapIds
	b.CoinGeckoIds = from.CoinGeckoIds
	b.MinAmounts = from.MinAmounts
	b.MonitoredAccounts = from.MonitoredAccounts
	b.BlacklistedAccounts = from.BlacklistedAccounts
}

type BridgeHedera struct {
	BridgeAccount  string
	PayerAccount   string
	Members        []string
	FeePercentages map[string]int64
	MintCost       float64
	UnlockCost     float64
}

type BridgeEvm struct {
	RouterContractAddress  string
	ContractOperationsCost ContractOperationsCost
}

type ContractOperationsCost struct {
	MintGasUnits   uint64
	UnlockGasUnits uint64
}

func NewBridge(bridge parser.Bridge) *Bridge {
	cfg := &Bridge{
		TopicId:             bridge.TopicId,
		Hedera:              nil,
		EVMs:                make(map[uint64]BridgeEvm),
		MonitoredAccounts:   bridge.MonitoredAccounts,
		BlacklistedAccounts: bridge.BlacklistedAccounts,
		CoinGeckoIds:        make(map[uint64]map[string]string),
		CoinMarketCapIds:    make(map[uint64]map[string]string),
		MinAmounts:          make(map[uint64]map[string]*big.Int),
	}

	for networkId, networkInfo := range bridge.Networks.Hedera {
		constants.HederaNetworkId = networkId
		constants.NetworksByName[networkInfo.Name] = networkId
		constants.NetworksById[networkId] = networkInfo.Name
		cfg.CoinGeckoIds[networkId] = make(map[string]string)
		cfg.CoinMarketCapIds[networkId] = make(map[string]string)
		cfg.MinAmounts[networkId] = make(map[string]*big.Int)

		cfg.Hedera = &BridgeHedera{
			BridgeAccount:  networkInfo.BridgeAccount,
			PayerAccount:   networkInfo.PayerAccount,
			Members:        networkInfo.Members,
			MintCost:       networkInfo.MintCost,
			UnlockCost:     networkInfo.UnlockCost,
			FeePercentages: make(map[string]int64),
		}
	}

	for networkId, networkInfo := range bridge.Networks.EVM {
		constants.NetworksByName[networkInfo.Name] = networkId
		constants.NetworksById[networkId] = networkInfo.Name
		cfg.CoinGeckoIds[networkId] = make(map[string]string)
		cfg.CoinMarketCapIds[networkId] = make(map[string]string)
		cfg.MinAmounts[networkId] = make(map[string]*big.Int)

		cfg.EVMs[networkId] = BridgeEvm{
			RouterContractAddress: networkInfo.RouterContractAddress,
			ContractOperationsCost: ContractOperationsCost{
				MintGasUnits:   networkInfo.ContractOperationsCost.MintGasUnits,
				UnlockGasUnits: networkInfo.ContractOperationsCost.UnlockGasUnits,
			},
		}
	}

	for tokenName, tokenInfo := range bridge.RegularTokens {
		nativeChainId := tokenInfo.NativeChain
		nativeAddr := tokenName
		if tokenInfo.Address != nil && *tokenInfo.Address != "" {
			nativeAddr = *tokenInfo.Address
		}

		if cfg.CoinGeckoIds[nativeChainId] == nil {
			cfg.CoinGeckoIds[nativeChainId] = make(map[string]string)
		}
		if cfg.CoinMarketCapIds[nativeChainId] == nil {
			cfg.CoinMarketCapIds[nativeChainId] = make(map[string]string)
		}
		if cfg.MinAmounts[nativeChainId] == nil {
			cfg.MinAmounts[nativeChainId] = make(map[string]*big.Int)
		}

		if tokenInfo.CoinGeckoId != "" {
			cfg.CoinGeckoIds[nativeChainId][nativeAddr] = tokenInfo.CoinGeckoId
		}
		if tokenInfo.CoinMarketCapId != "" {
			cfg.CoinMarketCapIds[nativeChainId][nativeAddr] = tokenInfo.CoinMarketCapId
		}
		cfg.MinAmounts[nativeChainId][nativeAddr] = big.NewInt(0)

		if nativeChainId == constants.HederaNetworkId && cfg.Hedera != nil {
			cfg.Hedera.FeePercentages[nativeAddr] = tokenInfo.FeePercentage
		}

		for wrappedChainId, wrappedAddr := range tokenInfo.AddressesPerNetwork {
			if cfg.MinAmounts[wrappedChainId] == nil {
				cfg.MinAmounts[wrappedChainId] = make(map[string]*big.Int)
			}
			cfg.MinAmounts[wrappedChainId][wrappedAddr] = big.NewInt(0)
		}
	}

	return cfg
}

func (b Bridge) LoadStaticMinAmountsForWrappedFungibleTokens(parsedBridge parser.Bridge, assetsService service.Assets) {
	for tokenName, tokenInfo := range parsedBridge.RegularTokens {
		if tokenInfo.MinAmount == nil {
			continue
		}
		nativeChainId := tokenInfo.NativeChain
		nativeAsset := tokenName
		if tokenInfo.Address != nil && *tokenInfo.Address != "" {
			nativeAsset = *tokenInfo.Address
		}

		nativeFungibleAssetsInfo, _ := assetsService.FungibleAssetInfo(nativeChainId, nativeAsset)
		for wrappedNetworkId, wrappedAddress := range tokenInfo.AddressesPerNetwork {
			b.MinAmounts[wrappedNetworkId][wrappedAddress] = big.NewInt(0)
			wrappedFungibleAssetsInfo, _ := assetsService.FungibleAssetInfo(wrappedNetworkId, wrappedAddress)
			if nativeFungibleAssetsInfo != nil && wrappedFungibleAssetsInfo != nil {
				targetAmount := decimalHelper.TargetAmount(nativeFungibleAssetsInfo.Decimals, wrappedFungibleAssetsInfo.Decimals, tokenInfo.MinAmount)
				b.MinAmounts[wrappedNetworkId][wrappedAddress] = targetAmount
			}
		}
	}
}
