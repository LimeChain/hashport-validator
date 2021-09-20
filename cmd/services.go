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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/services/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/contracts"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/services/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/messages"
	read_only "github.com/limechain/hedera-eth-bridge-validator/app/services/read-only"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduled"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/transfers"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

type Services struct {
	signers          map[int64]service.Signer
	contractServices map[int64]service.Contracts
	transfers        service.Transfers
	messages         service.Messages
	burnEvents       service.BurnEvent
	lockEvents       service.LockEvent
	fees             service.Fee
	distributor      service.Distributor
	scheduled        service.Scheduled
	readOnly         service.ReadOnly
}

// PrepareServices instantiates all the necessary services with their required context and parameters
func PrepareServices(c config.Config, clients Clients, repositories Repositories) *Services {
	evmSigners := make(map[int64]service.Signer)
	contractServices := make(map[int64]service.Contracts)
	for _, client := range clients.EVMClients {
		chainId := client.ChainID().Int64()
		evmSigners[chainId] = evm.NewEVMSigner(client.GetPrivateKey())
		contractServices[chainId] = contracts.NewService(client, c.Bridge.EVMs[chainId].RouterContractAddress)
	}

	fees := calculator.New(c.Bridge.Hedera.FeePercentages)
	distributor := distributor.New(c.Bridge.Hedera.Members)
	scheduled := scheduled.New(c.Bridge.Hedera.PayerAccount, clients.HederaNode, clients.MirrorNode)

	messages := messages.NewService(
		evmSigners,
		contractServices,
		repositories.transfer,
		repositories.message,
		clients.MirrorNode,
		clients.EVMClients,
		c.Bridge.TopicId,
		c.Bridge.Assets)

	transfers := transfers.NewService(
		clients.HederaNode,
		clients.MirrorNode,
		contractServices,
		repositories.transfer,
		repositories.schedule,
		repositories.fee,
		fees,
		distributor,
		c.Bridge.TopicId,
		c.Bridge.Hedera.BridgeAccount,
		scheduled,
		messages)

	burnEvent := burn_event.NewService(
		c.Bridge.Hedera.BridgeAccount,
		repositories.transfer,
		repositories.schedule,
		repositories.fee,
		distributor,
		scheduled,
		fees)

	lockEvent := lock_event.NewService(
		c.Bridge.Hedera.BridgeAccount,
		repositories.transfer,
		repositories.schedule,
		scheduled)

	readOnly := read_only.New(clients.MirrorNode, repositories.transfer)

	return &Services{
		signers:          evmSigners,
		contractServices: contractServices,
		transfers:        transfers,
		messages:         messages,
		burnEvents:       burnEvent,
		lockEvents:       lockEvent,
		fees:             fees,
		distributor:      distributor,
		readOnly:         readOnly,
	}
}

// PrepareApiOnlyServices instantiates all the necessary services with their
// required context and parameters for running the Validator node in API Only mode
func PrepareApiOnlyServices(c config.Config, clients Clients) *Services {
	contractServices := make(map[int64]service.Contracts)
	for _, client := range clients.EVMClients {
		contractService := contracts.NewService(client, c.Bridge.EVMs[client.ChainID().Int64()].RouterContractAddress)
		contractServices[client.ChainID().Int64()] = contractService
	}

	return &Services{
		contractServices: contractServices,
	}
}
