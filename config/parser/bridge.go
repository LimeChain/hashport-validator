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
	"math/big"
	"time"
)

/*
Structs used to parse the bridge YAML configuration
*/
type Bridge struct {
	UseLocalConfig    bool                `yaml:"use_local_config,omitempty" json:"useLocalConfig,omitempty"`
	ConfigTopicId     string              `yaml:"config_topic_id,omitempty" json:"configTopicId,omitempty"`
	FeePolicyTopicId  string              `yaml:"fee_policy_topic_id,omitempty" json:"feePolicyTopicId,omitempty"`
	PollingInterval   time.Duration       `yaml:"polling_interval,omitempty" json:"pollingInterval,omitempty"`
	TopicId           string              `yaml:"topic_id,omitempty" json:"topicId,omitempty"`
	Networks          map[uint64]*Network `yaml:"networks,omitempty" json:"networks,omitempty"`
	MonitoredAccounts map[string]string   `yaml:"monitored_accounts,omitempty" json:"monitoredAccounts,omitempty"`
}

func (b *Bridge) Update(from *Bridge) {
	b.UseLocalConfig = from.UseLocalConfig
	b.ConfigTopicId = from.ConfigTopicId
	b.FeePolicyTopicId = from.FeePolicyTopicId
	b.PollingInterval = from.PollingInterval
	b.TopicId = from.TopicId
	b.Networks = from.Networks
	b.MonitoredAccounts = from.MonitoredAccounts
}

type Network struct {
	Name                  string   `yaml:"name,omitempty" json:"name,omitempty"`
	BridgeAccount         string   `yaml:"bridge_account,omitempty" json:"bridgeAccount,omitempty"`
	PayerAccount          string   `yaml:"payer_account,omitempty" json:"payerAccount,omitempty"`
	RouterContractAddress string   `yaml:"router_contract_address,omitempty" json:"routerContractAddress,omitempty"`
	Members               []string `yaml:"members,omitempty" json:"members,omitempty"`
	Tokens                Tokens   `yaml:"tokens,omitempty" json:"tokens,omitempty"`
}

type Tokens struct {
	Fungible map[string]Token `yaml:"fungible,omitempty" json:"fungible,omitempty"`
	Nft      map[string]Token `yaml:"nft,omitempty" json:"nft,omitempty"`
}

type Token struct {
	Fee               int64             `yaml:"fee,omitempty" json:"fee,omitempty"`                                 // Represent a constant fee for Non-Fungible tokens. Applies only for Hedera Native Tokens
	FeeAmountInUsd    string            `yaml:"fee_amount_in_usd,omitempty" json:"feeAmountInUsd,omitempty"`        // Represent a dynamic fee amount in $USD for Non-Fungible tokens. Applies only for Hedera Native Tokens
	FeePercentage     int64             `yaml:"fee_percentage,omitempty" json:"feePercentage,omitempty"`            // Represents a constant fee for Fungible Tokens. Applies only for Hedera Native Tokens
	MinFeeAmountInUsd string            `yaml:"min_fee_amount_in_usd,omitempty" json:"minFeeAmountInUsd,omitempty"` // Represents a constant minimum fee amount in USD which is needed for the validator not to be on a loss
	MinAmount         *big.Int          `yaml:"min_amount,omitempty" json:"minAmount,omitempty"`                    // Represents a constant for minimum amount which is used when there is no 'coin_gecko_id' or 'coin_market_cap_id' supplied in the config.
	Networks          map[uint64]string `yaml:"networks,omitempty" json:"networks,omitempty"`
	CoinGeckoId       string            `yaml:"coin_gecko_id,omitempty" json:"coinGeckoId,omitempty"`
	CoinMarketCapId   string            `yaml:"coin_market_cap_id,omitempty" json:"coinMarketCapId,omitempty"`
	ReleaseTimestamp  uint64            `yaml:"release_timestamp,omitempty" json:"releaseTimestamp,omitempty"`
}
