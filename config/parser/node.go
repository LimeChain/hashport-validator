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

import "time"

/*
	Structs used to parse the node YAML configuration
*/
type Node struct {
	Database            Database   `yaml:"database"`
	Clients             Clients    `yaml:"clients"`
	LogLevel            string     `yaml:"log_level"`
	Port                string     `yaml:"port"`
	Validator           bool       `yaml:"validator"`
	Monitoring          Monitoring `yaml:"monitoring"`
	BridgeConfigTopicId Monitoring `yaml:"bridge_config_topic_id"`
}

type Database struct {
	Host     string `yaml:"host" env:"VALIDATOR_DATABASE_HOST"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
}

type Clients struct {
	Evm           map[uint64]Evm `yaml:"evm"`
	Hedera        Hedera         `yaml:"hedera"`
	MirrorNode    MirrorNode     `yaml:"mirror_node"`
	CoinGecko     CoinGecko      `yaml:"coingecko"`
	CoinMarketCap CoinMarketCap  `yaml:"coin_market_cap"`
}

// Evm //

type Evm struct {
	BlockConfirmations uint64        `yaml:"block_confirmations"`
	NodeUrl            string        `yaml:"node_url"`
	PrivateKey         string        `yaml:"private_key"`
	StartBlock         int64         `yaml:"start_block"`
	PollingInterval    time.Duration `yaml:"polling_interval"`
	MaxLogsBlocks      int64         `yaml:"max_logs_blocks"`
}

// Hedera //

type Hedera struct {
	Operator       Operator          `yaml:"operator"`
	Network        string            `yaml:"network"`
	Rpc            map[string]string `yaml:"rpc"`
	StartTimestamp int64             `yaml:"start_timestamp"`
	MaxRetry       int               `yaml:"max_retry" default:"5"`
}

type Operator struct {
	AccountId  string `yaml:"account_id"`
	PrivateKey string `yaml:"private_key"`
}

// MirrorNode //

type MirrorNode struct {
	ClientAddress     string        `yaml:"client_address"`
	ApiAddress        string        `yaml:"api_address"`
	PollingInterval   time.Duration `yaml:"polling_interval"`
	QueryMaxLimit     int64         `yaml:"query_max_limit"`
	QueryDefaultLimit int64         `yaml:"query_default_limit"`
}

// CoinGecko //

type CoinGecko struct {
	ApiAddress string `yaml:"api_address" json:"apiAddress,omitempty"`
}

// CoinMarketCap //

type CoinMarketCap struct {
	ApiKey     string `yaml:"api_key" json:"apiKey,omitempty"`
	ApiAddress string `yaml:"api_address" json:"apiAddress,omitempty"`
}

type Monitoring struct {
	Enable           bool          `yaml:"enable"`
	DashboardPolling time.Duration `yaml:"dashboard_polling"`
}
