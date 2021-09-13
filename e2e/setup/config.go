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
	config.LoadWrappedToNativeAssets(&configuration.AssetMappings)
	config.LoadNativeHederaFees(&configuration.AssetMappings, &configuration.Hedera.FeePercentages)
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
	BridgeAccount  hederaSDK.AccountID
	TopicID        hederaSDK.TopicID
	TokenID        hederaSDK.TokenID
	NativeEvmToken EvmToken
	FeePercentages map[string]int64
	Members        []hederaSDK.AccountID
	Clients        *clients
	DbValidator    *db_validation.Service
	AssetMappings  config.AssetMappings
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

	for _, f := range config.Hedera.FeePercentages {
		if f < fee.MinPercentage || f > fee.MaxPercentage {
			return nil, errors.New(fmt.Sprintf("invalid fee percentage [%d]", config.Hedera.FeePercentages))
		}
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
		BridgeAccount:  bridgeAccount,
		TopicID:        topicID,
		TokenID:        tokenID,
		NativeEvmToken: config.Tokens.EvmNativeToken,
		FeePercentages: config.Hedera.FeePercentages,
		Members:        members,
		Clients:        clients,
		DbValidator:    dbValidator,
		AssetMappings:  config.AssetMappings,
	}, nil
}

// clients used by the e2e tests
type clients struct {
	Hedera          *hederaSDK.Client
	EVM             map[int64]EVMUtils
	MirrorNode      *mirror_node.Client
	ValidatorClient *e2eClients.Validator
	FeeCalculator   service.Fee
	Distributor     service.Distributor
}

// newClients instantiates the clients for the e2e tests
func newClients(config Config) (*clients, error) {
	hederaClient, err := initHederaClient(config.Hedera.Sender, config.Hedera.NetworkType)
	if err != nil {
		return nil, err
	}

	EVM := make(map[int64]EVMUtils)
	for chainId, conf := range config.EVM {
		evmClient := evm.NewClient(conf)
		routerContractAddress := common.HexToAddress(conf.RouterContractAddress)
		routerInstance, err := router.NewRouter(routerContractAddress, evmClient)

		signer := evm_signer.NewEVMSigner(evmClient.GetPrivateKey())
		keyTransactor, err := signer.NewKeyTransactor(evmClient.ChainID())
		if err != nil {
			return nil, err
		}

		EVM[chainId] = EVMUtils{
			EVMClient:             evmClient,
			RouterContract:        routerInstance,
			KeyTransactor:         keyTransactor,
			Signer:                signer,
			Receiver:              common.HexToAddress(signer.Address()),
			RouterAddress:         routerContractAddress,
			WTokenContractAddress: config.Tokens.WToken,
		}
	}

	validatorClient := e2eClients.NewValidatorClient(config.ValidatorUrl)

	mirrorNode := mirror_node.NewClient(config.Hedera.MirrorNode.ApiAddress, config.Hedera.MirrorNode.PollingInterval)

	return &clients{
		Hedera:          hederaClient,
		EVM:             EVM,
		ValidatorClient: validatorClient,
		MirrorNode:      mirrorNode,
		FeeCalculator:   fee.New(config.Hedera.FeePercentages),
		Distributor:     distributor.New(config.Hedera.Members),
	}, nil
}

func InitWrappedAssetContract(nativeAsset string, nativeAssets config.AssetMappings, sourceChain, targetChain int64, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	wTokenContractAddress, err := NativeToWrappedAsset(nativeAssets, sourceChain, targetChain, nativeAsset)
	if err != nil {
		return nil, err
	}

	return InitAssetContract(wTokenContractAddress, evmClient)
}

func InitAssetContract(asset string, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	return wtoken.NewWtoken(common.HexToAddress(asset), evmClient.Client)
}

func NativeToWrappedAsset(assetMappings config.AssetMappings, sourceChain, targetChain int64, nativeAsset string) (string, error) {
	wrappedAsset := assetMappings.NativeToWrapped(nativeAsset, sourceChain, targetChain)

	if wrappedAsset == "" {
		return "", errors.New(fmt.Sprintf("Token [%s] is not supported", nativeAsset))
	}

	return wrappedAsset, nil
}

func WrappedToNativeAsset(assetMappings config.AssetMappings, sourceChainId int64, asset string) (*config.NativeAsset, error) {
	targetAsset := assetMappings.WrappedToNative(asset, sourceChainId)
	if targetAsset == nil {
		return nil, errors.New(fmt.Sprintf("Wrapped token [%s] on [%d] is not supported", asset, sourceChainId))
	}

	return targetAsset, nil
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
	EVM           map[int64]config.EVM `yaml:"evm"`
	Tokens        Tokens               `yaml:"tokens"`
	ValidatorUrl  string               `yaml:"validator_url"`
	AssetMappings config.AssetMappings `yaml:"asset-mappings"`
}

type EVMUtils struct {
	EVMClient             *evm.Client
	RouterContract        *router.Router
	KeyTransactor         *bind.TransactOpts
	Signer                *evm_signer.Signer
	Receiver              common.Address
	RouterAddress         common.Address
	WTokenContractAddress string
}

type Tokens struct {
	WHbar          string   `yaml:"whbar"`
	WToken         string   `yaml:"wtoken"`
	EvmNativeToken EvmToken `yaml:"evm_native_token"`
}

type EvmToken struct {
	Address  string `yaml:"address"`
	Decimals int64  `yaml:"decimals"`
}

// hedera props from the application.yml
type Hedera struct {
	NetworkType       string `yaml:"network_type"`
	BridgeAccount     string `yaml:"bridge_account"`
	FeePercentages    map[string]int64
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
