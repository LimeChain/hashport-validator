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

package setup

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/assets"
	fee "github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	evm_signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/bootstrap"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
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
	if err := config.GetConfig(&e2eConfig, e2eConfigPath); err != nil {
		panic(err)
	}
	if err := config.GetConfig(&e2eConfig, e2eBridgeConfigPath); err != nil {
		panic(err)
	}

	config.NewBridge(e2eConfig.Bridge)

	configuration := Config{
		Hedera: Hedera{
			NetworkType:       e2eConfig.Hedera.NetworkType,
			BridgeAccount:     e2eConfig.Hedera.BridgeAccount,
			Members:           e2eConfig.Hedera.Members,
			TopicID:           e2eConfig.Hedera.TopicID,
			Sender:            Sender(e2eConfig.Hedera.Sender),
			DbValidationProps: make([]config.Database, len(e2eConfig.Hedera.DbValidationProps)),
			MirrorNode:        *new(config.MirrorNode).DefaultOrConfig(&e2eConfig.Hedera.MirrorNode),
		},
		EVM:             make(map[uint64]config.Evm),
		Tokens:          e2eConfig.Tokens,
		ValidatorUrl:    e2eConfig.ValidatorUrl,
		Bridge:          e2eConfig.Bridge,
		FeePercentages:  map[string]int64{},
		NftConstantFees: map[string]int64{},
		NftDynamicFees:  map[string]decimal.Decimal{},
	}

	if e2eConfig.Bridge.Networks[constants.HederaNetworkId] != nil {
		feeInfo := config.LoadHederaFees(e2eConfig.Bridge.Networks[constants.HederaNetworkId].Tokens)
		configuration.FeePercentages = feeInfo.FungiblePercentages
		configuration.NftConstantFees = feeInfo.ConstantNftFees
		configuration.NftDynamicFees = feeInfo.DynamicNftFees
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

	routerClients, evmFungibleTokenClients, evmNftClients := routerAndEVMTokenClientsFromEVMUtils(setup.Clients.EVM)
	setup.AssetMappings = assets.NewService(e2eConfig.Bridge.Networks, e2eConfig.Hedera.BridgeAccount, configuration.FeePercentages, routerClients, setup.Clients.MirrorNode, evmFungibleTokenClients, evmNftClients)

	return setup
}

// Setup used by the e2e tests. Preloaded with all necessary dependencies
type Setup struct {
	BridgeAccount   hederaSDK.AccountID
	TopicID         hederaSDK.TopicID
	TokenID         hederaSDK.TokenID
	NativeEvmToken  string
	NftTokenID      hederaSDK.TokenID
	NftSerialNumber int64
	NftConstantFees map[string]int64
	NftDynamicFees  map[string]decimal.Decimal
	FeePercentages  map[string]int64
	Members         []hederaSDK.AccountID
	Clients         *clients
	DbValidator     *db_validation.Service
	AssetMappings   service.Assets
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

	nftTokenID, err := hederaSDK.TokenIDFromString(config.Tokens.NftToken)
	if err != nil {
		return nil, err
	}

	for token, fp := range config.FeePercentages {
		if fp < constants.FeeMinPercentage || fp > constants.FeeMaxPercentage {
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
		BridgeAccount:   bridgeAccount,
		TopicID:         topicID,
		TokenID:         tokenID,
		NftTokenID:      nftTokenID,
		NftSerialNumber: config.Tokens.NftSerialNumber,
		NativeEvmToken:  config.Tokens.EvmNativeToken,
		NftConstantFees: config.NftConstantFees,
		NftDynamicFees:  config.NftDynamicFees,
		FeePercentages:  config.FeePercentages,
		Members:         members,
		Clients:         clients,
		DbValidator:     dbValidator,
		AssetMappings:   config.AssetMappings,
	}, nil
}

// clients used by the e2e tests
type clients struct {
	Hedera          *hederaSDK.Client
	EVM             map[uint64]EVMUtils
	MirrorNode      *mirror_node.Client
	ValidatorClient *e2eClients.Validator
	FeeCalculator   service.Fee
	Distributor     service.Distributor
}

func routerAndEVMTokenClientsFromEVMUtils(evmUtils map[uint64]EVMUtils) (
	routerClients map[uint64]client.DiamondRouter,
	evmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken,
	evmNftClients map[uint64]map[string]client.EvmNft,
) {
	routerClients = make(map[uint64]client.DiamondRouter)
	evmFungibleTokenClients = make(map[uint64]map[string]client.EvmFungibleToken)
	evmNftClients = make(map[uint64]map[string]client.EvmNft)
	for networkId, evmUtil := range evmUtils {
		routerClients[networkId] = evmUtil.RouterContract

		evmFungibleTokenClients[networkId] = make(map[string]client.EvmFungibleToken)
		for tokenAddress, evmTokenClient := range evmUtil.EVMFungibleTokenClients {
			evmFungibleTokenClients[networkId][tokenAddress] = evmTokenClient
		}

		evmNftClients[networkId] = make(map[string]client.EvmNft)
		for tokenAddress, evmTokenClient := range evmUtil.EVMNftClients {
			evmNftClients[networkId][tokenAddress] = evmTokenClient
		}
	}

	return routerClients, evmFungibleTokenClients, evmNftClients
}

// newClients instantiates the clients for the e2e tests
func newClients(config Config) (*clients, error) {
	hederaClient, err := initHederaClient(config.Hedera.Sender, config.Hedera.NetworkType)
	if err != nil {
		return nil, err
	}

	EVM := make(map[uint64]EVMUtils)
	evmClients := make(map[uint64]client.EVM)
	for configChainId, conf := range config.EVM {
		evmClient := evm.NewClient(conf, configChainId)
		evmClients[configChainId] = evmClient
		clientChainId, e := evmClient.ChainID(context.Background())
		if e != nil {
			return nil, errors.New("failed to retrieve chain ID on new client")
		}
		if configChainId == clientChainId.Uint64() {
			evmClient.SetChainID(clientChainId.Uint64())
		} else {
			return nil, errors.New("chain IDs mismatch config and actual")
		}
		routerContractAddress := common.HexToAddress(config.Bridge.Networks[configChainId].RouterContractAddress)
		routerInstance, err := router.NewRouter(routerContractAddress, evmClient)

		signer := evm_signer.NewEVMSigner(evmClient.GetPrivateKey())
		keyTransactor, err := signer.NewKeyTransactor(clientChainId)
		if err != nil {
			return nil, err
		}

		EVM[configChainId] = EVMUtils{
			EVMClient:               evmClient,
			RouterContract:          routerInstance,
			KeyTransactor:           keyTransactor,
			Signer:                  signer,
			Receiver:                common.HexToAddress(signer.Address()),
			RouterAddress:           routerContractAddress,
			WTokenContractAddress:   config.Tokens.WToken,
			EVMFungibleTokenClients: make(map[string]client.EvmFungibleToken),
			EVMNftClients:           make(map[string]client.EvmNft),
		}
	}

	evmFungibleTokenClients := bootstrap.InitEvmFungibleTokenClients(config.Bridge.Networks, evmClients)
	for networkId := range config.EVM {
		for tokenAddress, tokenClient := range evmFungibleTokenClients[networkId] {
			EVM[networkId].EVMFungibleTokenClients[tokenAddress] = tokenClient
		}
	}

	evmNftClients := bootstrap.InitEvmNftClients(config.Bridge.Networks, evmClients)
	for networkId := range config.EVM {
		for tokenAddress, tokenClient := range evmNftClients[networkId] {
			EVM[networkId].EVMNftClients[tokenAddress] = tokenClient
		}
	}
	validatorClient := e2eClients.NewValidatorClient(config.ValidatorUrl)

	mirrorNode := mirror_node.NewClient(config.Hedera.MirrorNode)

	return &clients{
		Hedera:          hederaClient,
		EVM:             EVM,
		ValidatorClient: validatorClient,
		MirrorNode:      mirrorNode,
		FeeCalculator:   fee.New(config.FeePercentages),
		Distributor:     distributor.New(config.Hedera.Members),
	}, nil
}

func InitWrappedAssetContract(nativeAsset string, assetsService service.Assets, sourceChain, targetChain uint64, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	wTokenContractAddress, err := NativeToWrappedAsset(assetsService, sourceChain, targetChain, nativeAsset)
	if err != nil {
		return nil, err
	}

	return InitAssetContract(wTokenContractAddress, evmClient)
}

func InitAssetContract(asset string, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	return wtoken.NewWtoken(common.HexToAddress(asset), evmClient.GetClient())
}

func NativeToWrappedAsset(assetsService service.Assets, sourceChain, targetChain uint64, nativeAsset string) (string, error) {
	wrappedAsset := assetsService.NativeToWrapped(nativeAsset, sourceChain, targetChain)

	if wrappedAsset == "" {
		return "", errors.New(fmt.Sprintf("EvmFungibleToken [%s] is not supported", nativeAsset))
	}

	return wrappedAsset, nil
}

func WrappedToNativeAsset(assetsService service.Assets, sourceChainId uint64, asset string) (*asset.NativeAsset, error) {
	targetAsset := assetsService.WrappedToNative(asset, sourceChainId)
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
	Hedera          Hedera
	EVM             map[uint64]config.Evm
	Tokens          e2eParser.Tokens
	ValidatorUrl    string
	Bridge          parser.Bridge
	AssetMappings   service.Assets
	FeePercentages  map[string]int64
	NftConstantFees map[string]int64
	NftDynamicFees  map[string]decimal.Decimal
}

type EVMUtils struct {
	EVMClient               *evm.Client
	EVMFungibleTokenClients map[string]client.EvmFungibleToken
	EVMNftClients           map[string]client.EvmNft
	RouterContract          *router.Router
	KeyTransactor           *bind.TransactOpts
	Signer                  *evm_signer.Signer
	Receiver                common.Address
	RouterAddress           common.Address
	WTokenContractAddress   string
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
}

// Sender props from the application.yml
type Sender struct {
	Account    string
	PrivateKey string
}

type Receiver struct {
	Account string
}
