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
	beh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/burn-event"
	mh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/message"
	th "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/recovery"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/ethereum"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/message"
	tw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/transfer"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/metadata"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Config
	configuration := config.LoadConfig()
	config.InitLogger(configuration.Hedera.LogLevel)

	// Prepare Clients
	clients := PrepareClients(configuration)

	// Prepare Node
	server := server.NewServer()

	var services *Services = nil
	if configuration.Hedera.RestApiOnly {
		log.Println("Starting Validator Node in REST-API Mode only. No Watchers or Handlers will start.")
		services = PrepareApiOnlyServices(configuration, *clients)
	} else {
		// Prepare repositories
		repositories := PrepareRepositories(configuration.Hedera.Validator.Db)
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
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func executeRecoveryProcess(configuration config.Config, services Services, repository Repositories, client Clients) (error, int64) {
	r, err := recovery.NewProcess(configuration.Hedera,
		services.transfers,
		services.messages,
		services.contracts,
		repository.transferStatus,
		repository.messageStatus,
		repository.transfer,
		client.MirrorNode,
		client.HederaNode)
	if err != nil {
		log.Fatalf("Could not prepare Recovery process. Error [%s]", err)
	}
	transfersRecoveryFrom, messagesRecoveryFrom, recoveryTo, err := r.ComputeIntervals()
	if err != nil {
		log.Fatalf("Could not compute recovery interval. Error [%s]", err)
	}
	if transfersRecoveryFrom <= 0 {
		log.Infof("Skipping Recovery process. Nothing to recover")
	} else {
		err = r.Start(transfersRecoveryFrom, messagesRecoveryFrom, recoveryTo)
		if err != nil {
			log.Fatalf("Recovery Process with interval [%d;%d] finished unsuccessfully. Error: [%s].", transfersRecoveryFrom, recoveryTo, err)
		}
	}
	return err, recoveryTo
}

func initializeAPIRouter(services *Services) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(metadata.Route, metadata.NewRouter(services.fees))
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transaction.Route, transaction.NewRouter(services.transfers))
	return apiRouter
}

func initializeServerPairs(server *server.Server, services *Services, repositories *Repositories, clients *Clients, configuration config.Config, watchersTimestamp int64) {
	server.AddPair(
		addTransferWatcher(
			&configuration,
			services.transfers,
			clients.MirrorNode,
			&repositories.transferStatus,
			watchersTimestamp,
			services.contracts),
		th.NewHandler(services.transfers))

	server.AddPair(
		addConsensusTopicWatcher(
			&configuration,
			clients.MirrorNode,
			repositories.messageStatus,
			watchersTimestamp),
		mh.NewHandler(
			configuration.Hedera.Handler.ConsensusMessage,
			repositories.transfer,
			repositories.message,
			services.contracts,
			services.messages))

	server.AddPair(ethereum.NewWatcher(services.contracts, clients.Ethereum, configuration.Hedera.Eth),
		beh.NewHandler(services.burnEvents))
}

func addTransferWatcher(configuration *config.Config,
	bridgeService service.Transfers,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	startTimestamp int64,
	contractService service.Contracts,
) *tw.Watcher {
	account := configuration.Hedera.Watcher.CryptoTransfer.Account

	log.Debugf("Added Transfer Watcher for account [%s]", account.Id)
	return tw.NewWatcher(
		bridgeService,
		mirrorNode,
		account.Id,
		configuration.Hedera.MirrorNode.PollingInterval,
		*repository,
		account.MaxRetries,
		startTimestamp,
		contractService)
}

func addConsensusTopicWatcher(configuration *config.Config,
	client client.MirrorNode,
	repository repository.Status,
	startTimestamp int64,
) *cmw.Watcher {
	topic := configuration.Hedera.Watcher.ConsensusMessage.Topic
	log.Debugf("Added Topic Watcher for topic [%s]\n", topic.Id)
	return cmw.NewWatcher(client,
		topic.Id,
		repository,
		configuration.Hedera.MirrorNode.PollingInterval,
		topic.MaxRetries,
		startTimestamp)
}
