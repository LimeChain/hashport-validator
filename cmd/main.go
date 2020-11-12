package main

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	consensusmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	initLogger()
	configuration := config.LoadConfig()
	hederaClient := hederaClient.NewHederaClient(configuration.Hedera.MirrorNode.ApiAddress, configuration.Hedera.MirrorNode.ClientAddress)

	db := persistence.RunDb(configuration.Hedera.Validator.Db)
	statusRepository := status.NewStatusRepository(db)

	server := server.NewServer()

	failure := addCryptoTransferWatchers(configuration, hederaClient, statusRepository, server)
	if failure != nil {
		log.Errorln(failure)
	}

	failure = addConsensusTopicWatchers(configuration, hederaClient, server)
	if failure != nil {
		log.Errorln(failure)
	}

	server.Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}

func addCryptoTransferWatchers(configuration *config.Config, hederaClient *hederaClient.HederaClient, repository *status.StatusRepository, server *server.HederaWatcherServer) error {
	if len(configuration.Hedera.Watcher.CryptoTransfer.Accounts) == 0 {
		log.Warningln("There are no Crypto Transfer Watchers.")
	}
	for _, account := range configuration.Hedera.Watcher.CryptoTransfer.Accounts {
		id, e := hedera.AccountIDFromString(account.Id)
		if e != nil {
			return errors.New(fmt.Sprintf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", account.Id, e))
		}

		server.AddWatcher(cryptotransfer.NewCryptoTransferWatcher(hederaClient, id, configuration.Hedera.MirrorNode.PollingInterval, repository))
		log.Infof("Added a Crypto Transfer Watcher for account [%s]\n", account.Id)
	}
	return nil
}

func addConsensusTopicWatchers(configuration *config.Config, hederaClient *hederaClient.HederaClient, server *server.HederaWatcherServer) error {
	if len(configuration.Hedera.Watcher.ConsensusMessage.Topics) == 0 {
		log.Warningln("There are no Consensus Topic Watchers.")
	}
	for _, topic := range configuration.Hedera.Watcher.ConsensusMessage.Topics {
		id, e := hedera.TopicIDFromString(topic.Id)
		if e != nil {
			return errors.New(fmt.Sprintf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topic.Id, e))
		}

		server.AddWatcher(consensusmessage.NewConsensusTopicWatcher(hederaClient, id))
		log.Infof("Added a Consensus Topic Watcher for topic [%s]\n", topic.Id)
	}
	return nil
}

func initLogger() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
