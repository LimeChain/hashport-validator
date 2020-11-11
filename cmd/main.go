package main

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
)

func main() {
	configuration := config.LoadConfig()
	persistence.RunDb(configuration.Hedera.Validator.Db)
	server.NewServer().Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}
