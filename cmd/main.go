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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	cth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/recovery"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/ethereum"
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/metadata"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/contracts"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
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
	client := clients.PrepareClients(configuration)

	// Prepare Services
	ethSigner := eth.NewEthSigner(configuration.Hedera.Client.Operator.EthPrivateKey)
	contractService := contracts.NewService(client.Ethereum, configuration.Hedera.Eth)
	//_ := scheduler.NewScheduler(configuration.Hedera.Handler.ConsensusMessage.TopicId, ethSigner.Address(),
	//	configuration.Hedera.Handler.ConsensusMessage.SendDeadline, contractService, client.HederaNode)
	feeCalculator := fees.NewCalculator(client.ExchangeRate, configuration.Hedera, contractService)
	bridgeService := bridge.NewService(
		*client,
		repository.transaction,
		repository.message,
		contractService,
		feeCalculator,
		ethSigner,
		configuration.Hedera.Watcher.ConsensusMessage.Topic.Id)

	// Execute Recovery Process. Computing Watchers starting timestamp
	err, recoveryTo := executeRecoveryProcess(configuration, bridgeService, repository, client)

	//from := r.getStartTimestampFor(r.accountStatusRepository, r.accountID.String())
	//to := time.Now().UnixNano()
	//if from < 0 {
	//	log.Info("Nothing to recover. Proceeding to start watchers and handlers.")
	//	return to, nil
	//}

	//now, err := recoverLostProgress(configuration.Hedera,
	//	&repository.transaction,
	//	&repository.cryptoTransferStatus,
	//	&repository.consensusMessageStatus,
	//	&client.mirrorNode,
	//	&client.hederaNode,
	//	bridgeService)
	//if err != nil {
	//	log.Fatalf("Could not recover last records of topics or accounts: Error - [%s]", err)
	//}

	server := server.NewServer()

	server.AddHandler(process.CryptoTransferMessageType, cth.NewHandler(
		configuration.Hedera.Handler.CryptoTransfer,
		ethSigner,
		client.MirrorNode,
		client.HederaNode,
		repository.transaction,
		bridgeService))

	err = addCryptoTransferWatcher(&configuration, bridgeService, client.MirrorNode, &repository.cryptoTransferStatus, server, recoveryTo)
	if err != nil {
		log.Fatal(err)
	}

	//server.AddHandler(process.HCSMessageType, cmh.NewHandler(
	//	configuration.Hedera.Handler.ConsensusMessage,
	//	repository.message,
	//	repository.transaction,
	//	client.Ethereum,
	//	client.HederaNode,
	//	schedulerService,
	//	contractService,
	//	ethSigner,
	//	bridgeService))
	//
	//err = addConsensusTopicWatcher(&configuration, &client.hederaNode, &client.mirrorNode, &repository.consensusMessageStatus, server, now)
	//if err != nil {
	//	log.Fatal(err)
	//}
	server.AddWatcher(ethereum.NewEthereumWatcher(contractService, configuration.Hedera.Eth))

	// Register API
	apiRouter := initializeAPIRouter(feeCalculator)

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func executeRecoveryProcess(configuration config.Config, bridgeService *bridge.Service, repository *Repositories, client *clients.Clients) (error, int64) {
	r, err := recovery.NewProcess(configuration.Hedera, bridgeService, repository.cryptoTransferStatus, client.MirrorNode, client.HederaNode)
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

func initializeAPIRouter(feeCalculator *fees.Calculator) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(metadata.NewMetadataRouter(feeCalculator))

	return apiRouter
}

func addCryptoTransferWatcher(configuration *config.Config,
	bridgeService services.Bridge,
	mirrorNode clients.MirrorNode,
	repository *repositories.Status,
	server *server.HederaWatcherServer,
	startTimestamp int64,
) error {
	account := configuration.Hedera.Watcher.CryptoTransfer.Account
	id, e := hedera.AccountIDFromString(account.Id)
	if e != nil {
		return errors.New(fmt.Sprintf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", account.Id, e))
	}

	server.AddWatcher(
		cryptotransfer.NewWatcher(
			bridgeService,
			mirrorNode,
			id,
			configuration.Hedera.MirrorNode.PollingInterval,
			*repository,
			account.MaxRetries,
			startTimestamp))
	log.Infof("Added a Crypto Transfer Watcher for account [%s]\n", account.Id)
	return nil
}

//func addConsensusTopicWatcher(configuration *config.Config,
//	hederaNodeClient *clients.HederaNode,
//	hederaMirrorClient *clients.MirrorNode,
//	repository *repositories.Status,
//	server *server.HederaWatcherServer,
//	startTimestamp int64,
//) error {
//	topic := configuration.Hedera.Watcher.ConsensusMessage.Topic
//	id, e := hedera.TopicIDFromString(topic.Id)
//	if e != nil {
//		return errors.New(fmt.Sprintf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topic.Id, e))
//	}
//
//	server.AddWatcher(cmw.NewConsensusTopicWatcher(*hederaNodeClient, *hederaMirrorClient, id, *repository, topic.MaxRetries, startTimestamp))
//	log.Infof("Added a Consensus Topic Watcher for topic [%s]\n", topic.Id)
//	return nil
//}
