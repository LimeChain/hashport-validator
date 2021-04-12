/*
 * Copyright 2021 LimeChain Ltd.
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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = "config/application.yml"
	mainConfigFile    = "application.yml"
)

func LoadConfig() Config {
	var configuration Config
	GetConfig(&configuration, defaultConfigFile)
	GetConfig(&configuration, mainConfigFile)

	if err := env.Parse(&configuration); err != nil {
		panic(err)
	}

	return configuration
}

func GetConfig(config *Config, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	filename, _ := filepath.Abs(path)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

type Config struct {
	Hedera Hedera `yaml:"hedera"`
}

type Hedera struct {
	LogLevel    string     `yaml:"log_level"`
	Validator   Validator  `yaml:"validator"`
	Eth         Ethereum   `yaml:"eth"`
	MirrorNode  MirrorNode `yaml:"mirror_node"`
	Client      Client     `yaml:"client"`
	Watcher     Watcher    `yaml:"watcher"`
	Handler     Handler    `yaml:"handler"`
	Recovery    Recovery   `yaml:"recovery"`
	RestApiOnly bool       `yaml:"rest_api_only"`
}

type Handler struct {
	CryptoTransfer   CryptoTransferHandler   `yaml:"crypto-transfer"`
	ConsensusMessage ConsensusMessageHandler `yaml:"consensus-message"`
}

type ConsensusMessageHandler struct {
	TopicId      string `yaml:"topic_id"`
	SendDeadline int64  `yaml:"send_deadline"`
}

type CryptoTransferHandler struct {
	TopicId         string        `yaml:"topic_id"`
	PollingInterval time.Duration `yaml:"polling_interval"`
}

type Watcher struct {
	CryptoTransfer   CryptoTransfer   `yaml:"crypto-transfer"`
	ConsensusMessage ConsensusMessage `yaml:"consensus-message"`
}

type Ethereum struct {
	NodeUrl               string `yaml:"node_url" env:"HEDERA_ETH_BRIDGE_ETH_NODE_URL"`
	RouterContractAddress string `yaml:"router_contract_address" env:"HEDERA_ETH_BRIDGE_ETH_ROUTER_CONTRACT_ADDRESS"`
}

type CryptoTransfer struct {
	Account ID `yaml:"account" env:"HEDERA_ETH_BRIDGE_WATCHER_CRYPTO_TRANSFER"`
}

type ConsensusMessage struct {
	Topic ID `yaml:"topic" env:"HEDERA_ETH_BRIDGE_WATCHER_CONSENSUS_MESSAGE"`
}

type ID struct {
	Id         string `yaml:"id"`
	MaxRetries int    `yaml:"max_retries"`
}

type Client struct {
	NetworkType      string   `yaml:"network_type" env:"HEDERA_ETH_BRIDGE_CLIENT_NETWORK_TYPE"`
	Operator         Operator `yaml:"operator"`
	BaseGasUsage     uint64   `yaml:"base_gas_usage"`
	GasPerValidator  uint64   `yaml:"gas_per_validator"`
	ThresholdAccount string   `yaml:"threshold_account" env:"HEDERA_ETH_BRIDGE_CLIENT_THRESHOLD_ACCOUNT"`
	PayerAccount     string   `yaml:"payer_account" env:"HEDERA_ETH_BRIDGE_CLIENT_PAYER_ACCOUNT"`
}

type Operator struct {
	AccountId     string `yaml:"account_id" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_ACCOUNT_ID"`
	EthPrivateKey string `yaml:"eth_private_key" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_ETH_PRIVATE_KEY"`
	PrivateKey    string `yaml:"private_key" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_PRIVATE_KEY"`
}

type MirrorNode struct {
	ClientAddress   string        `yaml:"client_address" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_CLIENT_ADDRESS"`
	ApiAddress      string        `yaml:"api_address" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_API_ADDRESS"`
	PollingInterval time.Duration `yaml:"polling_interval" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_POLLING_INTERVAL"`
}

type Validator struct {
	Db   Db     `yaml:"db"`
	Port string `yaml:"port" env:"HEDERA_ETH_BRIDGE_VALIDATOR_PORT"`
}

type Db struct {
	Host     string `yaml:"host" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_HOST"`
	Name     string `yaml:"name" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_NAME"`
	Password string `yaml:"password" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_PASSWORD"`
	Port     string `yaml:"port" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_PORT"`
	Username string `yaml:"username" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_USERNAME"`
}

type Recovery struct {
	Timestamp int64 `yaml:"timestamp" env:"HEDERA_ETH_BRIDGE_VALIDATOR_RECOVERY_TIMESTAMP"`
}
