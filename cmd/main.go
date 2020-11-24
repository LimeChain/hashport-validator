package main

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClients "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	cth "github.com/limechain/hedera-eth-bridge-validator/app/process/handler/crypto-transfer"
	consensusmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	initLogger()
	configuration := config.LoadConfig()
	db := persistence.RunDb(configuration.Hedera.Validator.Db)
	hederaMirrorClient := hederaClients.NewHederaMirrorClient(configuration.Hedera.MirrorNode.ApiAddress)
	hederaNodeClient := hederaClients.NewNodeClient(configuration.Hedera.Client)
	ethSigner := eth.NewEthSigner(configuration.Hedera.Client.Operator.EthPrivateKey)

	transactionRepository := transaction.NewTransactionRepository(db)
	server := server.NewServer()

	server.AddHandler("HCS_CRYPTO_TRANSFER", cth.NewCryptoTransferHandler(
		configuration.Hedera.Handler.CryptoTransfer,
		ethSigner,
		hederaMirrorClient,
		hederaNodeClient,
		transactionRepository))

	statusCryptoTransferRepository := status.NewStatusRepository(db, "CRYPTO_TRANSFER")
	statusConsensusMessageRepository := status.NewStatusRepository(db, "HCS_TOPIC")

	err := addCryptoTransferWatchers(configuration, hederaMirrorClient, statusCryptoTransferRepository, server)
	if err != nil {
		log.Fatal(err)
	}

	err = addConsensusTopicWatchers(configuration, hederaNodeClient, hederaMirrorClient, statusConsensusMessageRepository, server)
	if err != nil {
		log.Fatal(err)
	}

	server.Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
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

		server.AddWatcher(consensusmessage.NewConsensusTopicWatcher(hederaNodeClient, hederaMirrorClient, id, repository, topic.MaxRetries, topic.StartTimestamp))
		log.Infof("Added a Consensus Topic Watcher for topic [%s]\n", topic.Id)
	}
	return nil
}

func initLogger() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
