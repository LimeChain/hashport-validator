package main

import (
	"flag"

	"github.com/hashgraph/hedera-sdk-go/v2"
	log "github.com/sirupsen/logrus"
)

func main() {
	senderAccIdStr := flag.String("senderAccountId", "0.0.1", "Hedera Account ID")
	senderPrivateKeyStr := flag.String("senderPrivateKey", "0x0", "Hedera Private Key")
	receiverAccIdStr := flag.String("receiverAccountId", "0.0.1", "Hedera Account ID")
	amount := flag.Int64("amount", hedera.HbarFrom(10, hedera.HbarUnits.Hbar).AsTinybar(), "amount in tinybar")
	flag.Parse()

	senderAccId, err := hedera.AccountIDFromString(*senderAccIdStr)
	if err != nil {
		log.Fatalln(err)
	}
	senderPrivateKey, err := hedera.PrivateKeyFromString(*senderPrivateKeyStr)
	if err != nil {
		log.Fatalln(err)
	}
	receiverAccId, err := hedera.AccountIDFromString(*receiverAccIdStr)
	if err != nil {
		log.Fatalln(err)
	}

	client := hedera.ClientForTestnet()
	client.SetOperator(senderAccId, senderPrivateKey)

	sendTx, err := hedera.NewTransferTransaction().
		AddHbarTransfer(senderAccId, hedera.HbarFromTinybar(-*amount)).
		AddHbarTransfer(receiverAccId, hedera.HbarFromTinybar(*amount)).
		Sign(senderPrivateKey).
		Execute(client)
	if err != nil {
		log.Fatalln(err)
	}

	sendRx, err := sendTx.GetReceipt(client)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(sendRx.Status)
}
