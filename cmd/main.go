package main

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
	"strconv"
)

func main() {
	config := config.LoadConfig()
	persistence.RunDb(config.Hedera.Validator.Db)
	port, _ := strconv.ParseInt(config.Hedera.Validator.Port, 10, 64)
	server.NewServer().Run(int(port))
}
