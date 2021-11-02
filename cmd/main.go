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
	rbh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/burn"
	rfh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/fee"
	rfth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/fee-transfer"
	rmth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/mint-hts"
	rthh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/read-only/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/recovery"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/evm"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/message"
	tw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/transfer"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/router/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Config
	configuration := config.LoadConfig()
	config.InitLogger(configuration.Node.LogLevel)

	// Prepare Clients
	clients := PrepareClients(configuration.Node.Clients)

	// Prepare Node
	server := server.NewServer()

	var services *Services = nil
	db := persistence.NewDatabase(configuration.Node.Database)
	// Prepare repositories
	repositories := PrepareRepositories(db)
	// Prepare Services
	services = PrepareServices(configuration, *clients, *repositories)

	initializeServerPairs(server, services, repositories, clients, configuration)

	apiRouter := initializeAPIRouter(services)

	executeRecovery(repositories.fee, repositories.schedule, clients.MirrorNode)

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Node.Port))
}

func initializeAPIRouter(services *Services) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transfer.Route, transfer.NewRouter(services.transfers))
	apiRouter.AddV1Router(burn_event.Route, burn_event.NewRouter(services.burnEvents))
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
		services.contractServices))

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
		services.messages))

	for _, evmClient := range clients.EVMClients {
		chain, err := evmClient.ChainID(context.Background())
		if err != nil {
			panic(err)
		}
		server.AddWatcher(
			evm.NewWatcher(
				repositories.transferStatus,
				services.contractServices[chain.Int64()],
				evmClient,
				configuration.Bridge.Assets,
				configuration.Node.Clients.Evm[chain.Int64()].StartBlock,
				configuration.Node.Validator,
				configuration.Node.Clients.Evm[chain.Int64()].PollingInterval))
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
		services.readOnly))
	server.AddHandler(constants.ReadOnlyHederaFeeTransfer, rfh.NewHandler(
		repositories.transfer,
		repositories.fee,
		repositories.schedule,
		clients.MirrorNode,
		configuration.Bridge.Hedera.BridgeAccount,
		services.distributor,
		services.fees,
		services.transfers,
		services.readOnly))
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
		services.readOnly))
	server.AddHandler(constants.ReadOnlyTransferSave, rthh.NewHandler(services.transfers))
}

func addTransferWatcher(configuration *config.Config,
	bridgeService service.Transfers,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	contractServices map[int64]service.Contracts,
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
		configuration.Node.Validator)
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
