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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	fee "github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	e2eClients "github.com/limechain/hedera-eth-bridge-validator/e2e/clients"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	evm_signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/config"
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
	EthReceiver   common.Address
	RouterAddress common.Address
	TopicID       hederaSDK.TopicID
	TokenID       hederaSDK.TokenID
	FeePercentage int64
	Members       []hederaSDK.AccountID
	Clients       *clients
	DbValidator   *db_validation.Service
	AssetMappings config.AssetMappings
}

// newSetup instantiates new Setup struct
func newSetup(config Config) (*Setup, error) {
	bridgeAccount, err := hederaSDK.AccountIDFromString(config.Hedera.BridgeAccount)
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

	if config.Hedera.FeePercentage < fee.MinPercentage || config.Hedera.FeePercentage > fee.MaxPercentage {
		return nil, errors.New(fmt.Sprintf("invalid fee percentage [%d]", config.Hedera.FeePercentage))
	}

	if len(config.Hedera.Members) == 0 {
		return nil, errors.New(fmt.Sprintf("members account ids cannot be 0"))
	}

	var members []hederaSDK.AccountID
	for _, v := range config.Hedera.Members {
		account, err := hederaSDK.AccountIDFromString(v)
		if err != nil {
			return nil, err
		}
		members = append(members, account)
	}

	clients, err := newClients(config)
	if err != nil {
		return nil, err
	}

	dbValidator := db_validation.NewService(config.Hedera.DbValidationProps)

	return &Setup{
		BridgeAccount: bridgeAccount,
		EthReceiver:   common.HexToAddress(clients.Signer.Address()),
		TopicID:       topicID,
		TokenID:       tokenID,
		FeePercentage: config.Hedera.FeePercentage,
		Members:       members,
		Clients:       clients,
		RouterAddress: common.HexToAddress(config.EVM.RouterContractAddress),
		DbValidator:   dbValidator,
		AssetMappings: config.AssetMappings,
	}, nil
}

// clients used by the e2e tests
type clients struct {
	Hedera          *hederaSDK.Client
	EthClient       *evm.Client
	WHbarContract   *wtoken.Wtoken
	WTokenContract  *wtoken.Wtoken
	RouterContract  *router.Router
	KeyTransactor   *bind.TransactOpts
	MirrorNode      *mirror_node.Client
	ValidatorClient *e2eClients.Validator
	FeeCalculator   service.Fee
	Distributor     service.Distributor
	Signer          service.Signer
}

// newClients instantiates the clients for the e2e tests
func newClients(config Config) (*clients, error) {
	hederaClient, err := initHederaClient(config.Hedera.Sender, config.Hedera.NetworkType)
	if err != nil {
		return nil, err
	}
	ethClient := evm.NewClient(config.EVM)

	routerContractAddress := common.HexToAddress(config.EVM.RouterContractAddress)
	routerInstance, err := router.NewRouter(routerContractAddress, ethClient.Client)

	wHbarInstance, err := initAssetContract(config.Tokens.WHbar, config.AssetMappings, ethClient)
	if err != nil {
		return nil, err
	}

	wTokenInstance, err := initAssetContract(config.Tokens.WToken, config.AssetMappings, ethClient)
	if err != nil {
		return nil, err
	}

	signer := evm_signer.NewEVMSigner(config.EVM.PrivateKey)
	keyTransactor, err := signer.NewKeyTransactor(ethClient.ChainID())
	if err != nil {
		return nil, err
	}

	validatorClient := e2eClients.NewValidatorClient(config.ValidatorUrl)

	mirrorNode := mirror_node.NewClient(config.Hedera.MirrorNode.ApiAddress, config.Hedera.MirrorNode.PollingInterval)

	return &clients{
		Hedera:          hederaClient,
		EthClient:       ethClient,
		WHbarContract:   wHbarInstance,
		WTokenContract:  wTokenInstance,
		RouterContract:  routerInstance,
		ValidatorClient: validatorClient,
		KeyTransactor:   keyTransactor,
		MirrorNode:      mirrorNode,
		FeeCalculator:   fee.New(config.Hedera.FeePercentage),
		Distributor:     distributor.New(config.Hedera.Members),
		Signer:          signer,
	}, nil
}

func initAssetContract(nativeAsset string, nativeAssets config.AssetMappings, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	wTokenContractAddress, err := WrappedAsset(nativeAssets, nativeAsset)
	if err != nil {
		return nil, err
	}

	wTokenInstance, err := wtoken.NewWtoken(*wTokenContractAddress, evmClient.Client)
	if err != nil {
		return nil, err
	}

	return wTokenInstance, nil
}

func WrappedAsset(nativeAssets config.AssetMappings, nativeAsset string) (*common.Address, error) {
	wTokenContractHex := nativeAssets.NativeToWrappedByNetwork[0].NativeAssets[nativeAsset][1]

	if wTokenContractHex == "" {
		return nil, errors.New(fmt.Sprintf("Token [%s] is not supported", nativeAsset))
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
	Hedera        Hedera               `yaml:"hedera"`
	EVM           config.EVM           `yaml:"ethereum"`
	Tokens        Tokens               `yaml:"tokens"`
	ValidatorUrl  string               `yaml:"validator_url"`
	AssetMappings config.AssetMappings `yaml:"asset-mappings"`
}

type Tokens struct {
	WHbar  string `yaml:"whbar"`
	WToken string `yaml:"wtoken"`
}

// hedera props from the application.yml
type Hedera struct {
	NetworkType       string            `yaml:"network_type"`
	BridgeAccount     string            `yaml:"bridge_account"`
	FeePercentage     int64             `yaml:"fee_percentage"`
	Members           []string          `yaml:"members"`
	TopicID           string            `yaml:"topic_id"`
	Sender            Sender            `yaml:"sender"`
	DbValidationProps []config.Database `yaml:"dbs"`
	MirrorNode        config.MirrorNode `yaml:"mirror_node"`
}

// sender props from the application.yml
type Sender struct {
	Account    string `yaml:"account"`
	PrivateKey string `yaml:"private_key"`
}

type Receiver struct {
	Account string `yaml:"account"`
}
