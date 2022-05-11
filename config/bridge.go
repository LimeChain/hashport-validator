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

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Bridge struct {
	TopicId           string
	Hedera            *BridgeHedera
	EVMs              map[uint64]BridgeEvm
	CoinMarketCapIds  map[uint64]map[string]string
	CoinGeckoIds      map[uint64]map[string]string
	MinAmounts        map[uint64]map[string]*big.Int
	MonitoredAccounts map[string]string
}

type BridgeHedera struct {
	BridgeAccount  string
	PayerAccount   string
	Members        []string
	Tokens         map[string]HederaToken
	FeePercentages map[string]int64
	NftFees        map[string]int64
	NftDynamicFees map[string]int64
}

type HederaToken struct {
	Fee               int64
	FeePercentage     int64
	MinFeeAmountInUsd string
	MinAmount         *big.Int
	Networks          map[uint64]string
}

func NewHederaTokenFromToken(token parser.Token) HederaToken {
	return HederaToken{
		Fee:               token.Fee,
		FeePercentage:     token.FeePercentage,
		MinFeeAmountInUsd: token.MinFeeAmountInUsd,
		MinAmount:         token.MinAmount,
		Networks:          token.Networks,
	}
}

type Token struct {
	MinFeeAmountInUsd *big.Int
	Networks          map[uint64]string
}

type BridgeEvm struct {
	RouterContractAddress string
	Tokens                map[string]Token
}

func NewBridge(bridge parser.Bridge) Bridge {
	config := Bridge{
		TopicId:           bridge.TopicId,
		Hedera:            nil,
		EVMs:              make(map[uint64]BridgeEvm),
		MonitoredAccounts: bridge.MonitoredAccounts,
	}

	config.CoinGeckoIds = make(map[uint64]map[string]string)
	config.CoinMarketCapIds = make(map[uint64]map[string]string)
	config.MinAmounts = make(map[uint64]map[string]*big.Int)
	for networkId, networkInfo := range bridge.Networks {
		if networkInfo.Name == constants.HederaName {
			constants.HederaNetworkId = networkId
		}
		constants.NetworksByName[networkInfo.Name] = networkId
		constants.NetworksById[networkId] = networkInfo.Name
		config.CoinGeckoIds[networkId] = make(map[string]string)
		config.CoinMarketCapIds[networkId] = make(map[string]string)
		config.MinAmounts[networkId] = make(map[string]*big.Int)

		if networkId == constants.HederaNetworkId { // Hedera
			config.Hedera = &BridgeHedera{
				BridgeAccount: networkInfo.BridgeAccount,
				PayerAccount:  networkInfo.PayerAccount,
				Members:       networkInfo.Members,
				Tokens:        make(map[string]HederaToken),
			}

			for name, tokenInfo := range networkInfo.Tokens.Nft {
				config.Hedera.Tokens[name] = NewHederaTokenFromToken(tokenInfo)
			}
			hederaFeePercentages, constantHederaNftFees, dynamicHederaNftFees := LoadHederaFees(networkInfo.Tokens)
			config.Hedera.FeePercentages = hederaFeePercentages
			config.Hedera.NftFees = constantHederaNftFees
			config.Hedera.NftDynamicFees = dynamicHederaNftFees
		} else {
			config.EVMs[networkId] = BridgeEvm{
				RouterContractAddress: networkInfo.RouterContractAddress,
				Tokens:                make(map[string]Token),
			}
			// Currently, only EVM Fungible native tokens are supported
			for name, tokenInfo := range networkInfo.Tokens.Fungible {
				config.EVMs[networkId].Tokens[name] = Token{Networks: tokenInfo.Networks}
			}
		}

		for tokenAddress, tokenInfo := range networkInfo.Tokens.Fungible {
			if tokenInfo.CoinGeckoId != "" {
				config.CoinGeckoIds[networkId][tokenAddress] = tokenInfo.CoinGeckoId
			}

			if tokenInfo.CoinMarketCapId != "" {
				config.CoinMarketCapIds[networkId][tokenAddress] = tokenInfo.CoinMarketCapId
			}

			config.MinAmounts[networkId][tokenAddress] = big.NewInt(0)
			if tokenInfo.MinAmount != nil {
				config.MinAmounts[networkId][tokenAddress] = tokenInfo.MinAmount
			}
			for wrappedNetworkId, wrappedAddress := range tokenInfo.Networks {
				if config.MinAmounts[wrappedNetworkId] == nil {
					config.MinAmounts[wrappedNetworkId] = make(map[string]*big.Int)
				}
				config.MinAmounts[wrappedNetworkId][wrappedAddress] = big.NewInt(0)
			}

			if networkId == constants.HederaNetworkId {
				config.Hedera.Tokens[tokenAddress] = NewHederaTokenFromToken(tokenInfo)
			}
		}
	}

	return config
}

func (b Bridge) LoadStaticMinAmountsForWrappedFungibleTokens(parsedBridge parser.Bridge, assetsService service.Assets) {
	for networkId, networkInfo := range parsedBridge.Networks {
		for nativeAddress, tokenInfo := range networkInfo.Tokens.Fungible {
			nativeFungibleAssetsInfo, _ := assetsService.FungibleAssetInfo(networkId, nativeAddress)
			for wrappedNetworkId, wrappedAddress := range tokenInfo.Networks {
				b.MinAmounts[wrappedNetworkId][wrappedAddress] = big.NewInt(0)
				if tokenInfo.MinAmount != nil {
					wrappedFungibleAssetsInfo, _ := assetsService.FungibleAssetInfo(wrappedNetworkId, wrappedAddress)
					targetAmount := decimal.TargetAmount(nativeFungibleAssetsInfo.Decimals, wrappedFungibleAssetsInfo.Decimals, tokenInfo.MinAmount)
					b.MinAmounts[wrappedNetworkId][wrappedAddress] = targetAmount
				}
			}
		}
	}
}

func LoadHederaFees(tokens parser.Tokens) (fungiblePercentages map[string]int64, constantNftFees map[string]int64, dynamicNftFees map[string]int64) {
	for token, value := range tokens.Fungible {
		fungiblePercentages[token] = value.FeePercentage
	}
	for token, value := range tokens.Nft {
		if value.Fee == 0 {
			if value.FeeAmountInUsd != 0 {
				dynamicNftFees[token] = value.FeeAmountInUsd
				continue
			} else {
				log.Fatalf("NFT [%s] has zero fee", token)
			}
		}
		constantNftFees[token] = value.Fee
	}

	return fungiblePercentages, constantNftFees, dynamicNftFees
}
