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
	coin_gecko "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-gecko"
	coin_market_cap "github.com/limechain/hedera-eth-bridge-validator/app/clients/coin-market-cap"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	clientsHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/clients"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Clients struct used to initialise and store all available external clients for a validator node
type Clients struct {
	HederaNode    client.HederaNode
	MirrorNode    client.MirrorNode
	EVMClients    map[uint64]client.EVM
	CoinGecko     client.Pricing
	CoinMarketCap client.Pricing
	Routers       map[uint64]*router.Router
}

// PrepareClients instantiates all the necessary clients for a validator node
func PrepareClients(clientsCfg config.Clients, bridgeEVMsCfgs map[uint64]config.BridgeEvm) *Clients {
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

	return &Clients{
		HederaNode:    hedera.NewNodeClient(clientsCfg.Hedera),
		MirrorNode:    mirror_node.NewClient(clientsCfg.MirrorNode),
		EVMClients:    EVMClients,
		CoinGecko:     coin_gecko.NewClient(clientsCfg.CoinGecko),
		CoinMarketCap: coin_market_cap.NewClient(clientsCfg.CoinMarketCap),
		Routers:       clientsHelper.InitRouterClients(bridgeEVMsCfgs, EVMClients),
	}
}