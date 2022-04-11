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
	coin_gecko "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-gecko"
	coin_market_cap "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-market-cap"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/werc721"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

// Clients struct used to initialise and store all available external clients for a validator node
type Clients struct {
	HederaNode              client.HederaNode
	MirrorNode              client.MirrorNode
	EVMClients              map[uint64]client.EVM
	CoinGecko               client.Pricing
	CoinMarketCap           client.Pricing
	RouterClients           map[uint64]client.DiamondRouter
	EvmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken
	EvmNFTClients           map[uint64]map[string]client.EvmNFT
}

// PrepareClients instantiates all the necessary clients for a validator node
func PrepareClients(clientsCfg config.Clients, bridgeEVMsCfgs map[uint64]config.BridgeEvm, networks map[uint64]*parser.Network) *Clients {
	EvmClients := InitEVMClients(clientsCfg)

	return &Clients{
		HederaNode:              hedera.NewNodeClient(clientsCfg.Hedera),
		MirrorNode:              mirror_node.NewClient(clientsCfg.MirrorNode),
		EVMClients:              EvmClients,
		CoinGecko:               coin_gecko.NewClient(clientsCfg.CoinGecko),
		CoinMarketCap:           coin_market_cap.NewClient(clientsCfg.CoinMarketCap),
		RouterClients:           InitRouterClients(bridgeEVMsCfgs, EvmClients),
		EvmFungibleTokenClients: InitEvmFungibleTokenClients(networks, EvmClients),
		EvmNFTClients:           InitEvmNftClients(networks, EvmClients),
	}
}

func InitEVMClients(clientsCfg config.Clients) map[uint64]client.EVM {
	EVMClients := make(map[uint64]client.EVM)
	for configChainId, ec := range clientsCfg.Evm {
		EVMClients[configChainId] = evm.NewClient(ec, configChainId)
		clientChainId, e := EVMClients[configChainId].ChainID(context.Background())
		if e != nil {
			log.Fatalf("[%d] - Failed to retrieve chain ID on client prepare.", configChainId)
		}
		if configChainId != clientChainId.Uint64() {
			log.Fatalf("Chain IDs mismatch [%d] config, [%d] actual.", configChainId, clientChainId)
		}
		EVMClients[configChainId].SetChainID(clientChainId.Uint64())
	}
	return EVMClients
}

func InitRouterClients(bridgeEVMsCfgs map[uint64]config.BridgeEvm, evmClients map[uint64]client.EVM) map[uint64]client.DiamondRouter {
	routers := make(map[uint64]client.DiamondRouter)
	for networkId, bridgeEVMsCfg := range bridgeEVMsCfgs {
		contractAddress, err := evmClients[networkId].ValidateContractDeployedAt(bridgeEVMsCfg.RouterContractAddress)
		additionalMsg := "Failed to initialize Router Contract Instance at [%s]. Error [%s]"
		if err != nil {
			log.Fatalf(additionalMsg, bridgeEVMsCfg.RouterContractAddress, err)
		}

		contractInstance, err := router.NewRouter(*contractAddress, evmClients[networkId].GetClient())
		if err != nil {
			log.Fatalf(additionalMsg, bridgeEVMsCfg.RouterContractAddress, err)
		}
		routers[networkId] = contractInstance
	}

	return routers
}

func InitEvmFungibleTokenClients(networks map[uint64]*parser.Network, evmClients map[uint64]client.EVM) map[uint64]map[string]client.EvmFungibleToken {
	tokenClients := make(map[uint64]map[string]client.EvmFungibleToken)
	for networkId, network := range networks {

		if networkId != constants.HederaNetworkId {
			if _, exist := tokenClients[networkId]; !exist {
				tokenClients[networkId] = make(map[string]client.EvmFungibleToken)
			}
		}

		// Native Tokens
		for fungibleTokenAddress, tokenInfo := range network.Tokens.Fungible {

			if networkId != constants.HederaNetworkId {
				tokenInstance, err := wtoken.NewWtoken(common.HexToAddress(fungibleTokenAddress), evmClients[networkId])
				if err != nil {
					log.Fatalf("Failed to initialize Native EvmFungibleToken Contract Instance at token address [%s]. Error [%s]", fungibleTokenAddress, err)
				}
				tokenClients[networkId][fungibleTokenAddress] = tokenInstance
			}

			// Wrapped tokens
			for wrappedNetworkId, wrappedTokenAddress := range tokenInfo.Networks {
				if wrappedNetworkId == constants.HederaNetworkId {
					continue
				}

				if _, exist := tokenClients[wrappedNetworkId]; !exist {
					tokenClients[wrappedNetworkId] = make(map[string]client.EvmFungibleToken)
				}

				wrappedTokenInstance, err := wtoken.NewWtoken(common.HexToAddress(wrappedTokenAddress), evmClients[wrappedNetworkId])
				if err != nil {
					log.Fatalf("Failed to initialize Wrapped EvmFungibleToken Contract Instance at token address [%s]. Error [%s]", wrappedTokenAddress, err)
				}
				tokenClients[wrappedNetworkId][wrappedTokenAddress] = wrappedTokenInstance
			}
		}

	}

	return tokenClients
}

func InitEvmNftClients(networks map[uint64]*parser.Network, evmClients map[uint64]client.EVM) map[uint64]map[string]client.EvmNFT {
	tokenClients := make(map[uint64]map[string]client.EvmNFT)
	for networkId, network := range networks {

		if networkId != constants.HederaNetworkId {
			if _, exist := tokenClients[networkId]; !exist {
				tokenClients[networkId] = make(map[string]client.EvmNFT)
			}
		}

		// Native Tokens
		for fungibleTokenAddress, tokenInfo := range network.Tokens.Nft {

			if networkId != constants.HederaNetworkId {
				tokenInstance, err := werc721.NewWerc721(common.HexToAddress(fungibleTokenAddress), evmClients[networkId])
				if err != nil {
					log.Fatalf("Failed to initialize Native EvmFungibleToken Contract Instance at token address [%s]. Error [%s]", fungibleTokenAddress, err)
				}
				tokenClients[networkId][fungibleTokenAddress] = tokenInstance
			}

			// Wrapped tokens
			for wrappedNetworkId, wrappedTokenAddress := range tokenInfo.Networks {
				if wrappedNetworkId == constants.HederaNetworkId {
					continue
				}

				if _, exist := tokenClients[wrappedNetworkId]; !exist {
					tokenClients[wrappedNetworkId] = make(map[string]client.EvmNFT)
				}

				wrappedTokenInstance, err := werc721.NewWerc721(common.HexToAddress(wrappedTokenAddress), evmClients[wrappedNetworkId])
				if err != nil {
					log.Fatalf("Failed to initialize Wrapped EvmFungibleToken Contract Instance at token address [%s]. Error [%s]", wrappedTokenAddress, err)
				}
				tokenClients[wrappedNetworkId][wrappedTokenAddress] = wrappedTokenInstance
			}
		}

	}

	return tokenClients
}
