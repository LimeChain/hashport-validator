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
	"github.com/limechain/hedera-eth-bridge-validator/app/services/recovery"

	"github.com/hashgraph/hedera-sdk-go"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
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
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	debugMode := flag.Bool("debug", true, "run in debug mode")
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

	now, err := recoverLostProgress(configuration.Hedera,
		transactionRepository,
		statusCryptoTransferRepository,
		statusConsensusMessageRepository,
		messageRepository,
		hederaMirrorClient,
		ethClient,
		hederaNodeClient)
	if err != nil {
		log.Fatalf("Could not recover last records of topics or accounts: Error - [%s]", err)
	}

	server := server.NewServer()

	server.AddHandler(process.CryptoTransferMessageType, cth.NewCryptoTransferHandler(
		configuration.Hedera.Handler.CryptoTransfer,
		ethSigner,
		hederaMirrorClient,
		hederaNodeClient,
		transactionRepository,
		ethClient))

	err = addCryptoTransferWatcher(configuration, hederaMirrorClient, statusCryptoTransferRepository, server, now)
	if err != nil {
		log.Fatal(err)
	}

	server.AddHandler(process.HCSMessageType, cmh.NewConsensusMessageHandler(
		configuration.Hedera.Handler.ConsensusMessage,
		*messageRepository,
		transactionRepository,
		ethClient,
		hederaNodeClient,
		schedulerService,
		ethSigner))

	err = addConsensusTopicWatcher(configuration, hederaNodeClient, hederaMirrorClient, statusConsensusMessageRepository, server, now)
	if err != nil {
		log.Fatal(err)
	}

	server.AddWatcher(ethereum.NewEthereumWatcher(contractService, configuration.Hedera.Eth))

	server.Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func addCryptoTransferWatcher(configuration *config.Config,
	hederaClient *hederaClients.HederaMirrorClient,
	repository *status.StatusRepository,
	server *server.HederaWatcherServer,
	startTimestamp int64,
) error {
	account := configuration.Hedera.Watcher.CryptoTransfer.Account
	id, e := hedera.AccountIDFromString(account.Id)
	if e != nil {
		return errors.New(fmt.Sprintf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", account.Id, e))
	}

	server.AddWatcher(cryptotransfer.NewCryptoTransferWatcher(hederaClient, id, configuration.Hedera.MirrorNode.PollingInterval, repository, account.MaxRetries, startTimestamp))
	log.Infof("Added a Crypto Transfer Watcher for account [%s]\n", account.Id)
	return nil
}

func addConsensusTopicWatcher(configuration *config.Config,
	hederaNodeClient *hederaClients.HederaNodeClient,
	hederaMirrorClient *hederaClients.HederaMirrorClient,
	repository *status.StatusRepository,
	server *server.HederaWatcherServer,
	startTimestamp int64,
) error {
	topic := configuration.Hedera.Watcher.ConsensusMessage.Topic
	id, e := hedera.TopicIDFromString(topic.Id)
	if e != nil {
		return errors.New(fmt.Sprintf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topic.Id, e))
	}

	server.AddWatcher(cmw.NewConsensusTopicWatcher(hederaNodeClient, hederaMirrorClient, id, repository, topic.MaxRetries, startTimestamp))
	log.Infof("Added a Consensus Topic Watcher for topic [%s]\n", topic.Id)
	return nil
}

func recoverLostProgress(configuration config.Hedera,
	transactionRepository *transaction.TransactionRepository,
	statusCryptoTransferRepository *status.StatusRepository,
	statusConsensusMessageRepository *status.StatusRepository,
	messageRepository *message.MessageRepository,
	hederaMirrorClient *hederaClients.HederaMirrorClient,
	ethClient *ethclient.EthereumClient,
	hederaNodeClient *hederaClients.HederaNodeClient,
) (int64, error) {
	log.Infof("Initializing Recovery Service for Account [%s] and Topic [%s]", configuration.Watcher.CryptoTransfer.Account.Id, configuration.Watcher.ConsensusMessage.Topic.Id)
	account, err := hedera.AccountIDFromString(configuration.Watcher.CryptoTransfer.Account.Id)
	if err != nil {
		return 0, err
	}

	topic, err := hedera.TopicIDFromString(configuration.Watcher.ConsensusMessage.Topic.Id)
	if err != nil {
		return 0, err
	}

	recoveryService := recovery.NewRecoveryService(
		transactionRepository,
		statusConsensusMessageRepository,
		statusCryptoTransferRepository,
		messageRepository,
		hederaMirrorClient,
		ethClient,
		hederaNodeClient,
		account,
		topic)

	log.Infof("Starting Recovery Process for Account [%s] and Topic [%s]", configuration.Watcher.CryptoTransfer.Account.Id, configuration.Watcher.ConsensusMessage.Topic.Id)
	now, err := recoveryService.Recover()
	if err != nil {
		return 0, err
	}

	return now, nil
}
