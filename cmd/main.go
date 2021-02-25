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
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/metadata"

	"github.com/hashgraph/hedera-sdk-go/v2"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	exchangerate "github.com/limechain/hedera-eth-bridge-validator/app/clients/exchange-rate"
	hederaClients "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	cmh "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/consensus-message"
	cth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/crypto-transfer"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	debugMode := flag.Bool("debug", false, "run in debug mode")
	flag.Parse()
	config.InitLogger(debugMode)
	configuration := config.LoadConfig()
	db := persistence.RunDb(configuration.Hedera.Validator.Db)
	hederaMirrorClient := hederaClients.NewHederaMirrorClient(configuration.Hedera.MirrorNode.ApiAddress)
	hederaNodeClient := hederaClients.NewNodeClient(configuration.Hedera.Client)
	ethClient := ethclient.NewEthereumClient(configuration.Hedera.Eth)
	ethSigner := eth.NewEthSigner(configuration.Hedera.Client.Operator.EthPrivateKey)
	contractService := bridge.NewBridgeContractService(ethClient, configuration.Hedera.Eth)
	schedulerService := scheduler.NewScheduler(configuration.Hedera.Handler.ConsensusMessage.TopicId, ethSigner.Address(),
		configuration.Hedera.Handler.ConsensusMessage.SendDeadline, contractService, hederaNodeClient)

	transactionRepository := transaction.NewTransactionRepository(db)
	statusCryptoTransferRepository := status.NewStatusRepository(db, process.CryptoTransferMessageType)
	statusConsensusMessageRepository := status.NewStatusRepository(db, process.HCSMessageType)
	messageRepository := message.NewMessageRepository(db)
	exchangeRateService := exchangerate.NewExchangeRateProvider("hedera-hashgraph", "eth")

	feeCalculator := fees.NewFeeCalculator(&exchangeRateService, configuration.Hedera)

	server := server.NewServer()

	server.AddHandler(process.CryptoTransferMessageType, cth.NewCryptoTransferHandler(
		configuration.Hedera.Handler.CryptoTransfer,
		ethSigner,
		hederaMirrorClient,
		hederaNodeClient,
		transactionRepository,
		feeCalculator))

	err := addCryptoTransferWatchers(configuration, hederaMirrorClient, statusCryptoTransferRepository, server)
	if err != nil {
		log.Fatal(err)
	}

	server.AddHandler(process.HCSMessageType, cmh.NewConsensusMessageHandler(
		configuration.Hedera.Handler.ConsensusMessage,
		configuration.Hedera.Eth.BridgeContractAddress,
		*messageRepository,
		transactionRepository,
		ethClient,
		hederaNodeClient,
		schedulerService,
		ethSigner))

	err = addConsensusTopicWatchers(configuration, hederaNodeClient, hederaMirrorClient, statusConsensusMessageRepository, server)
	if err != nil {
		log.Fatal(err)
	}

	apiRouter := initializeAPIRouter(feeCalculator)

	server.AddWatcher(ethereum.NewEthereumWatcher(contractService, configuration.Hedera.Eth, hederaNodeClient))

	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func initializeAPIRouter(feeCalculator *fees.FeeCalculator) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(metadata.NewMetadataRouter(feeCalculator))

	return apiRouter
}

func addCryptoTransferWatchers(configuration *config.Config, hederaClient *hederaClients.HederaMirrorClient, repository *status.StatusRepository, server *server.HederaWatcherServer) error {
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

func addConsensusTopicWatchers(configuration *config.Config, hederaNodeClient *hederaClients.HederaNodeClient, hederaMirrorClient *hederaClients.HederaMirrorClient, repository *status.StatusRepository, server *server.HederaWatcherServer) error {
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
