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
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/verify"
	evmSetup "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/evm"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/assets"
	fee "github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	evm_signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/bootstrap"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	e2eClients "github.com/limechain/hedera-eth-bridge-validator/e2e/clients"
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
			NetworkType:               e2eConfig.Hedera.NetworkType,
			BridgeAccount:             e2eConfig.Hedera.BridgeAccount,
			PayerAccount:              e2eConfig.Hedera.PayerAccount,
			Members:                   e2eConfig.Hedera.Members,
			TopicID:                   e2eConfig.Hedera.TopicID,
			Treasury:                  e2eConfig.Hedera.Treasury,
			ValidatorRewardPercentage: e2eConfig.Hedera.ValidatorRewardPercentage,
			TreasuryRewardPercentage:  e2eConfig.Hedera.TreasuryRewardPercentage,
			Sender:                    Sender(e2eConfig.Hedera.Sender),
			DbValidationProps:         make([]config.Database, len(e2eConfig.Hedera.DbValidationProps)),
			MirrorNode:                *new(config.MirrorNode).DefaultOrConfig(&e2eConfig.Hedera.MirrorNode),
		},
		EVM:             make(map[uint64]config.Evm),
		Tokens:          e2eConfig.Tokens,
		ValidatorUrl:    e2eConfig.ValidatorUrl,
		Bridge:          e2eConfig.Bridge,
		FeePercentages:  map[string]int64{},
		NftConstantFees: map[string]int64{},
		NftDynamicFees:  map[string]decimal.Decimal{},
		Scenario:        e2eConfig.Scenario,
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

	routerClients, evmFungibleTokenClients, evmNftClients := evmSetup.RouterAndEVMTokenClientsFromEVMUtils(setup.Clients.EVM)
	setup.AssetMappings = assets.NewService(e2eConfig.Bridge.Networks, e2eConfig.Hedera.BridgeAccount, configuration.FeePercentages, routerClients, setup.Clients.MirrorNode, evmFungibleTokenClients, evmNftClients)

	return setup
}

// Setup used by the e2e tests. Preloaded with all necessary dependencies
type Setup struct {
	BridgeAccount             hederaSDK.AccountID
	PayerAccount              hederaSDK.AccountID
	Treasury                  hederaSDK.AccountID
	ValidatorRewardPercentage int
	TreasuryRewardPercentage  int
	TopicID                   hederaSDK.TopicID
	TokenID                   hederaSDK.TokenID
	NativeEvmToken            string
	NftTokenID                hederaSDK.TokenID
	NftSerialNumber           int64
	NftConstantFees           map[string]int64
	NftDynamicFees            map[string]decimal.Decimal
	FeePercentages            map[string]int64
	Members                   []hederaSDK.AccountID
	Clients                   *clients
	DbValidator               *verify.Service
	AssetMappings             service.Assets
	Scenario                  *ScenarioConfig
}

// newSetup instantiates new Setup struct
func newSetup(config Config) (*Setup, error) {
	bridgeAccount, err := hederaSDK.AccountIDFromString(config.Hedera.BridgeAccount)
	if err != nil {
		return nil, err
	}

	payerAccount, err := hederaSDK.AccountIDFromString(config.Hedera.PayerAccount)
	if err != nil {
		return nil, err
	}

	topicID, err := hederaSDK.TopicIDFromString(config.Hedera.TopicID)
	if err != nil {
		return nil, err
	}

	treasuryID, err := hederaSDK.AccountIDFromString(config.Hedera.Treasury)
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
			return nil, fmt.Errorf("[%s] - invalid fee percentage [%d]", token, fp)
		}
	}

	if len(config.Hedera.Members) == 0 {
		return nil, fmt.Errorf("members account ids cannot be 0")
	}

	var members []hederaSDK.AccountID
	for _, v := range config.Hedera.Members {
		account, err := hederaSDK.AccountIDFromString(v)
		if err != nil {
			return nil, err
		}
		members = append(members, account)
	}

	scenario, err := newScenario(config)
	if err != nil {
		return nil, err
	}

	clients, err := newClients(config)
	if err != nil {
		return nil, err
	}

	clients.ValidatorClient.ExpectedValidatorsCount = scenario.ExpectedValidatorsCount
	clients.ValidatorClient.WebRetryCount = scenario.WebRetryCount
	clients.ValidatorClient.WebRetryTimeout = scenario.WebRetryTimeout

	dbValidator := verify.NewService(config.Hedera.DbValidationProps)
	dbValidator.DatabaseRetryCount = scenario.DatabaseRetryCount
	dbValidator.DatabaseRetryTimeout = scenario.DatabaseRetryTimeout
	dbValidator.ExpectedValidatorsCount = scenario.ExpectedValidatorsCount

	return &Setup{
		BridgeAccount:             bridgeAccount,
		PayerAccount:              payerAccount,
		TopicID:                   topicID,
		TokenID:                   tokenID,
		Treasury:                  treasuryID,
		ValidatorRewardPercentage: config.Hedera.ValidatorRewardPercentage,
		TreasuryRewardPercentage:  config.Hedera.TreasuryRewardPercentage,
		NftTokenID:                nftTokenID,
		NftSerialNumber:           config.Tokens.NftSerialNumber,
		NativeEvmToken:            config.Tokens.EvmNativeToken,
		NftConstantFees:           config.NftConstantFees,
		NftDynamicFees:            config.NftDynamicFees,
		FeePercentages:            config.FeePercentages,
		Members:                   members,
		Clients:                   clients,
		DbValidator:               dbValidator,
		AssetMappings:             config.AssetMappings,
		Scenario:                  scenario,
	}, nil
}

// clients used by the e2e tests
type clients struct {
	Hedera          *hederaSDK.Client
	EVM             map[uint64]evmSetup.Utils
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

	EVM := make(map[uint64]evmSetup.Utils)
	evmClients := make(map[uint64]client.EVM)
	for configChainId, conf := range config.EVM {
		evmClient := evm.NewClient(conf, configChainId)
		evmClients[configChainId] = evmClient

		clientChainId, err := evmClient.ChainID(context.Background())
		if err != nil {
			return nil, errors.New("failed to retrieve chain ID on new client")
		}

		if configChainId == clientChainId.Uint64() {
			evmClient.SetChainID(clientChainId.Uint64())
		} else {
			return nil, errors.New("chain IDs mismatch config and actual")
		}

		network, ok := config.Bridge.Networks[configChainId]
		if !ok || network.RouterContractAddress == "" {
			continue
		}

		routerContractAddress := common.HexToAddress(network.RouterContractAddress)
		routerInstance, _ := router.NewRouter(routerContractAddress, evmClient)

		signer := evm_signer.NewEVMSigner(evmClient.GetPrivateKey())
		keyTransactor, err := signer.NewKeyTransactor(clientChainId)
		if err != nil {
			return nil, err
		}

		EVM[configChainId] = evmSetup.Utils{
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
		Distributor:     distributor.New(config.Hedera.Members, config.Hedera.Treasury, config.Hedera.TreasuryRewardPercentage, config.Hedera.ValidatorRewardPercentage),
	}, nil
}

func newScenario(config Config) (*ScenarioConfig, error) {
	scenario := ScenarioConfig{
		ExpectedValidatorsCount: config.Scenario.ExpectedValidatorsCount,
		FirstEvmChainId:         config.Scenario.FirstEvmChainId,
		SecondEvmChainId:        config.Scenario.SecondEvmChainId,
		DatabaseRetryCount:      config.Scenario.DatabaseRetryCount,
		DatabaseRetryTimeout:    config.Scenario.DatabaseRetryTimeout,
		WebRetryCount:           config.Scenario.WebRetryCount,
		WebRetryTimeout:         config.Scenario.WebRetryTimeout,
		AmountHederaHbar:        config.Scenario.AmountHederaHbar,
		AmountHederaNative:      config.Scenario.AmountHederaNative,
		AmountEvmWrappedHbar:    config.Scenario.AmountEvmWrappedHbar,
		AmountEvmWrapped:        config.Scenario.AmountEvmWrapped,
		AmountEvmNative:         config.Scenario.AmountEvmNative,
		AmountHederaWrapped:     config.Scenario.AmountHederaWrapped,
	}

	// apply default scenario values if not set in yml or set with invalid values
	if scenario.DatabaseRetryCount <= 0 {
		scenario.DatabaseRetryCount = 5
	}

	if scenario.DatabaseRetryTimeout <= 0 {
		scenario.DatabaseRetryTimeout = 10
	}

	if scenario.WebRetryCount <= 0 {
		scenario.WebRetryCount = 5
	}

	if scenario.WebRetryTimeout <= 0 {
		scenario.WebRetryTimeout = 10
	}

	// amounts
	if scenario.AmountHederaHbar <= 0 {
		scenario.AmountHederaHbar = 1000000000
	}
	if scenario.AmountHederaNative <= 0 {
		scenario.AmountHederaNative = 1000000000
	}
	if scenario.AmountEvmWrappedHbar <= 0 {
		scenario.AmountEvmWrappedHbar = 100000000
	}
	if scenario.AmountEvmWrapped <= 0 {
		scenario.AmountEvmWrapped = 100000000
	}
	if scenario.AmountEvmNative <= 0 {
		scenario.AmountEvmNative = 1000000000000
	}
	if scenario.AmountHederaWrapped <= 0 {
		scenario.AmountHederaWrapped = 10
	}

	return &scenario, nil
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
	Scenario        e2eParser.ScenarioParser
}

// Hedera props from the application.yml
type Hedera struct {
	NetworkType               string
	BridgeAccount             string
	PayerAccount              string
	Members                   []string
	Treasury                  string
	ValidatorRewardPercentage int
	TreasuryRewardPercentage  int
	TopicID                   string
	Sender                    Sender
	DbValidationProps         []config.Database
	MirrorNode                config.MirrorNode
}

// Sender props from the application.yml
type Sender struct {
	Account    string
	PrivateKey string
}

type Receiver struct {
	Account string
}

type ScenarioConfig struct {
	ExpectedValidatorsCount int
	FirstEvmChainId         uint64
	SecondEvmChainId        uint64
	DatabaseRetryCount      int
	DatabaseRetryTimeout    time.Duration
	WebRetryCount           int
	WebRetryTimeout         time.Duration
	AmountHederaHbar        int64
	AmountHederaNative      int64
	AmountEvmWrappedHbar    int64
	AmountEvmWrapped        int64
	AmountEvmNative         int64
	AmountHederaWrapped     int64
}
