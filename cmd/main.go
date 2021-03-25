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
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	cmh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/message"
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
	"github.com/limechain/hedera-watcher-sdk/server"
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
		server.AddHandler(process.CryptoTransferMessageType, th.NewHandler(services.transfers))

		err = addTransferWatchers(&configuration, services.transfers, clients.MirrorNode, &repositories.transferStatus, server, watchersStartTimestamp, services.contracts)
		if err != nil {
			log.Fatal(err)
		}

		server.AddHandler(process.HCSMessageType, cmh.NewHandler(
			configuration.Hedera.Handler.ConsensusMessage,
			repositories.message,
			services.contracts,
			services.messages))

		err = addConsensusTopicWatcher(&configuration, clients.HederaNode, repositories.messageStatus, server, watchersStartTimestamp)
		if err != nil {
			log.Fatal(err)
		}
		server.AddWatcher(ethereum.NewWatcher(services.contracts, configuration.Hedera.Eth))
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
		log.Fatalf("Could not prepare Recovery process. Err %s", err)
	}
	transfersRecoveryFrom, messagesRecoveryFrom, recoveryTo, err := r.ComputeIntervals()
	if err != nil {
		log.Fatalf("Could not compute recovery interval. Error %s", err)
	}
	if transfersRecoveryFrom <= 0 {
		log.Infof("Skipping Recovery process. Nothing to recover")
	} else {
		err = r.Start(transfersRecoveryFrom, messagesRecoveryFrom, recoveryTo)
		if err != nil {
			log.Fatalf("Recovery Process with interval [%d;%d] finished unsuccessfully. Error: %s", transfersRecoveryFrom, recoveryTo, err)
		}
	}
	return err, recoveryTo
}

func initializeAPIRouter(services *Services) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(metadata.Route, metadata.NewRouter(services.fees))
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transaction.Route, transaction.NewRouter(services.messages))
	return apiRouter
}

func addTransferWatchers(configuration *config.Config,
	bridgeService service.Transfers,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	server *server.HederaWatcherServer,
	startTimestamp int64,
	contractService service.Contracts,
) error {
	account := configuration.Hedera.Watcher.CryptoTransfer.Account
	id, e := hedera.AccountIDFromString(account.Id)
	if e != nil {
		return errors.New(fmt.Sprintf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", account.Id, e))
	}

	server.AddWatcher(
		tw.NewWatcher(
			bridgeService,
			mirrorNode,
			id,
			configuration.Hedera.MirrorNode.PollingInterval,
			*repository,
			account.MaxRetries,
			startTimestamp,
			contractService))
	log.Debugf("Added Transfer Watcher for account [%s]", account.Id)
	return nil
}

func addConsensusTopicWatcher(configuration *config.Config,
	hederaNodeClient client.HederaNode,
	repository repository.Status,
	server *server.HederaWatcherServer,
	startTimestamp int64,
) error {
	topic := configuration.Hedera.Watcher.ConsensusMessage.Topic
	id, e := hedera.TopicIDFromString(topic.Id)
	if e != nil {
		return errors.New(fmt.Sprintf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topic.Id, e))
	}

	server.AddWatcher(cmw.NewWatcher(hederaNodeClient, id, repository, startTimestamp))
	log.Debugf("Added Topic Watcher for topic [%s]\n", topic.Id)
	return nil
}
