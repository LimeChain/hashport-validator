package main

import (
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	http "github.com/limechain/hedera-eth-bridge-validator/app/clients/http"
	consensus_message "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	crypto_transfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
)

func main() {
	configuration := config.LoadConfig()
	//persistence.RunDb(configuration.Hedera.Validator.Db)
	server := server.NewServer()

	mirrorNodeClient, _ := hedera.NewMirrorClient(configuration.Hedera.MirrorNode.Client)
	httpClient := http.NewClient()

	for _, account := range configuration.Hedera.Watcher.CryptoTransfer.Accounts {
		id, e := hedera.AccountIDFromString(account.Id)
		if e != nil {
			panic(e)
		}

		server.AddWatcher(crypto_transfer.NewCryptoTransferWatcher(httpClient, id))
	}
	for _, topic := range configuration.Hedera.Watcher.ConsensusMessage.Topics {
		id, e := hedera.TopicIDFromString(topic.Id)
		if e != nil {
			panic(e)
		}

		server.AddWatcher(consensus_message.NewConsensusTopicWatcher(mirrorNodeClient, id))
	}
	server.Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}
