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

	"github.com/shopspring/decimal"
)

/*
Structs used to parse the bridge YAML configuration
*/
type Bridge struct {
	UseLocalConfig      bool                     `yaml:"use_local_config,omitempty" json:"useLocalConfig,omitempty"`
	ConfigTopicId       string                   `yaml:"config_topic_id,omitempty" json:"configTopicId,omitempty"`
	PollingInterval     time.Duration            `yaml:"polling_interval,omitempty" json:"pollingInterval,omitempty"`
	TopicId             string                   `yaml:"topic_id,omitempty" json:"topicId,omitempty"`
	Networks            Networks                 `yaml:"networks,omitempty" json:"networks,omitempty"`
	RegularTokens       map[string]*RegularToken `yaml:"regular_tokens,omitempty" json:"regularTokens,omitempty"`
	MonitoredAccounts   map[string]string        `yaml:"monitored_accounts,omitempty" json:"monitoredAccounts,omitempty"`
	BlacklistedAccounts []string                 `yaml:"blacklist,omitempty" json:"blacklistedAccounts,omitempty"`
}

func (b *Bridge) Update(from *Bridge) {
	b.UseLocalConfig = from.UseLocalConfig
	b.ConfigTopicId = from.ConfigTopicId
	b.PollingInterval = from.PollingInterval
	b.TopicId = from.TopicId
	b.Networks = from.Networks
	b.RegularTokens = from.RegularTokens
	b.MonitoredAccounts = from.MonitoredAccounts
	b.BlacklistedAccounts = from.BlacklistedAccounts
}

type Networks struct {
	Hedera map[uint64]*HederaNetwork `yaml:"hedera,omitempty" json:"hedera,omitempty"`
	EVM    map[uint64]*EVMNetwork    `yaml:"evm,omitempty" json:"evm,omitempty"`
}

type HederaNetwork struct {
	Name          string   `yaml:"name,omitempty" json:"name,omitempty"`
	BridgeAccount string   `yaml:"bridge_account,omitempty" json:"bridgeAccount,omitempty"`
	PayerAccount  string   `yaml:"payer_account,omitempty" json:"payerAccount,omitempty"`
	Members       []string `yaml:"members,omitempty" json:"members,omitempty"`
	MintCost      float64  `yaml:"mint_cost,omitempty" json:"mintCost,omitempty"`
	UnlockCost    float64  `yaml:"unlock_cost,omitempty" json:"unlockCost,omitempty"`
}

type EVMNetwork struct {
	Name                   string                 `yaml:"name,omitempty" json:"name,omitempty"`
	RouterContractAddress  string                 `yaml:"router_contract_address,omitempty" json:"routerContractAddress,omitempty"`
	ContractOperationsCost ContractOperationsCost `yaml:"contract_operations_cost,omitempty" json:"contractOperationsCost,omitempty"`
}

type ContractOperationsCost struct {
	MintGasUnits   uint64 `yaml:"mint_gas_units,omitempty" json:"mintGasUnits,omitempty"`
	UnlockGasUnits uint64 `yaml:"unlock_gas_units,omitempty" json:"unlockGasUnits,omitempty"`
}

type RegularToken struct {
	NativeChain         uint64            `yaml:"native_chain,omitempty" json:"nativeChain,omitempty"`
	Address             *string           `yaml:"address" json:"address,omitempty"`
	MinFeeAmountInUsd   *decimal.Decimal  `yaml:"min_fee_amount_in_usd,omitempty" json:"minFeeAmountInUsd,omitempty"`
	CoinGeckoId         string            `yaml:"coin_gecko_id,omitempty" json:"coinGeckoId,omitempty"`
	CoinMarketCapId     string            `yaml:"coin_market_cap_id,omitempty" json:"coinMarketCapId,omitempty"`
	FeePercentage       int64             `yaml:"fee_percentage,omitempty" json:"feePercentage,omitempty"`
	MinAmount           *big.Int          `yaml:"min_amount,omitempty" json:"minAmount,omitempty"`
	ReleaseTimestamp    uint64            `yaml:"release_timestamp,omitempty" json:"releaseTimestamp,omitempty"`
	AddressesPerNetwork map[uint64]string `yaml:"addresses_per_network,omitempty" json:"addressesPerNetwork,omitempty"`
}
