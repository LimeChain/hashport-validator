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

package bootstrap

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gookit/event"
	coin_gecko "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-gecko"
	coin_market_cap "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-market-cap"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	mirrornode "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	eventHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/events"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

// Clients struct used to initialise and store all available external clients for a validator node
type Clients struct {
	HederaNode              client.HederaNode
	MirrorNode              client.MirrorNode
	EvmClients              map[uint64]client.EVM
	CoinGecko               client.Pricing
	CoinMarketCap           client.Pricing
	RouterClients           map[uint64]client.DiamondRouter
	EvmRegularTokenClients map[uint64]map[string]client.EvmRegularToken
	ClientsConfig           config.Clients
}

// PrepareClients instantiates all the necessary clients for a validator node
func PrepareClients(clientsCfg config.Clients, parsedBridge *parser.Bridge) *Clients {
	EvmClients := InitEVMClients(clientsCfg, parsedBridge.Networks.EVM)
	instance := &Clients{
		HederaNode:             hedera.NewNodeClient(clientsCfg.Hedera),
		MirrorNode:             mirrornode.NewClient(clientsCfg.MirrorNode),
		EvmClients:             EvmClients,
		CoinGecko:              coin_gecko.NewClient(clientsCfg.CoinGecko),
		CoinMarketCap:          coin_market_cap.NewClient(clientsCfg.CoinMarketCap),
		RouterClients:          InitRouterClients(parsedBridge.Networks.EVM, EvmClients),
		EvmRegularTokenClients: InitEvmRegularTokenClients(parsedBridge.Networks.EVM, EvmClients, parsedBridge.RegularTokens),
		ClientsConfig:          clientsCfg,
	}

	event.On(constants.EventBridgeConfigUpdate, event.ListenerFunc(func(e event.Event) error {
		return bridgeCfgEventHandler(e, instance)
	}), constants.ClientsEventPriority)

	return instance
}

func bridgeCfgEventHandler(e event.Event, instance *Clients) error {
	params, err := eventHelper.GetBridgeCfgUpdateEventParams(e)
	if err != nil {
		return err
	}
	instance.EvmClients = InitEVMClients(instance.ClientsConfig, params.ParsedBridge.Networks.EVM)
	evmRegularTokenClients := InitEvmRegularTokenClients(params.ParsedBridge.Networks.EVM, instance.EvmClients, params.ParsedBridge.RegularTokens)
	routerClients := InitRouterClients(params.ParsedBridge.Networks.EVM, instance.EvmClients)
	for networkId, ftClients := range evmRegularTokenClients {
		_, ok := instance.EvmRegularTokenClients[networkId]
		if !ok {
			instance.EvmRegularTokenClients[networkId] = make(map[string]client.EvmRegularToken)
		}
		for key, ftClient := range ftClients {
			instance.EvmRegularTokenClients[networkId][key] = ftClient
		}
	}

	for networkId, routerClient := range routerClients {
		instance.RouterClients[networkId] = routerClient
	}

	params.EvmRegularTokenClients = evmRegularTokenClients
	params.RouterClients = routerClients

	return nil
}

func InitEVMClients(clientsCfg config.Clients, evmNetworks map[uint64]*parser.EVMNetwork) map[uint64]client.EVM {
	EVMClients := make(map[uint64]client.EVM)
	for configChainId, ec := range clientsCfg.EvmPool {
		network, ok := evmNetworks[configChainId]
		if !ok || network.RouterContractAddress == "" {
			continue
		}
		evmClient, e := evm.NewClientPool(ec, configChainId)
		if e != nil {
			log.Fatalf("[%d] - Failed to initialize EVM Client. Error: [%s]", configChainId, e)
		}
		EVMClients[configChainId] = evmClient
		clientChainId, e := EVMClients[configChainId].ChainID(context.Background())
		if e != nil {
			log.Fatalf("[%d] - Failed to retrieve chain ID on client prepare. Error: [%s]", configChainId, e)
		}
		if configChainId != clientChainId.Uint64() {
			log.Fatalf("Chain IDs mismatch [%d] config, [%d] actual.", configChainId, clientChainId)
		}
		EVMClients[configChainId].SetChainID(clientChainId.Uint64())
	}
	return EVMClients
}

func InitRouterClients(evmNetworks map[uint64]*parser.EVMNetwork, evmClients map[uint64]client.EVM) map[uint64]client.DiamondRouter {
	routers := make(map[uint64]client.DiamondRouter)
	for networkId, evmNetwork := range evmNetworks {
		if evmNetwork.RouterContractAddress == "" {
			continue
		}
		evmClient, ok := evmClients[networkId]
		if !ok {
			log.Fatalf("failed to initialize RouterClient because of missing EVM client for network id: [%d]", networkId)
		}
		contractAddress, err := evmClient.ValidateContractDeployedAt(evmNetwork.RouterContractAddress)
		additionalMsg := "Failed to initialize Router Contract Instance at [%s]. Error [%s]"
		if err != nil {
			log.Fatalf(additionalMsg, evmNetwork.RouterContractAddress, err)
		}

		contractInstance, err := router.NewRouter(*contractAddress, evmClient.GetClient())
		if err != nil {
			log.Fatalf(additionalMsg, evmNetwork.RouterContractAddress, err)
		}
		routers[networkId] = contractInstance
	}

	return routers
}

func InitEvmRegularTokenClients(evmNetworks map[uint64]*parser.EVMNetwork, evmClients map[uint64]client.EVM, regularTokens map[string]*parser.RegularToken) map[uint64]map[string]client.EvmRegularToken {
	tokenClients := make(map[uint64]map[string]client.EvmRegularToken)
	for networkId := range evmNetworks {
		if _, ok := tokenClients[networkId]; !ok {
			tokenClients[networkId] = make(map[string]client.EvmRegularToken)
		}
	}

	for _, tokenInfo := range regularTokens {
		nativeChainId := tokenInfo.NativeChain

		// Native EVM token
		if tokenInfo.Address != nil && *tokenInfo.Address != "" {
			if _, isEVM := evmNetworks[nativeChainId]; isEVM {
				if evmClient, ok := evmClients[nativeChainId]; ok {
					nativeAddr := *tokenInfo.Address
					tokenInstance, err := wtoken.NewWtoken(common.HexToAddress(nativeAddr), evmClient)
					if err != nil {
						log.Fatalf("Failed to initialize Native EvmRegularToken Contract Instance at token address [%s]. Error [%s]", nativeAddr, err)
					}
					tokenClients[nativeChainId][nativeAddr] = tokenInstance
				}
			}
		}

		// Wrapped tokens on EVM networks
		for wrappedNetworkId, wrappedTokenAddress := range tokenInfo.AddressesPerNetwork {
			if _, isEVM := evmNetworks[wrappedNetworkId]; !isEVM {
				continue
			}

			if _, ok := tokenClients[wrappedNetworkId]; !ok {
				tokenClients[wrappedNetworkId] = make(map[string]client.EvmRegularToken)
			}

			evmClient, ok := evmClients[wrappedNetworkId]
			if !ok {
				continue
			}

			wrappedTokenInstance, err := wtoken.NewWtoken(common.HexToAddress(wrappedTokenAddress), evmClient)
			if err != nil {
				log.Fatalf("Failed to initialize Wrapped EvmRegularToken Contract Instance at token address [%s]. Error [%s]", wrappedTokenAddress, err)
			}
			tokenClients[wrappedNetworkId][wrappedTokenAddress] = wrappedTokenInstance
		}
	}

	return tokenClients
}