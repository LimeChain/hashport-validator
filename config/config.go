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

	// TODO: Replace this configuration with an external configuration service
	LoadWrappedToNativeAssets(&configuration.AssetMappings)
	LoadNativeHederaFees(&configuration.AssetMappings, &configuration.Validator.Clients.Hedera.FeePercentages)

	if err := env.Parse(&configuration); err != nil {
		panic(err)
	}

	return configuration
}

func LoadNativeHederaFees(mappings *AssetMappings, hederaFeePercentages *map[string]int64) {
	hederaMappings := mappings.NativeToWrappedByNetwork[0]
	feePercentages := map[string]int64{}
	for nativeTokenID, nativeToken := range hederaMappings.Native {
		feePercentages[nativeTokenID] = nativeToken.FeePercentage
	}
	*hederaFeePercentages = feePercentages
}

func LoadWrappedToNativeAssets(a *AssetMappings) {
	a.WrappedToNativeByNetwork = map[int64]map[string]*NativeAsset{}
	for nativeChainId, network := range a.NativeToWrappedByNetwork {
		for nativeAsset, native := range network.Native {
			for wrappedChainId, wrappedAsset := range native.Assets {
				if a.WrappedToNativeByNetwork[wrappedChainId] == nil {
					a.WrappedToNativeByNetwork[wrappedChainId] = make(map[string]*NativeAsset)
				}
				a.WrappedToNativeByNetwork[wrappedChainId][wrappedAsset] = &NativeAsset{
					ChainId: nativeChainId,
					Asset:   nativeAsset,
				}
			}
		}
	}
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
	Validator     Validator     `yaml:"validator"`
	AssetMappings AssetMappings `yaml:"asset-mappings"`
}

type AssetMappings struct {
	NativeToWrappedByNetwork map[int64]Network `yaml:"networks,omitempty"`
	WrappedToNativeByNetwork map[int64]map[string]*NativeAsset
}

type NativeAsset struct {
	ChainId int64
	Asset   string
}

func (a *AssetMappings) NativeToWrapped(nativeAsset string, nativeChainId, targetChainId int64) string {
	return a.NativeToWrappedByNetwork[nativeChainId].Native[nativeAsset].Assets[targetChainId]
}

func (a *AssetMappings) WrappedToNative(wrappedAsset string, wrappedChainId int64) *NativeAsset {
	return a.WrappedToNativeByNetwork[wrappedChainId][wrappedAsset]
}

type Network struct {
	EVMClient EVM                 `yaml:"evm_client"`
	Native    map[string]Mappings `yaml:"tokens"`
}

type Mappings struct {
	FeePercentage int64            `yaml:"fee_percentage"`
	Assets        map[int64]string `yaml:"mappings"`
}

type Validator struct {
	LogLevel    string   `yaml:"log_level" env:"VALIDATOR_LOG_LEVEL"`
	RestApiOnly bool     `yaml:"rest_api_only" env:"VALIDATOR_REST_API_ONLY"`
	Port        string   `yaml:"port" env:"VALIDATOR_PORT"`
	Database    Database `yaml:"database"`
	Clients     Clients  `yaml:"clients"`
	Recovery    Recovery `yaml:"recovery"`
}

type Clients struct {
	EVM        map[int64]EVM `yaml:"evm"`
	MirrorNode MirrorNode    `yaml:"mirror_node"`
	Hedera     Hedera        `yaml:"hedera"`
}

type Recovery struct {
	StartTimestamp int64 `yaml:"start_timestamp" env:"VALIDATOR_RECOVERY_START_TIMESTAMP"`
}

type EVM struct {
	NodeUrl               string `yaml:"node_url" env:"VALIDATOR_CLIENTS_ETHEREUM_NODE_URL"`
	RouterContractAddress string `yaml:"router_contract_address" env:"VALIDATOR_CLIENTS_ETHEREUM_ROUTER_CONTRACT_ADDRESS"`
	BlockConfirmations    uint64 `yaml:"block_confirmations" env:"VALIDATOR_CLIENTS_ETHEREUM_BLOCK_CONFIRMATIONS"`
	PrivateKey            string `yaml:"private_key" env:"VALIDATOR_CLIENTS_ETHEREUM_PRIVATE_KEY"`
}

type Hedera struct {
	NetworkType    string   `yaml:"network_type" env:"VALIDATOR_CLIENTS_HEDERA_NETWORK_TYPE"`
	Operator       Operator `yaml:"operator"`
	BridgeAccount  string   `yaml:"bridge_account" env:"VALIDATOR_CLIENTS_HEDERA_BRIDGE_ACCOUNT"`
	PayerAccount   string   `yaml:"payer_account" env:"VALIDATOR_CLIENTS_HEDERA_PAYER_ACCOUNT"`
	TopicId        string   `yaml:"topic_id" env:"VALIDATOR_CLIENTS_HEDERA_TOPIC_ID"`
	FeePercentages map[string]int64
	Members        []string `yaml:"members" env:"VALIDATOR_CLIENTS_HEDERA_MEMBERS"`
}

type Operator struct {
	AccountId  string `yaml:"account_id" env:"VALIDATOR_CLIENTS_HEDERA_OPERATOR_ACCOUNT_ID"`
	PrivateKey string `yaml:"private_key" env:"VALIDATOR_CLIENTS_HEDERA_OPERATOR_PRIVATE_KEY"`
}

type MirrorNode struct {
	ClientAddress   string        `yaml:"client_address" env:"VALIDATOR_CLIENTS_MIRROR_NODE_CLIENT_ADDRESS"`
	ApiAddress      string        `yaml:"api_address" env:"VALIDATOR_CLIENTS_MIRROR_NODE_API_ADDRESS"`
	PollingInterval time.Duration `yaml:"polling_interval" env:"VALIDATOR_CLIENTS_MIRROR_NODE_POLLING_INTERVAL"`
}

type Database struct {
	Host     string `yaml:"host" env:"VALIDATOR_DATABASE_HOST"`
	Name     string `yaml:"name" env:"VALIDATOR_DATABASE_NAME"`
	Password string `yaml:"password" env:"VALIDATOR_DATABASE_DB_PASSWORD"`
	Port     string `yaml:"port" env:"VALIDATOR_DATABASE_DB_PORT"`
	Username string `yaml:"username" env:"VALIDATOR_DATABASE_DB_USERNAME"`
}
