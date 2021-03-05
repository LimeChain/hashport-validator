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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	cmh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/consensus-message"
	cth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/crypto-transfer"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/ethereum"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/metadata"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
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
	repository := PrepareRepositories(configuration.Hedera.Validator.Db)

	// Prepare Clients
	client := PrepareClients(configuration)

	// Prepare Services
	ethSigner := eth.NewEthSigner(configuration.Hedera.Client.Operator.EthPrivateKey)
	contractService := bridge.NewContractService(client.ethereum, configuration.Hedera.Eth)
	schedulerService := scheduler.NewScheduler(configuration.Hedera.Handler.ConsensusMessage.TopicId, ethSigner.Address(),
		configuration.Hedera.Handler.ConsensusMessage.SendDeadline, contractService, client.hederaNode)

	feeCalculator := fees.NewCalculator(client.exchangeRate, configuration.Hedera, contractService)

	// Prepare Node
	server := server.NewServer()

	server.AddHandler(process.CryptoTransferMessageType, cth.NewHandler(
		configuration.Hedera.Handler.CryptoTransfer,
		ethSigner,
		client.mirrorNode,
		client.hederaNode,
		repository.transaction,
		feeCalculator))

	err := addCryptoTransferWatchers(configuration, client.mirrorNode, repository.cryptoTransferStatus, server)
	if err != nil {
		log.Fatal(err)
	}

	server.AddHandler(process.HCSMessageType, cmh.NewHandler(
		configuration.Hedera.Handler.ConsensusMessage,
		repository.message,
		repository.transaction,
		client.ethereum,
		client.hederaNode,
		schedulerService,
		contractService,
		ethSigner))

	err = addConsensusTopicWatchers(configuration, client.hederaNode, client.mirrorNode, repository.consensusMessageStatus, server)
	if err != nil {
		log.Fatal(err)
	}
	server.AddWatcher(ethereum.NewEthereumWatcher(contractService, configuration.Hedera.Eth))

	// Register API
	apiRouter := initializeAPIRouter(feeCalculator)

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func initializeAPIRouter(feeCalculator *fees.Calculator) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(metadata.MetadataRoute, metadata.NewMetadataRouter(feeCalculator))
	apiRouter.AddV1Router(healthcheck.HealthCheckRoute, healthcheck.NewHealthCheckRouter())
	return apiRouter
}

func addCryptoTransferWatchers(configuration config.Config, hederaClient clients.MirrorNode, repository repositories.Status, server *server.HederaWatcherServer) error {
	if len(configuration.Hedera.Watcher.CryptoTransfer.Accounts) == 0 {
		log.Warnln("CryptoTransfer Accounts list is empty. No Crypto Transfer Watchers will be started")
	}
	for _, account := range configuration.Hedera.Watcher.CryptoTransfer.Accounts {
		id, e := hedera.AccountIDFromString(account.Id)
		if e != nil {
			return errors.New(fmt.Sprintf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", account.Id, e))
		}

		server.AddWatcher(cryptotransfer.NewCryptoTransferWatcher(hederaClient, id, configuration.Hedera.MirrorNode.PollingInterval, repository, account.MaxRetries, account.StartTimestamp))
		log.Infof("Added a Crypto Transfer Watcher for account [%s]\n", account.Id)
	}
	return nil
}

func addConsensusTopicWatchers(configuration config.Config, hederaNodeClient clients.HederaNode, hederaMirrorClient clients.MirrorNode, repository repositories.Status, server *server.HederaWatcherServer) error {
	if len(configuration.Hedera.Watcher.ConsensusMessage.Topics) == 0 {
		log.Warnln("Consensus Message Topics list is empty. No Consensus Topic Watchers will be started")
	}
	for _, topic := range configuration.Hedera.Watcher.ConsensusMessage.Topics {
		id, e := hedera.TopicIDFromString(topic.Id)
		if e != nil {
			return errors.New(fmt.Sprintf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topic.Id, e))
		}

		server.AddWatcher(cmw.NewConsensusTopicWatcher(hederaNodeClient, hederaMirrorClient, id, repository, topic.MaxRetries, topic.StartTimestamp))
		log.Infof("Added a Consensus Topic Watcher for topic [%s]\n", topic.Id)
	}
	return nil
}
