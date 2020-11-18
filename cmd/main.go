package main

import (
	"fmt"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	cth "github.com/limechain/hedera-eth-bridge-validator/app/process/handlers/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/server"
)

func main() {
	configuration := config.LoadConfig()
	db := persistence.RunDb(configuration.Hedera.Validator.Db)
	hederaClient := hederaClient.NewClient(configuration.Hedera.Client)
	ethSigner := eth.NewEthSigner(configuration.Hedera.Client.Operator.PrivateKey)

	transactionRepository := transaction.NewTransactionRepository(db)
	server := server.NewServer()

	server.AddHandler("HCS_CRYPTO_TRANSFER",
		cth.NewCryptoTransferHandler(configuration.Hedera.Handlers.CryptoTransferHandler, ethSigner, hederaClient, transactionRepository))

	server.Run(fmt.Sprintf(":%s", configuration.Hedera.Validator.Port))
}
