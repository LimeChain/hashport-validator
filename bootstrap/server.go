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
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/server"
	burn_message "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/burn-message"
	fee_message "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/fee-message"
	fee_transfer "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/fee-transfer"
	mh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/message"
	message_submission "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/message-submission"
	mint_hts "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/mint-hts"
	nfmh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/nft/fee-message"
	nth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/nft/transfer"
	rbh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/burn"
	rfh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/fee"
	rfth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/fee-transfer"
	rmth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/mint-hts"
	rnfmh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/nft/fee"
	rnth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/nft/transfer"
	rthh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/transfer"
	bridge_config "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/bridge-config"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/evm"
	fee_policy "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/fee-policy-config"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/price"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

func InitializeServerPairs(
	server *server.Server,
	services *Services,
	repositories *Repositories,
	clients *Clients,
	configuration *config.Config,
	parsedBridge *parser.Bridge,
	bridgeCfgTopicId hedera.TopicID,
	parsedFeePolicyTopicId hedera.TopicID) {
	// Transfer Message Watcher
	registerTransferWatcher(server, services, repositories, clients, configuration)

	// Transfer Message Handlers
	registerTransferMessageHandlers(server, services, repositories, clients, configuration)

	// Validation
	registerValidationServerPairs(server, services, repositories, clients, configuration)

	// Evm Clients
	registerEvmClients(server, services, repositories, clients, configuration)

	// Read-only handlers
	registerReadOnlyHandlers(server, services, repositories, clients, configuration)

	// Hedera Native Nft handlers
	registerHederaNativeNFTHandlers(server, services, repositories, clients, configuration)

	// Hedera Native unlock Nft Handlers
	registerHederaNativeUnlockNftHandlers(server, services, repositories, configuration)

	// Assets Watcher
	registerAssetsWatcher(server, services, configuration, clients)

	// Prometheus Watcher
	registerPrometheusWatcher(server, services, configuration, clients)

	// Pricing Watcher
	server.AddWatcher(price.NewWatcher(services.Pricing))

	// Bridge Config Watcher
	registerBridgeConfigWatcher(server, services, parsedBridge.UseLocalConfig, bridgeCfgTopicId, parsedBridge.PollingInterval)

	// Fee policy watcher
	registerFeePolicyWatcher(server, services, parsedFeePolicyTopicId, parsedBridge.PollingInterval)
}

func registerFeePolicyWatcher(server *server.Server, services *Services, parsedFeePolicyTopicId hedera.TopicID, pollingInterval time.Duration) {
	server.AddWatcher(fee_policy.NewWatcher(services.FeePolicyHandler, parsedFeePolicyTopicId, pollingInterval))
}

func registerBridgeConfigWatcher(server *server.Server, services *Services, useLocalConfig bool, bridgeCfgTopicId hedera.TopicID, pollingInterval time.Duration) {
	if useLocalConfig {
		log.Infoln("Using local bridge config. Skipping initialization of BridgeConfigWatcher ...")
	} else {
		server.AddWatcher(bridge_config.NewWatcher(services.BridgeConfig, bridgeCfgTopicId, pollingInterval))
	}
}

func registerTransferWatcher(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration *config.Config) {
	server.AddWatcher(createTransferWatcher(
		configuration,
		services.transfers,
		services.Assets,
		clients.MirrorNode,
		&repositories.TransferStatus,
		services.ContractServices,
		services.Prometheus,
		services.Pricing))
}

func registerValidationServerPairs(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration *config.Config) {
	// Watcher - ConsensusTopic
	server.AddWatcher(
		createConsensusTopicWatcher(
			configuration,
			clients.MirrorNode,
			repositories.MessageStatus))

	// Handler - TopicMessageValidation
	server.AddHandler(constants.TopicMessageValidation, mh.NewHandler(
		configuration.Bridge.TopicId,
		repositories.Transfer,
		repositories.Message,
		services.ContractServices,
		services.Messages,
		services.Prometheus,
		services.Assets))
}

func registerTransferMessageHandlers(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration *config.Config) {
	// TopicMessageSubmission
	server.AddHandler(constants.TopicMessageSubmission,
		message_submission.NewHandler(
			clients.HederaNode,
			clients.MirrorNode,
			services.transfers,
			repositories.Transfer,
			services.Messages,
			configuration.Bridge.TopicId))

	// HederaMintHtsTransfer
	server.AddHandler(constants.HederaMintHtsTransfer, mint_hts.NewHandler(services.LockEvents))

	// HederaBurnMessageSubmission
	server.AddHandler(constants.HederaBurnMessageSubmission, burn_message.NewHandler(services.transfers))

	// HederaFeeTransfer
	server.AddHandler(constants.HederaFeeTransfer, fee_transfer.NewHandler(services.BurnEvents))

	// HederaTransferMessageSubmission
	server.AddHandler(constants.HederaTransferMessageSubmission, fee_message.NewHandler(services.transfers))
}

func registerEvmClients(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration *config.Config) {
	for _, evmClient := range clients.EvmClients {
		chain := evmClient.GetChainID()
		contractService := services.ContractServices[chain]
		// Given that addresses between different
		// EVM networks might be the same, a concatenation between
		// <chain-id>-<contract-address> removes possible duplication.
		dbIdentifier := fmt.Sprintf("%d-%s", chain, contractService.Address().String())

		server.AddWatcher(
			evm.NewWatcher(
				repositories.TransferStatus,
				contractService,
				services.Prometheus,
				services.Pricing,
				evmClient,
				services.Assets,
				dbIdentifier,
				configuration.Node.Clients.Evm[chain].StartBlock,
				configuration.Node.Validator,
				configuration.Node.Clients.Evm[chain].PollingInterval,
				configuration.Node.Clients.Evm[chain].MaxLogsBlocks,
			))
	}
}

func registerAssetsWatcher(server *server.Server, services *Services, configuration *config.Config, clients *Clients) {
	server.AddWatcher(createAssetsWatcher(
		clients.MirrorNode,
		configuration,
		clients.EvmFungibleTokenClients,
		clients.EvmNFTClients,
		services.Assets))
}

func registerPrometheusWatcher(server *server.Server, services *Services, configuration *config.Config, clients *Clients) {
	if configuration.Node.Monitoring.Enable {
		dashboardPolling := configuration.Node.Monitoring.DashboardPolling * time.Minute
		log.Infoln("Dashboard Polling interval: ", dashboardPolling)
		server.AddWatcher(createPrometheusWatcher(
			dashboardPolling,
			clients.MirrorNode,
			configuration,
			services.Prometheus,
			clients.EvmFungibleTokenClients,
			clients.EvmNFTClients,
			services.Assets))
	} else {
		log.Infoln("Monitoring is disabled. No metrics will be added.")
	}
}

func registerHederaNativeUnlockNftHandlers(server *server.Server, services *Services, repositories *Repositories, configuration *config.Config) {
	// HederaNftTransfer
	server.AddHandler(constants.HederaNftTransfer, nth.NewHandler(
		configuration.Bridge.Hedera.BridgeAccount,
		repositories.Transfer,
		repositories.Schedule,
		services.transfers,
		services.Scheduled))

	// ReadOnlyHederaUnlockNftTransfer
	server.AddHandler(constants.ReadOnlyHederaUnlockNftTransfer, rnth.NewHandler(
		configuration.Bridge.Hedera.BridgeAccount,
		configuration.Bridge.Hedera.PayerAccount,
		repositories.Transfer,
		repositories.Schedule,
		services.ReadOnly,
		services.transfers))
}

func registerHederaNativeNFTHandlers(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration *config.Config) {
	// HederaNativeNftTransfer
	server.AddHandler(constants.HederaNativeNftTransfer, nfmh.NewHandler(services.transfers))

	// ReadOnlyHederaNativeNftTransfer
	server.AddHandler(constants.ReadOnlyHederaNativeNftTransfer, rnfmh.NewHandler(
		repositories.Transfer,
		repositories.Fee,
		repositories.Schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.Distributor,
		services.transfers,
		services.ReadOnly))
}

func registerReadOnlyHandlers(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration *config.Config) {
	// ReadOnlyHederaTransfer
	server.AddHandler(constants.ReadOnlyHederaTransfer, rfth.NewHandler(
		repositories.Transfer,
		repositories.Fee,
		repositories.Schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.Distributor,
		services.Fees,
		services.transfers,
		services.ReadOnly,
		services.Prometheus))

	// ReadOnlyHederaFeeTransfer
	server.AddHandler(constants.ReadOnlyHederaFeeTransfer, rfh.NewHandler(
		repositories.Transfer,
		repositories.Fee,
		repositories.Schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.Distributor,
		services.Fees,
		services.transfers,
		services.ReadOnly,
		services.Prometheus))

	// ReadOnlyHederaBurn
	server.AddHandler(constants.ReadOnlyHederaBurn, rbh.NewHandler(
		configuration.Bridge.Hedera.BridgeAccount,
		clients.MirrorNode,
		repositories.Schedule,
		services.transfers,
		services.ReadOnly))

	// ReadOnlyHederaMintHts
	server.AddHandler(constants.ReadOnlyHederaMintHtsTransfer, rmth.NewHandler(
		repositories.Schedule,
		repositories.Transfer,
		configuration.Bridge.Hedera.BridgeAccount,
		configuration.Bridge.Hedera.PayerAccount,
		clients.MirrorNode,
		services.transfers,
		services.ReadOnly,
		services.Prometheus))

	// ReadOnlyTransferSave
	server.AddHandler(constants.ReadOnlyTransferSave, rthh.NewHandler(services.transfers))
}
