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
	"errors"
	"fmt"

	"github.com/limechain/hedera-eth-bridge-validator/app/services/assets"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	evmSetup "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/evm"
	e2eParser "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/parser"

	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/verify"
)

// Setup used by the e2e tests. Preloaded with all necessary dependencies
type Setup struct {
	BridgeAccount   hederaSDK.AccountID
	TopicID         hederaSDK.TopicID
	TokenID         hederaSDK.TokenID
	NativeEvmToken  string
	NftTokenID      hederaSDK.TokenID
	NftSerialNumber int64
	NftFees         map[string]int64
	FeePercentages  map[string]int64
	Members         []hederaSDK.AccountID
	Clients         *clients
	DbValidator     *verify.Service
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

	dbValidator := verify.NewService(config.Hedera.DbValidationProps)

	return &Setup{
		BridgeAccount:   bridgeAccount,
		TopicID:         topicID,
		TokenID:         tokenID,
		NftTokenID:      nftTokenID,
		NftSerialNumber: config.Tokens.NftSerialNumber,
		NativeEvmToken:  config.Tokens.EvmNativeToken,
		NftFees:         config.NftFees,
		FeePercentages:  config.FeePercentages,
		Members:         members,
		Clients:         clients,
		DbValidator:     dbValidator,
		AssetMappings:   config.AssetMappings,
	}, nil
}

// Load loads the e2e application.yml from the ./e2e/setup folder and parses it to suitable working struct for the e2e tests
func Load() *Setup {
	var e2eConfig e2eParser.Config
	if err := config.GetConfig(&e2eConfig, e2eConfigPath); err != nil {
		panic(err)
	}
	if err := config.GetConfig(&e2eConfig, e2eBridgeConfigPath); err != nil {
		panic(err)
	}

	configuration := Config{
		Hedera: Hedera{
			NetworkType:       e2eConfig.Hedera.NetworkType,
			BridgeAccount:     e2eConfig.Hedera.BridgeAccount,
			Members:           e2eConfig.Hedera.Members,
			TopicID:           e2eConfig.Hedera.TopicID,
			Sender:            Sender(e2eConfig.Hedera.Sender),
			DbValidationProps: make([]config.Database, len(e2eConfig.Hedera.DbValidationProps)),
			MirrorNode:        config.MirrorNode(e2eConfig.Hedera.MirrorNode),
		},
		EVM:            make(map[uint64]config.Evm),
		Tokens:         e2eConfig.Tokens,
		ValidatorUrl:   e2eConfig.ValidatorUrl,
		Bridge:         e2eConfig.Bridge,
		FeePercentages: map[string]int64{},
		NftFees:        map[string]int64{},
	}

	if e2eConfig.Bridge.Networks[constants.HederaNetworkId] != nil {
		feePercentages, nftFees := config.LoadHederaFees(e2eConfig.Bridge.Networks[constants.HederaNetworkId].Tokens)
		configuration.FeePercentages = feePercentages
		configuration.NftFees = nftFees
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
