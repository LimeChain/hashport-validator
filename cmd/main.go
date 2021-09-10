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
	if !configuration.Node.Validator {
		log.Println("Starting Validator Node in REST-API Mode only. No Watchers or Handlers will start.")
		services = PrepareApiOnlyServices(configuration, *clients)
	} else {
		db := persistence.NewDatabase(configuration.Node.Database)
		// Prepare repositories
		repositories := PrepareRepositories(db)
		// Prepare Services
		services = PrepareServices(configuration, *clients, *repositories)

		// Execute Recovery Process. Computing Watchers starting timestamp
		err, watchersStartTimestamp := executeRecoveryProcess(configuration, *services, *repositories, *clients)
		if err != nil {
			log.Fatal(err)
		}
		initializeServerPairs(server, services, repositories, clients, configuration, watchersStartTimestamp)
	}

	apiRouter := initializeAPIRouter(services)

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Node.Port))
}

func executeRecoveryProcess(configuration config.Config, services Services, repository Repositories, client Clients) (error, int64) {
	r, err := recovery.NewProcess(configuration.Node,
		services.transfers,
		services.messages,
		services.contractServices,
		repository.transferStatus,
		repository.messageStatus,
		repository.transfer,
		client.MirrorNode,
		client.HederaNode,
		configuration.Bridge.Assets)
	if err != nil {
		log.Fatalf("Could not prepare Recovery process. Error [%s]", err)
	}
	transfersRecoveryFrom, _, recoveryTo, err := r.ComputeIntervals()
	if err != nil {
		log.Fatalf("Could not compute recovery interval. Error [%s]", err)
	}
	if transfersRecoveryFrom <= 0 {
		log.Infof("Skipping Recovery process. Nothing to recover")
	}
	//else {
	//err = r.Start(transfersRecoveryFrom, messagesRecoveryFrom, recoveryTo)
	//if err != nil {
	//	log.Fatalf("Recovery Process with interval [%d;%d] finished unsuccessfully. Error: [%s].", transfersRecoveryFrom, recoveryTo, err)
	//}
	//}
	return err, recoveryTo
}

func initializeAPIRouter(services *Services) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transfer.Route, transfer.NewRouter(services.transfers))
	apiRouter.AddV1Router(burn_event.Route, burn_event.NewRouter(services.burnEvents))
	return apiRouter
}

func initializeServerPairs(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration config.Config, watchersTimestamp int64) {
	server.AddWatcher(addTransferWatcher(
		&configuration,
		services.transfers,
		clients.MirrorNode,
		&repositories.transferStatus,
		watchersTimestamp,
		services.contractServices))

	server.AddHandler(constants.TopicMessageSubmission,
		message_submission.NewHandler(
			clients.HederaNode,
			clients.MirrorNode,
			services.signers,
			services.transfers,
			repositories.transfer,
			configuration.Bridge.TopicId))

	server.AddHandler(constants.HederaMintHtsTransfer, mint_hts.NewHandler(services.lockEvents))
	server.AddHandler(constants.HederaBurnMessageSubmission, burn_message.NewHandler(services.transfers))
	server.AddHandler(constants.HederaFeeTransfer, fee_transfer.NewHandler(services.burnEvents))
	server.AddHandler(constants.HederaTransferMessageSubmission, fee_message.NewHandler(services.transfers))

	server.AddWatcher(
		addConsensusTopicWatcher(
			&configuration,
			clients.MirrorNode,
			repositories.messageStatus,
			watchersTimestamp))
	server.AddHandler(constants.TopicMessageValidation, mh.NewHandler(
		configuration.Bridge.TopicId,
		repositories.transfer,
		repositories.message,
		services.contractServices,
		services.messages))

	for _, evmClient := range clients.EVMClients {
		chainId := evmClient.ChainID().Int64()
		server.AddWatcher(
			evm.NewWatcher(services.contractServices[chainId], evmClient, configuration.Bridge.Assets))
	}
}

func addTransferWatcher(configuration *config.Config,
	bridgeService service.Transfers,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	startTimestamp int64,
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
		startTimestamp,
		contractServices,
		configuration.Bridge.Assets)
}

func addConsensusTopicWatcher(configuration *config.Config,
	client client.MirrorNode,
	repository repository.Status,
	startTimestamp int64,
) *cmw.Watcher {
	topic := configuration.Bridge.TopicId
	log.Debugf("Added Topic Watcher for topic [%s]\n", topic)
	return cmw.NewWatcher(client,
		topic,
		repository,
		configuration.Node.Clients.MirrorNode.PollingInterval,
		startTimestamp)
}
