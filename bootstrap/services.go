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
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/assets"
	bridge_config "github.com/limechain/hedera-eth-bridge-validator/app/services/bridge-config"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/services/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/contracts"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/calculator"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/services/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/messages"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/pricing"
	prometheusServices "github.com/limechain/hedera-eth-bridge-validator/app/services/prometheus"
	read_only "github.com/limechain/hedera-eth-bridge-validator/app/services/read-only"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduled"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/transfers"
	utilsSvc "github.com/limechain/hedera-eth-bridge-validator/app/services/utils"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

type Services struct {
	Signers          map[uint64]service.Signer
	ContractServices map[uint64]service.Contracts
	transfers        service.Transfers
	Messages         service.Messages
	BurnEvents       service.BurnEvent
	LockEvents       service.LockEvent
	Fees             service.Fee
	Distributor      service.Distributor
	Scheduled        service.Scheduled
	ReadOnly         service.ReadOnly
	Prometheus       service.Prometheus
	Pricing          service.Pricing
	Assets           service.Assets
	Utils            service.Utils
	BridgeConfig     service.BridgeConfig
}

// PrepareServices instantiates all the necessary services with their required context and parameters
func PrepareServices(c *config.Config, parsedBridge *parser.Bridge, clients *Clients, repositories Repositories, parsedBridgeConfigTopicId hedera.TopicID) *Services {

	bridgeCfgService := bridge_config.NewService(c, parsedBridge, clients.MirrorNode)
	if !parsedBridge.UseLocalConfig {
		var err error
		fetchedParsedBridge, err := bridgeCfgService.ProcessLatestConfig(parsedBridgeConfigTopicId)
		if err != nil {
			panic(fmt.Sprintf("failed to process latest bridge config from topic. Err: [%s]", err))
		}
		parsedBridge.Update(fetchedParsedBridge)
	}

	evmSigners := make(map[uint64]service.Signer)
	contractServices := make(map[uint64]service.Contracts)
	assetsService := assets.NewService(
		parsedBridge.Networks,
		parsedBridge.Networks[constants.HederaNetworkId].BridgeAccount,
		c.Bridge.Hedera.FeePercentages,
		clients.RouterClients,
		clients.MirrorNode,
		clients.EvmFungibleTokenClients,
		clients.EvmNFTClients,
	)
	c.Bridge.LoadStaticMinAmountsForWrappedFungibleTokens(*parsedBridge, assetsService)

	for _, client := range clients.EvmClients {
		chainId := client.GetChainID()
		evmSigners[chainId] = evm.NewEVMSigner(client.GetPrivateKey())
		contractServices[chainId] = contracts.NewService(client, c.Bridge.EVMs[chainId].RouterContractAddress, clients.RouterClients[chainId])
	}

	fees := calculator.New(c.Bridge.Hedera.FeePercentages)
	distributor := distributor.New(c.Bridge.Hedera.Members)
	scheduled := scheduled.New(c.Bridge.Hedera.PayerAccount, clients.HederaNode, clients.MirrorNode)

	prometheus := prometheusServices.NewService(assetsService, c.Node.Monitoring.Enable)
	messages := messages.NewService(
		evmSigners,
		contractServices,
		repositories.Transfer,
		repositories.Message,
		clients.MirrorNode,
		clients.EvmClients,
		c.Bridge.TopicId,
		assetsService)

	transfers := transfers.NewService(
		clients.HederaNode,
		clients.MirrorNode,
		contractServices,
		repositories.Transfer,
		repositories.Schedule,
		repositories.Fee,
		fees,
		distributor,
		c.Bridge.TopicId,
		c.Bridge.Hedera.BridgeAccount,
		scheduled,
		messages,
		prometheus,
		assetsService)

	burnEvent := burn_event.NewService(
		c.Bridge.Hedera.BridgeAccount,
		repositories.Transfer,
		repositories.Schedule,
		repositories.Fee,
		distributor,
		scheduled,
		fees,
		transfers,
		prometheus)

	lockEvent := lock_event.NewService(
		c.Bridge.Hedera.BridgeAccount,
		repositories.Transfer,
		repositories.Schedule,
		scheduled,
		transfers,
		prometheus)

	readOnly := read_only.New(clients.MirrorNode, repositories.Transfer, c.Node.Clients.MirrorNode.PollingInterval)

	pricingService := pricing.NewService(
		c.Bridge,
		assetsService,
		clients.RouterClients,
		clients.MirrorNode,
		clients.CoinGecko,
		clients.CoinMarketCap)

	utilsService := utilsSvc.New(clients.EvmClients, burnEvent)

	return &Services{
		Signers:          evmSigners,
		ContractServices: contractServices,
		transfers:        transfers,
		Messages:         messages,
		BurnEvents:       burnEvent,
		LockEvents:       lockEvent,
		Fees:             fees,
		Distributor:      distributor,
		Scheduled:        scheduled,
		ReadOnly:         readOnly,
		Prometheus:       prometheus,
		Pricing:          pricingService,
		Assets:           assetsService,
		Utils:            utilsService,
		BridgeConfig:     bridgeCfgService,
	}
}
