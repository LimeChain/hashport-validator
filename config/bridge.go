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
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"math/big"
)

type Bridge struct {
	TopicId          string
	Hedera           *BridgeHedera
	EVMs             map[uint64]BridgeEvm
	CoinMarketCapIds map[uint64]map[string]string
	CoinGeckoIds     map[uint64]map[string]string
}

type BridgeHedera struct {
	BridgeAccount  string
	PayerAccount   string
	Members        []string
	Tokens         map[string]HederaToken
	FeePercentages map[string]int64
	NftFees        map[string]int64
}

type HederaToken struct {
	Fee               int64
	FeePercentage     int64
	MinFeeAmountInUsd string
	Networks          map[uint64]string
}

func NewHederaTokenFromToken(token parser.Token) HederaToken {
	return HederaToken{
		Fee:               token.Fee,
		FeePercentage:     token.FeePercentage,
		MinFeeAmountInUsd: token.MinFeeAmountInUsd,
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
		TopicId: bridge.TopicId,
		Hedera:  nil,
		EVMs:    make(map[uint64]BridgeEvm),
	}

	config.CoinGeckoIds = make(map[uint64]map[string]string)
	config.CoinMarketCapIds = make(map[uint64]map[string]string)
	for networkId, networkInfo := range bridge.Networks {
		constants.NetworksByName[networkInfo.Name] = networkId
		constants.NetworksById[networkId] = networkInfo.Name
		config.CoinGeckoIds[networkId] = make(map[string]string)
		config.CoinMarketCapIds[networkId] = make(map[string]string)

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
			hederaFeePercentages, hederaNftFees := LoadHederaFees(networkInfo.Tokens)
			config.Hedera.FeePercentages = hederaFeePercentages
			config.Hedera.NftFees = hederaNftFees
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

		for name, tokenInfo := range networkInfo.Tokens.Fungible {
			config.CoinGeckoIds[networkId][name] = tokenInfo.CoinGeckoId
			config.CoinMarketCapIds[networkId][name] = tokenInfo.CoinMarketCapId
			if networkId == constants.HederaNetworkId {
				config.Hedera.Tokens[name] = NewHederaTokenFromToken(tokenInfo)
			}
		}
	}

	return config
}

func LoadHederaFees(tokens parser.Tokens) (fungiblePercentages map[string]int64, nftFees map[string]int64) {
	feePercentages := map[string]int64{}
	fees := map[string]int64{}
	for token, value := range tokens.Fungible {
		feePercentages[token] = value.FeePercentage
	}
	for token, value := range tokens.Nft {
		if value.Fee == 0 {
			log.Fatalf("NFT [%s] has zero fee", token)
		}
		fees[token] = value.Fee
	}

	return feePercentages, fees
}
