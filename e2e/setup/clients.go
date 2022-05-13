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

	fee "github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"

	"github.com/ethereum/go-ethereum/common"
	hederaSDK "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	evm_signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/bootstrap"
	e2eClients "github.com/limechain/hedera-eth-bridge-validator/e2e/clients"
	evmSetup "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/evm"
)

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
		Distributor:     distributor.New(config.Hedera.Members),
	}, nil
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
