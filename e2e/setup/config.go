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

package setup

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	e2eClients "github.com/limechain/hedera-eth-bridge-validator/e2e/clients"
	db_validation "github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	"gopkg.in/yaml.v2"
)

const (
	// The configuration file for the e2e tests. Placed at ./e2e/setup/application.yml
	e2eConfigPath = "setup/application.yml"
)

// Load loads the e2e application.yml from the ./e2e/setup folder and parses it to suitable working struct for the e2e tests
func Load() *Setup {
	var configuration Config
	err := getConfig(&configuration, e2eConfigPath)
	if err := env.Parse(&configuration); err != nil {
		panic(err)
	}
	setup, err := newSetup(configuration)
	if err != nil {
		panic(err)
	}
	return setup
}

// getConfig parses the application.yml file from the provided path and unmarshalls it to the provided config struct
func getConfig(config *Config, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	filename, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, config)
	return err
}

// Setup used by the e2e tests. Preloaded with all necessary dependencies
type Setup struct {
	BridgeAccount hederaSDK.AccountID
	SenderAccount hederaSDK.AccountID
	TopicID       hederaSDK.TopicID
	TokenID       hederaSDK.TokenID
	Clients       *clients
	DbValidation  *db_validation.Service
}

// newSetup instantiates new Setup struct
func newSetup(config Config) (*Setup, error) {
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

	tokenID, err := hederaSDK.TokenIDFromString(config.Tokens.WToken)
	if err != nil {
		return nil, err
	}

	clients, err := newClients(config)
	if err != nil {
		return nil, err
	}

	return &Setup{
		BridgeAccount: bridgeAccount,
		SenderAccount: senderAccount,
		TopicID:       topicID,
		TokenID:       tokenID,
		Clients:       clients,
		DbValidation:  db_validation.NewService(config.Hedera.DbValidationProps),
	}, nil
}

// clients used by the e2e tests
type clients struct {
	Hedera          *hederaSDK.Client
	EthClient       *ethereum.Client
	WHbarContract   *wtoken.Wtoken
	WTokenContract  *wtoken.Wtoken
	RouterContract  *router.Router
	ValidatorClient *e2eClients.Validator
	KeyTransactor   *bind.TransactOpts
}

// newClients instantiates the clients for the e2e tests
func newClients(config Config) (*clients, error) {
	hederaClient, err := initHederaClient(config.Hedera.Sender, config.Hedera.NetworkType)
	if err != nil {
		return nil, err
	}
	ethClient := ethereum.NewClient(config.Ethereum)

	routerContractAddress := common.HexToAddress(config.Ethereum.RouterContractAddress)
	routerInstance, err := router.NewRouter(routerContractAddress, ethClient.Client)

	wHbarInstance, err := initTokenContract(config.Tokens.WHbar, routerInstance, ethClient)
	if err != nil {
		return nil, err
	}

	wTokenInstance, err := initTokenContract(config.Tokens.WToken, routerInstance, ethClient)
	if err != nil {
		return nil, err
	}

	validatorClient := e2eClients.NewValidatorClient(config.ValidatorUrl)

	signer := eth.NewEthSigner(config.Signer)
	keyTransactor, err := signer.NewKeyTransactor(ethClient.ChainID())
	if err != nil {
		return nil, err
	}

	return &clients{
		Hedera:          hederaClient,
		EthClient:       ethClient,
		WHbarContract:   wHbarInstance,
		WTokenContract:  wTokenInstance,
		RouterContract:  routerInstance,
		ValidatorClient: validatorClient,
		KeyTransactor:   keyTransactor,
	}, nil
}

func initTokenContract(nativeToken string, routerInstance *router.Router, ethClient *ethereum.Client) (*wtoken.Wtoken, error) {
	wTokenContractAddress, err := ParseToken(routerInstance, nativeToken)
	if err != nil {
		return nil, err
	}

	wTokenInstance, err := wtoken.NewWtoken(*wTokenContractAddress, ethClient.Client)
	if err != nil {
		return nil, err
	}

	return wTokenInstance, nil
}

func ParseToken(routerInstance *router.Router, nativeToken string) (*common.Address, error) {
	nilErc20Address := "0x0000000000000000000000000000000000000000"
	wrappedToken, err := routerInstance.NativeToWrappedToken(
		nil,
		common.RightPadBytes([]byte(nativeToken), 32),
	)
	if err != nil {
		return nil, err
	}

	wTokenContractHex := wrappedToken.String()
	if wTokenContractHex == nilErc20Address {
		return nil, errors.New(fmt.Sprintf("Token [%s] is not supported", nativeToken))
	}

	address := common.HexToAddress(wTokenContractHex)
	return &address, nil
}

func initHederaClient(sender Sender, networkType string) (*hederaSDK.Client, error) {
	var client *hederaSDK.Client
	switch networkType {
	case "mainnet":
		client = hederaSDK.ClientForMainnet()
	case "testnet":
		client = hederaSDK.ClientForTestnet()
	case "previewnet":
		client = hederaSDK.ClientForPreviewnet()
	default:
		panic(fmt.Sprintf("Invalid Client NetworkType provided: [%s]", networkType))
	}
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

// e2eConfig used to load and parse from application.yml
type Config struct {
	Hedera       Hedera          `yaml:"hedera"`
	Ethereum     config.Ethereum `yaml:"ethereum"`
	Tokens       Tokens          `yaml:"tokens"`
	ValidatorUrl string          `yaml:"validator_url"`
	Signer       string          `yaml:"eth_signer"`
}

type Tokens struct {
	WHbar  string `yaml:"whbar"`
	WToken string `yaml:"wtoken"`
}

// hedera props from the application.yml
type Hedera struct {
	NetworkType       string      `yaml:"network_type"`
	BridgeAccount     string      `yaml:"bridge_account"`
	TopicID           string      `yaml:"topic_id"`
	Sender            Sender      `yaml:"sender"`
	DbValidationProps []config.Db `yaml:"dbs"`
}

// sender props from the application.yml
type Sender struct {
	Account    string `yaml:"account"`
	PrivateKey string `yaml:"private_key"`
}
