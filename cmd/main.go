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
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/server"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
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
	"github.com/limechain/hedera-eth-bridge-validator/app/process/recovery"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/evm"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/price"
	pw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/prometheus"
	tw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/transfer"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/router/burn-event"
	config_bridge "github.com/limechain/hedera-eth-bridge-validator/app/router/config-bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	min_amounts "github.com/limechain/hedera-eth-bridge-validator/app/router/min-amounts"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	// Config
	configuration, parsedBridge := config.LoadConfig()
	config.InitLogger(configuration.Node.LogLevel)

	// Prepare Clients
	clients := PrepareClients(configuration.Node.Clients, configuration.Bridge.EVMs)

	// Prepare Node
	server := server.NewServer()

	var services *Services = nil
	db := persistence.NewDatabase(configuration.Node.Database)
	// Prepare repositories
	repositories := PrepareRepositories(db)

	// Prepare Services
	services = PrepareServices(configuration, parsedBridge, *clients, *repositories)

	// Set Assets Service
	configuration.Bridge.Assets = services.assets

	initializeServerPairs(server, services, repositories, clients, configuration)

	initializeMonitoring(services.prometheus, server, configuration, clients.MirrorNode, clients.EVMClients)

	apiRouter := initializeAPIRouter(services, parsedBridge)

	executeRecovery(repositories.fee, repositories.schedule, clients.MirrorNode)

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Node.Port))
}

func initializeMonitoring(
	prometheusService service.Prometheus,
	s *server.Server,
	configuration config.Config,
	mirrorNode client.MirrorNode,
	EVMClients map[uint64]client.EVM,
) {
	if configuration.Node.Monitoring.Enable {
		initializePrometheusWatcher(s, configuration, mirrorNode, prometheusService, EVMClients)
	} else {
		log.Infoln("Monitoring is disabled. No metrics will be added.")
	}
}

func initializeAPIRouter(services *Services, bridgeConfig parser.Bridge) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transfer.Route, transfer.NewRouter(services.transfers))
	apiRouter.AddV1Router(burn_event.Route, burn_event.NewRouter(services.burnEvents))
	apiRouter.AddV1Router("/metrics", promhttp.Handler())
	apiRouter.AddV1Router(config_bridge.Route, config_bridge.NewRouter(bridgeConfig))
	apiRouter.AddV1Router(min_amounts.Route, min_amounts.NewRouter(services.pricing))

	return apiRouter
}

func executeRecovery(feeRepository repository.Fee, scheduleRepository repository.Schedule, client client.MirrorNode) {
	r := recovery.New(feeRepository, scheduleRepository, client)

	r.Execute()
}

func initializeServerPairs(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration config.Config) {
	server.AddWatcher(addTransferWatcher(
		&configuration,
		services.transfers,
		clients.MirrorNode,
		&repositories.transferStatus,
		services.contractServices,
		services.prometheus,
		services.pricing))

	server.AddHandler(constants.TopicMessageSubmission,
		message_submission.NewHandler(
			clients.HederaNode,
			clients.MirrorNode,
			services.transfers,
			repositories.transfer,
			services.messages,
			configuration.Bridge.TopicId))

	server.AddHandler(constants.HederaMintHtsTransfer, mint_hts.NewHandler(services.lockEvents))
	server.AddHandler(constants.HederaBurnMessageSubmission, burn_message.NewHandler(services.transfers))
	server.AddHandler(constants.HederaFeeTransfer, fee_transfer.NewHandler(services.burnEvents))
	server.AddHandler(constants.HederaTransferMessageSubmission, fee_message.NewHandler(services.transfers))

	server.AddWatcher(
		addConsensusTopicWatcher(
			&configuration,
			clients.MirrorNode,
			repositories.messageStatus))
	server.AddHandler(constants.TopicMessageValidation, mh.NewHandler(
		configuration.Bridge.TopicId,
		repositories.transfer,
		repositories.message,
		services.contractServices,
		services.messages,
		services.prometheus,
		configuration.Bridge.Assets))

	for _, evmClient := range clients.EVMClients {
		chain, err := evmClient.ChainID(context.Background())
		if err != nil {
			panic(err)
		}
		contractService := services.contractServices[chain.Uint64()]
		// Given that addresses between different
		// EVM networks might be the same, a concatenation between
		// <chain-id>-<contract-address> removes possible duplication.
		dbIdentifier := fmt.Sprintf("%d-%s", chain.Uint64(), contractService.Address().String())

		server.AddWatcher(
			evm.NewWatcher(
				repositories.transferStatus,
				contractService,
				services.prometheus,
				services.pricing,
				evmClient,
				configuration.Bridge.Assets,
				dbIdentifier,
				configuration.Node.Clients.Evm[chain.Uint64()].StartBlock,
				configuration.Node.Validator,
				configuration.Node.Clients.Evm[chain.Uint64()].PollingInterval,
				configuration.Node.Clients.Evm[chain.Uint64()].MaxLogsBlocks,
			))
	}

	// Register read-only handlers
	server.AddHandler(constants.ReadOnlyHederaTransfer, rfth.NewHandler(
		repositories.transfer,
		repositories.fee,
		repositories.schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.distributor,
		services.fees,
		services.transfers,
		services.readOnly,
		services.prometheus))
	server.AddHandler(constants.ReadOnlyHederaFeeTransfer, rfh.NewHandler(
		repositories.transfer,
		repositories.fee,
		repositories.schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.distributor,
		services.fees,
		services.transfers,
		services.readOnly,
		services.prometheus))
	server.AddHandler(constants.ReadOnlyHederaBurn, rbh.NewHandler(
		configuration.Bridge.Hedera.BridgeAccount,
		clients.MirrorNode,
		repositories.schedule,
		services.transfers,
		services.readOnly))
	server.AddHandler(constants.ReadOnlyHederaMintHtsTransfer, rmth.NewHandler(
		repositories.schedule,
		configuration.Bridge.Hedera.BridgeAccount,
		clients.MirrorNode,
		services.transfers,
		services.readOnly,
		services.prometheus))
	server.AddHandler(constants.ReadOnlyTransferSave, rthh.NewHandler(services.transfers))

	// Hedera Native Nft handlers
	server.AddHandler(constants.HederaNativeNftTransfer, nfmh.NewHandler(services.transfers))
	server.AddHandler(constants.ReadOnlyHederaNativeNftTransfer, rnfmh.NewHandler(
		repositories.transfer,
		repositories.fee,
		repositories.schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.distributor,
		services.transfers,
		configuration.Bridge.Hedera.NftFees,
		services.readOnly))

	// Hedera Native unlock Nft Handlers
	server.AddHandler(constants.HederaNftTransfer, nth.NewHandler(
		configuration.Bridge.Hedera.BridgeAccount,
		repositories.transfer,
		repositories.schedule,
		services.transfers,
		services.scheduled))
	server.AddHandler(constants.ReadOnlyHederaUnlockNftTransfer, rnth.NewHandler(
		configuration.Bridge.Hedera.BridgeAccount,
		repositories.transfer,
		repositories.schedule,
		services.readOnly,
		services.transfers))

	server.AddWatcher(price.NewWatcher(services.pricing))
}

func initializePrometheusWatcher(
	server *server.Server,
	configuration config.Config,
	mirrorNode client.MirrorNode,
	prometheusService service.Prometheus,
	EVMClients map[uint64]client.EVM,
) {
	dashboardPolling := configuration.Node.Monitoring.DashboardPolling * time.Minute
	log.Infoln("Dashboard Polling interval: ", dashboardPolling)
	server.AddWatcher(addPrometheusWatcher(
		dashboardPolling,
		mirrorNode,
		configuration,
		prometheusService,
		EVMClients))
}

func addTransferWatcher(configuration *config.Config,
	bridgeService service.Transfers,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	contractServices map[uint64]service.Contracts,
	prometheusService service.Prometheus,
	pricingService service.Pricing,
) *tw.Watcher {
	account := configuration.Bridge.Hedera.BridgeAccount

	log.Debugf("Added Transfer Watcher for account [%s]", account)
	return tw.NewWatcher(
		bridgeService,
		mirrorNode,
		account,
		configuration.Node.Clients.MirrorNode.PollingInterval,
		*repository,
		configuration.Node.Clients.Hedera.StartTimestamp,
		contractServices,
		configuration.Bridge.Assets,
		configuration.Bridge.Hedera.NftFees,
		configuration.Node.Validator,
		prometheusService,
		pricingService,
	)
}

func addConsensusTopicWatcher(configuration *config.Config,
	client client.MirrorNode,
	repository repository.Status,
) *cmw.Watcher {
	topic := configuration.Bridge.TopicId
	log.Debugf("Added Topic Watcher for topic [%s]\n", topic)
	return cmw.NewWatcher(client,
		topic,
		repository,
		configuration.Node.Clients.MirrorNode.PollingInterval,
		configuration.Node.Clients.Hedera.StartTimestamp)
}

func addPrometheusWatcher(
	dashboardPolling time.Duration,
	mirrorNode client.MirrorNode,
	configuration config.Config,
	prometheusService service.Prometheus,
	EVMClients map[uint64]client.EVM,
) *pw.Watcher {
	log.Debugf("Added Prometheus Watcher for dashboard metrics")
	return pw.NewWatcher(
		dashboardPolling,
		mirrorNode,
		configuration,
		prometheusService,
		EVMClients,
		configuration.Bridge.Assets)
}
