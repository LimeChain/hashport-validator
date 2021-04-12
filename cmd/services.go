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
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/messages"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/transfers"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

// TODO extract new service only for Ethereum TX handling
type Services struct {
	signer     service.Signer
	scheduler  service.Scheduler
	contracts  service.Contracts
	transfers  service.Transfers
	messages   service.Messages
	fees       service.Fees
	burnEvents service.BurnEvent
}

// PrepareServices instantiates all the necessary services with their required context and parameters
func PrepareServices(c config.Config, clients Clients, repositories Repositories) *Services {
	ethSigner := eth.NewEthSigner(c.Hedera.Client.Operator.EthPrivateKey)
	contracts := contracts.NewService(clients.Ethereum, c.Hedera.Eth)
	schedulerService := scheduler.NewScheduler(c.Hedera.Handler.ConsensusMessage.SendDeadline)
	fees := fees.NewCalculator(clients.ExchangeRate, c.Hedera, contracts)

	transfers := transfers.NewService(
		clients.HederaNode,
		clients.MirrorNode,
		contracts,
		fees,
		ethSigner,
		repositories.transfer,
		c.Hedera.Watcher.ConsensusMessage.Topic.Id)

	messages := messages.NewService(
		ethSigner,
		contracts,
		schedulerService,
		repositories.transfer,
		repositories.message,
		clients.HederaNode,
		clients.MirrorNode,
		clients.Ethereum,
		c.Hedera.Handler.ConsensusMessage.TopicId)

	burnEvent := burn_event.NewService(
		c.Hedera.Client.ThresholdAccount,
		c.Hedera.Client.PayerAccount,
		clients.HederaNode,
		clients.MirrorNode,
		repositories.burnEvent)

	return &Services{
		signer:     ethSigner,
		scheduler:  schedulerService,
		contracts:  contracts,
		transfers:  transfers,
		messages:   messages,
		fees:       fees,
		burnEvents: burnEvent,
	}
}

// PrepareApiOnlyServices instantiates all the necessary services with their
// required context and parameters for running the Validator node in API Only mode
func PrepareApiOnlyServices(c config.Config, clients Clients) *Services {
	contractService := contracts.NewService(clients.Ethereum, c.Hedera.Eth)
	feeService := fees.NewCalculator(clients.ExchangeRate, c.Hedera, contractService)

	return &Services{
		contracts: contractService,
		fees:      feeService,
	}
}
