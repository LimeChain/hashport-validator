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

package parser

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"math/big"
	"time"
)

type ExtendedConfig struct {
	Bridge ExtendedBridge `yaml:"bridge"`
}

type ExtendedBridge struct {
	UseLocalConfig    bool                         `yaml:"use_local_config,omitempty" json:"useLocalConfig,omitempty"`
	ConfigTopicId     string                       `yaml:"config_topic_id,omitempty" json:"configTopicId,omitempty"`
	PollingInterval   time.Duration                `yaml:"polling_interval,omitempty" json:"pollingInterval,omitempty"`
	TopicId           string                       `yaml:"topic_id,omitempty" json:"topicId,omitempty"`
	MonitoredAccounts map[string]string            `yaml:"monitored_accounts,omitempty" json:"monitoredAccounts,omitempty"`
	Networks          map[uint64]*NetworkForDeploy `yaml:"networks,omitempty" json:"networks,omitempty"`
}

func (b *ExtendedBridge) Validate(hederaNetworkId uint64) {
	for network, networkInfo := range b.Networks {

		if network == hederaNetworkId {
			// Native
			for tokenAddress, tokenInfo := range networkInfo.Tokens.Fungible {
				tokenInfo.Validate(tokenAddress)
			}
			for tokenAddress, tokenInfo := range networkInfo.Tokens.Nft {
				tokenInfo.Validate(tokenAddress)
			}
		} else {
			// Wrapped
			for tokenAddress, tokenInfo := range networkInfo.Tokens.Fungible {
				if _, ok := tokenInfo.Networks[hederaNetworkId]; !ok {
					continue
				}
				tokenInfo.Validate(tokenAddress)
			}
		}
	}
}

func (b *ExtendedBridge) ToBridgeParser() *parser.Bridge {
	parsedBridge := new(parser.Bridge)
	parsedBridge.UseLocalConfig = false
	parsedBridge.ConfigTopicId = b.ConfigTopicId
	parsedBridge.PollingInterval = b.PollingInterval
	parsedBridge.TopicId = b.TopicId
	parsedBridge.MonitoredAccounts = b.MonitoredAccounts
	parsedBridge.Networks = make(map[uint64]*parser.Network)

	for network, networkInfo := range b.Networks {
		parsedBridge.Networks[network] = &parser.Network{
			Name:                  networkInfo.Name,
			BridgeAccount:         networkInfo.BridgeAccount,
			PayerAccount:          networkInfo.PayerAccount,
			RouterContractAddress: networkInfo.RouterContractAddress,
			Members:               networkInfo.Members,
			Tokens:                parser.Tokens{},
		}

		parsedBridge.Networks[network].Tokens.Fungible = make(map[string]parser.Token)
		for tokenAddress, tokenInfo := range networkInfo.Tokens.Fungible {
			parsedBridge.Networks[network].Tokens.Fungible[tokenAddress] = parser.Token{
				Fee:               tokenInfo.Fee,
				FeeAmountInUsd:    tokenInfo.FeeAmountInUsd,
				FeePercentage:     tokenInfo.FeePercentage,
				MinFeeAmountInUsd: tokenInfo.MinFeeAmountInUsd,
				MinAmount:         tokenInfo.MinAmount,
				Networks:          tokenInfo.Networks,
				CoinGeckoId:       tokenInfo.CoinGeckoId,
				CoinMarketCapId:   tokenInfo.CoinMarketCapId,
				ReleaseTimestamp:  tokenInfo.ReleaseTimestamp,
			}
		}

		parsedBridge.Networks[network].Tokens.Nft = make(map[string]parser.Token)
		for tokenAddress, tokenInfo := range networkInfo.Tokens.Nft {
			parsedBridge.Networks[network].Tokens.Nft[tokenAddress] = parser.Token{
				Fee:               tokenInfo.Fee,
				FeeAmountInUsd:    tokenInfo.FeeAmountInUsd,
				FeePercentage:     tokenInfo.FeePercentage,
				MinFeeAmountInUsd: tokenInfo.MinFeeAmountInUsd,
				MinAmount:         tokenInfo.MinAmount,
				Networks:          tokenInfo.Networks,
				CoinGeckoId:       tokenInfo.CoinGeckoId,
				CoinMarketCapId:   tokenInfo.CoinMarketCapId,
				ReleaseTimestamp:  tokenInfo.ReleaseTimestamp,
			}
		}
	}
	return parsedBridge
}

type NetworkForDeploy struct {
	Name                  string          `yaml:"name,omitempty" json:"name,omitempty"`
	BridgeAccount         string          `yaml:"bridge_account,omitempty" json:"bridgeAccount,omitempty"`
	PayerAccount          string          `yaml:"payer_account,omitempty" json:"payerAccount,omitempty"`
	RouterContractAddress string          `yaml:"router_contract_address,omitempty" json:"routerContractAddress,omitempty"`
	Members               []string        `yaml:"members,omitempty" json:"members,omitempty"`
	Tokens                TokensForDeploy `yaml:"tokens,omitempty" json:"tokens,omitempty"`
}

type TokensForDeploy struct {
	Fungible map[string]*FungibleTokenForDeploy    `yaml:"fungible,omitempty" json:"fungible,omitempty"`
	Nft      map[string]*NonFungibleTokenForDeploy `yaml:"nft,omitempty" json:"nft,omitempty"`
}

type NonFungibleTokenForDeploy struct {
	Fee               int64             `yaml:"fee,omitempty" json:"fee,omitempty"`
	FeeAmountInUsd    string            `yaml:"fee_amount_in_usd,omitempty" json:"feeAmountInUsd,omitempty"`
	FeePercentage     int64             `yaml:"fee_percentage,omitempty" json:"feePercentage,omitempty"`
	MinFeeAmountInUsd string            `yaml:"min_fee_amount_in_usd,omitempty" json:"minFeeAmountInUsd,omitempty"`
	MinAmount         *big.Int          `yaml:"min_amount,omitempty" json:"minAmount,omitempty"`
	Networks          map[uint64]string `yaml:"networks,omitempty" json:"networks,omitempty"`
	CoinGeckoId       string            `yaml:"coin_gecko_id,omitempty" json:"coinGeckoId,omitempty"`
	CoinMarketCapId   string            `yaml:"coin_market_cap_id,omitempty" json:"coinMarketCapId,omitempty"`
	ReleaseTimestamp  uint64            `yaml:"release_timestamp,omitempty" json:"releaseTimestamp,omitempty"`
	Name              string            `yaml:"name,omitempty"`
	Symbol            string            `yaml:"symbol,omitempty"`
}

func (tokenInfo *NonFungibleTokenForDeploy) Validate(tokenAddress string) {
	if tokenInfo.Name == "" {
		panic(fmt.Sprintf("invalid [Name] from bridge config for non-fungible token '%s'", tokenAddress))
	}
	if tokenInfo.Symbol == "" {
		panic(fmt.Sprintf("invalid [Symbol] from bridge config for non-fungible token '%s'", tokenAddress))
	}
}

type FungibleTokenForDeploy struct {
	Fee               int64             `yaml:"fee,omitempty" json:"fee,omitempty"`
	FeeAmountInUsd    string            `yaml:"fee_amount_in_usd,omitempty" json:"feeAmountInUsd,omitempty"`
	FeePercentage     int64             `yaml:"fee_percentage,omitempty" json:"feePercentage,omitempty"`
	MinFeeAmountInUsd string            `yaml:"min_fee_amount_in_usd,omitempty" json:"minFeeAmountInUsd,omitempty"`
	MinAmount         *big.Int          `yaml:"min_amount,omitempty" json:"minAmount,omitempty"`
	Networks          map[uint64]string `yaml:"networks,omitempty" json:"networks,omitempty"`
	CoinGeckoId       string            `yaml:"coin_gecko_id,omitempty" json:"coinGeckoId,omitempty"`
	CoinMarketCapId   string            `yaml:"coin_market_cap_id,omitempty" json:"coinMarketCapId,omitempty"`
	ReleaseTimestamp  uint64            `yaml:"release_timestamp,omitempty" json:"releaseTimestamp,omitempty"`
	Name              string            `yaml:"name,omitempty"`
	Symbol            string            `yaml:"symbol,omitempty"`
	Decimals          uint              `yaml:"decimals,omitempty"`
	Supply            uint64            `yaml:"supply,omitempty"`
}

func (tokenInfo *FungibleTokenForDeploy) Validate(tokenAddress string) {
	if tokenAddress == constants.Hbar {
		return
	}

	if tokenInfo.Name == "" {
		panic(fmt.Sprintf("invalid [Name] from bridge config for fungible token '%s'", tokenAddress))
	}
	if tokenInfo.Symbol == "" {
		panic(fmt.Sprintf("invalid [Symbol] from bridge config for fungible token '%s'", tokenAddress))
	}
	if tokenInfo.Decimals == 0 {
		panic(fmt.Sprintf("invalid [Decimals] from bridge config for fungible token '%s'", tokenAddress))
	}
	if tokenInfo.Supply == 0 {
		panic(fmt.Sprintf("invalid [Supply] from bridge config for fungible token '%s'", tokenAddress))
	}
}
