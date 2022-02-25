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
	"context"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"log"
)

// Clients struct used to initialise and store all available external clients for a validator node
type Clients struct {
	HederaNode client.HederaNode
	MirrorNode client.MirrorNode
	EVMClients map[uint64]client.EVM
}

// PrepareClients instantiates all the necessary clients for a validator node
func PrepareClients(config config.Clients) *Clients {
	EVMClients := make(map[uint64]client.EVM)
	for configChainId, ec := range config.Evm {
		EVMClients[configChainId] = evm.NewClient(ec, configChainId)
		clientChainId, e := EVMClients[configChainId].ChainID(context.Background())
		if e != nil {
			log.Fatalf("[%d] - Failed to retrieve chain ID on client prepare.", configChainId)
			return nil
		}
		if configChainId == clientChainId.Uint64() {
			EVMClients[configChainId].SetChainID(clientChainId.Uint64())
		} else {
			log.Fatalf("Chain IDs mismatch [%d] config, [%d] actual.", configChainId, clientChainId)
			return nil
		}
	}

	return &Clients{
		HederaNode: hedera.NewNodeClient(config.Hedera),
		MirrorNode: mirror_node.NewClient(config.MirrorNode.ApiAddress, config.MirrorNode.PollingInterval),
		EVMClients: EVMClients,
	}
}
