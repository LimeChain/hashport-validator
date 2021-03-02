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

package main

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/exchange-rate"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

// Clients struct used to initialise and store all available external clients for a validator node
type Clients struct {
	hederaNode   clients.HederaNode
	mirrorNode   clients.MirrorNode
	ethereum     clients.Ethereum
	exchangeRate clients.ExchangeRate
}

// PrepareClients instantiates all the necessary clients for a validator node
func PrepareClients(config config.Config) *Clients {
	return &Clients{
		hederaNode:   hedera.NewNodeClient(config.Hedera.Client),
		mirrorNode:   hedera.NewMirrorNodeClient(config.Hedera.MirrorNode.ApiAddress),
		ethereum:     ethereum.NewClient(config.Hedera.Eth),
		exchangeRate: exchangerate.NewProvider("hedera-hashgraph", "eth"),
	}
}
