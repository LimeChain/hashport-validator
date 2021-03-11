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
	"flag"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	cmh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/consensus-message"
	th "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/recovery"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	tw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/ethereum"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/metadata"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/metadata"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Parse Flags
	debugMode := flag.Bool("debug", false, "run in debug mode")
	flag.Parse()

	// Config
	config.InitLogger(debugMode)
	configuration := config.LoadConfig()

	// Prepare repositories
	repositories := PrepareRepositories(configuration.Hedera.Validator.Db)

	// Prepare Clients
	clients := PrepareClients(configuration)

	// Prepare Services
	services := PrepareServices(configuration, *clients, *repositories)

	apiRouter := initializeAPIRouter(services.fees)

	// Prepare Node
	server := server.NewServer()

	if !configuration.Hedera.RestApiOnly {
		// Execute Recovery Process. Computing Watchers starting timestamp
		err, watchersStartTimestamp := executeRecoveryProcess(configuration, *services, *repositories, *clients)
		server.AddHandler(process.CryptoTransferMessageType, th.NewHandler(services.transfers))

		err = addCryptoTransferWatcher(&configuration, services.transfers, clients.MirrorNode, &repositories.cryptoTransferStatus, server, watchersStartTimestamp)
		if err != nil {
			log.Fatal(err)
		}

		server.AddHandler(process.HCSMessageType, cmh.NewHandler(
			configuration.Hedera.Handler.ConsensusMessage,
			repositories.message,
			services.contracts,
			services.signatures))

		err = addConsensusTopicWatcher(&configuration, clients.HederaNode, repositories.consensusMessageStatus, server, watchersStartTimestamp)
		if err != nil {
			log.Fatal(err)
		}
		server.AddWatcher(ethereum.NewEthereumWatcher(services.contracts, configuration.Hedera.Eth))
	} else {
		log.Println("Starting Validator Node in REST-API Mode only. No Watchers or Handlers will start.")
	}

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))

}

func executeRecoveryProcess(configuration config.Config, services Services, repository Repositories, client Clients) (error, int64) {
	r, err := recovery.NewProcess(configuration.Hedera, services.transfers, services.signatures, repository.cryptoTransferStatus, client.MirrorNode, client.HederaNode)
	if err != nil {
		log.Fatalf("Could not prepare Recovery process. Err %s", err)
	}
	recoveryFrom, recoveryTo, err := r.ComputeInterval()
	if err != nil {
		log.Fatalf("Could not compute recovery interval. Error %s", err)
	}
	if recoveryFrom <= 0 {
		log.Infof("Skipping Recovery process. Nothing to recover")
	} else {
		err = r.Start(recoveryFrom, recoveryTo)
		if err != nil {
			log.Fatalf("Recovery Process with interval [%d;%d] finished unsuccessfully. Error: %s", recoveryFrom, recoveryTo, err)
		}
	}
	return err, recoveryTo
}

func initializeAPIRouter(feeCalculator service.Fees) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(metadata.MetadataRoute, metadata.NewRouter(feeCalculator))
	apiRouter.AddV1Router(healthcheck.HealthCheckRoute, healthcheck.NewRouter())
	return apiRouter
}

func addCryptoTransferWatcher(configuration *config.Config,
	bridgeService service.Transfers,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	server *server.HederaWatcherServer,
	startTimestamp int64,
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
			startTimestamp))
	log.Infof("Added Transfer Watcher for account [%s]", account.Id)
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
	log.Infof("Added Topic Watcher for topic [%s]\n", topic.Id)
	return nil
}
