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
	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/whbar"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

const (
	// The configuration file for the e2e tests. Placed at ./e2e/config/application.yml
	defaultConfigFile = "config/application.yml"
)

// LoadE2EConfig loads the e2e application.yml from the ./e2e/config folder and parses it to suitable working struct for the e2e tests
func LoadE2EConfig() *Setup {
	configuration, err := config.GetConfig(testConfig{}, defaultConfigFile)
	if err := env.Parse(&configuration); err != nil {
		panic(err)
	}
	setup, err := newSetup(configuration.(testConfig))
	if err != nil {
		panic(err)
	}
	return setup
}

// Setup used by the e2e tests. Preloaded with all necessary dependencies
type Setup struct {
	BridgeAccount hederaSDK.AccountID
	SenderAccount hederaSDK.AccountID
	TopicID       hederaSDK.TopicID
	Clients       *clients
}

// newSetup instantiates new Setup struct
func newSetup(config testConfig) (*Setup, error) {
	bridgeAccount, err := hederaSDK.AccountIDFromString(config.Hedera.BridgeAccount)
	if err != nil {
		return nil, err
	}
	senderAccount, err := hederaSDK.AccountIDFromString(config.Hedera.Sender.Account)
	if err != nil {
		return nil, err
	}
	topicID, err := hederaSDK.TopicIDFromString(config.Hedera.TopicID)
	if err != nil {
		return nil, err
	}

	clients, err := newClients(config)
	if err != nil {
		return nil, err
	}
	return &Setup{BridgeAccount: bridgeAccount, SenderAccount: senderAccount, TopicID: topicID, Clients: clients}, nil
}

// clients used by teh e2e tests
type clients struct {
	Hedera        *hederaSDK.Client
	EthClient     *ethereum.EthereumClient
	WHbarContract *whbar.Whbar
}

// newClients instantiates the clients for the e2e tests
func newClients(config testConfig) (*clients, error) {
	hederaClient, err := initHederaClient(config.Hedera.Sender)
	if err != nil {
		return nil, err
	}
	ethClient := ethereum.NewEthereumClient(config.Ethereum)
	whbarContractAddress := common.HexToAddress(config.Ethereum.WhbarContractAddress)
	whbarInstance, err := whbar.NewWhbar(whbarContractAddress, ethClient.Client)
	if err != nil {
		return nil, err
	}
	return &clients{Hedera: hederaClient, EthClient: ethClient, WHbarContract: whbarInstance}, nil
}

func initHederaClient(sender sender) (*hederaSDK.Client, error) {
	client := hederaSDK.ClientForTestnet()
	senderAccount, err := hederaSDK.AccountIDFromString(sender.Account)
	if err != nil {
		return nil, err
	}
	privateKey, err := hederaSDK.PrivateKeyFromString(sender.PrivateKey)
	if err != nil {
		return nil, err
	}
	client.SetOperator(senderAccount, privateKey)

	return client, nil
}

// testConfig used to load and parse from application.yml
type testConfig struct {
	Hedera   hedera          `yaml:"hedera"`
	Ethereum config.Ethereum `yaml:"ethereum"`
}

// hedera props from the application.yml
type hedera struct {
	BridgeAccount string `yaml:"bridge_account"`
	TopicID       string `yaml:"topic_id"`
	Sender        sender `yaml:"sender"`
}

// sender props from the application.yml
type sender struct {
	Account    string `yaml:"account"`
	PrivateKey string `yaml:"private_key"`
}
