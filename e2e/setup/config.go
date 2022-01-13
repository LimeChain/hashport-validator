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
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	fee "github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	evm_signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	e2eClients "github.com/limechain/hedera-eth-bridge-validator/e2e/clients"
	db_validation "github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	e2eParser "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/parser"
)

const (
	// The configuration file for the e2e tests. Placed at ./e2e/setup/application.yml
	e2eConfigPath       = "setup/application.yml"
	e2eBridgeConfigPath = "setup/bridge.yml"
)

// Load loads the e2e application.yml from the ./e2e/setup folder and parses it to suitable working struct for the e2e tests
func Load() *Setup {
	var e2eConfig e2eParser.Config
	config.GetConfig(&e2eConfig, e2eConfigPath)
	config.GetConfig(&e2eConfig, e2eBridgeConfigPath)

	configuration := Config{
		Hedera: Hedera{
			NetworkType:       e2eConfig.Hedera.NetworkType,
			BridgeAccount:     e2eConfig.Hedera.BridgeAccount,
			Members:           e2eConfig.Hedera.Members,
			TopicID:           e2eConfig.Hedera.TopicID,
			Sender:            Sender(e2eConfig.Hedera.Sender),
			DbValidationProps: make([]config.Database, len(e2eConfig.Hedera.DbValidationProps)),
			MirrorNode:        config.MirrorNode(e2eConfig.Hedera.MirrorNode),
			Monitoring:        config.Monitoring(e2eConfig.Hedera.Monitoring),
		},
		EVM:            make(map[int64]config.Evm),
		Tokens:         e2eConfig.Tokens,
		ValidatorUrl:   e2eConfig.ValidatorUrl,
		Bridge:         e2eConfig.Bridge,
		AssetMappings:  config.LoadAssets(e2eConfig.Bridge.Networks),
		FeePercentages: map[string]int64{},
	}

	if e2eConfig.Bridge.Networks[0] != nil {
		configuration.FeePercentages = config.LoadHederaFeePercentages(e2eConfig.Bridge.Networks[0].Tokens)
	}

	for i, props := range e2eConfig.Hedera.DbValidationProps {
		configuration.Hedera.DbValidationProps[i] = config.Database(props)
	}

	for key, value := range e2eConfig.EVM {
		configuration.EVM[key] = config.Evm(value)
	}
	setup, err := newSetup(configuration)
	if err != nil {
		panic(err)
	}
	return setup
}

// Setup used by the e2e tests. Preloaded with all necessary dependencies
type Setup struct {
	BridgeAccount  hederaSDK.AccountID
	TopicID        hederaSDK.TopicID
	TokenID        hederaSDK.TokenID
	NativeEvmToken string
	FeePercentages map[string]int64
	Members        []hederaSDK.AccountID
	Clients        *clients
	DbValidator    *db_validation.Service
	AssetMappings  config.Assets
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

	for token, fp := range config.FeePercentages {
		if fp < fee.MinPercentage || fp > fee.MaxPercentage {
			return nil, errors.New(fmt.Sprintf("[%s] - invalid fee percentage [%d]", token, fp))
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
		FeePercentages: config.FeePercentages,
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
		routerContractAddress := common.HexToAddress(config.Bridge.Networks[chainId].RouterContractAddress)
		routerInstance, err := router.NewRouter(routerContractAddress, evmClient)

		chain, err := evmClient.ChainID(context.Background())
		signer := evm_signer.NewEVMSigner(evmClient.GetPrivateKey())
		keyTransactor, err := signer.NewKeyTransactor(chain)
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
		FeeCalculator:   fee.New(config.FeePercentages),
		Distributor:     distributor.New(config.Hedera.Members),
	}, nil
}

func InitWrappedAssetContract(nativeAsset string, nativeAssets config.Assets, sourceChain, targetChain int64, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	wTokenContractAddress, err := NativeToWrappedAsset(nativeAssets, sourceChain, targetChain, nativeAsset)
	if err != nil {
		return nil, err
	}

	return InitAssetContract(wTokenContractAddress, evmClient)
}

func InitAssetContract(asset string, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	return wtoken.NewWtoken(common.HexToAddress(asset), evmClient.GetClient())
}

func NativeToWrappedAsset(assetMappings config.Assets, sourceChain, targetChain int64, nativeAsset string) (string, error) {
	wrappedAsset := assetMappings.NativeToWrapped(nativeAsset, sourceChain, targetChain)

	if wrappedAsset == "" {
		return "", errors.New(fmt.Sprintf("Token [%s] is not supported", nativeAsset))
	}

	return wrappedAsset, nil
}

func WrappedToNativeAsset(assetMappings config.Assets, sourceChainId int64, asset string) (*config.NativeAsset, error) {
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

// Config used to load and parse from application.yml
type Config struct {
	Hedera         Hedera
	EVM            map[int64]config.Evm
	Tokens         e2eParser.Tokens
	ValidatorUrl   string
	Bridge         parser.Bridge
	AssetMappings  config.Assets
	FeePercentages map[string]int64
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

// Hedera props from the application.yml
type Hedera struct {
	NetworkType       string
	BridgeAccount     string
	Members           []string
	TopicID           string
	Sender            Sender
	DbValidationProps []config.Database
	MirrorNode        config.MirrorNode
	Monitoring        config.Monitoring
}

// Sender props from the application.yml
type Sender struct {
	Account    string
	PrivateKey string
}

type Receiver struct {
	Account string
}
