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

package main

import (
	coin_gecko_web_api "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-gecko/web-api"
	coin_market_cap_web_api "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-market-cap/web-api"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	hedera_web_api "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/web-api"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	clientsHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/clients"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Clients struct used to initialise and store all available external clients for a validator node
type Clients struct {
	HederaNode          client.HederaNode
	MirrorNode          client.MirrorNode
	EVMClients          map[uint64]client.EVM
	HederaWebApi        client.HederaWebAPI
	CoinGeckoWebApi     client.CoinGeckoWebAPI
	CoinMarketCapWebApi client.CoinMarketCapWebAPI
	Routers             map[uint64]*router.Router
	logger              *log.Entry
}

// PrepareClients instantiates all the necessary clients for a validator node
func PrepareClients(clientsCfg config.Clients, bridgeEVMsCfgs map[uint64]config.BridgeEvm) *Clients {
	EVMClients := make(map[uint64]client.EVM)
	for chainId, ec := range clientsCfg.Evm {
		EVMClients[chainId] = evm.NewClient(ec, chainId)
	}

	logger := config.GetLoggerFor("Clients")

	return &Clients{
		HederaNode:          hedera.NewNodeClient(clientsCfg.Hedera),
		MirrorNode:          mirror_node.NewClient(clientsCfg.MirrorNode.ApiAddress, clientsCfg.MirrorNode.PollingInterval),
		EVMClients:          EVMClients,
		HederaWebApi:        hedera_web_api.NewClient(clientsCfg.WebAPIs.Hedera),
		CoinGeckoWebApi:     coin_gecko_web_api.NewClient(clientsCfg.WebAPIs.CoinGecko),
		CoinMarketCapWebApi: coin_market_cap_web_api.NewClient(clientsCfg.WebAPIs.CoinMarketCap),
		Routers:             clientsHelper.InitRouterClients(bridgeEVMsCfgs, EVMClients, logger),
		logger:              logger,
	}
}
