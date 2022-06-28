package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func main() {
	executorAccountID := flag.String("executorAccountID", "0.0", "Hedera Executor Account ID")
	bridgeAccountID := flag.String("bridgeAccountID", "0.0", "Hedera Treasury account ID")
	tokenID := flag.String("tokenID", "0.0", "Token ID")

	flag.Parse()

	if *executorAccountID == "0.0" {
		panic("executor account id was not provided")
	}

	if *bridgeAccountID == "0.0" {
		panic("bridge account id was not provided")
	}
	if *tokenID == "0.0" {
		panic("token id not provided")
	}

	executor, err := hedera.AccountIDFromString(*executorAccountID)
	if err != nil {
		panic(err)
	}
	bridgeAccount, err := hedera.AccountIDFromString(*bridgeAccountID)
	if err != nil {
		panic(err)
	}
	token, err := hedera.TokenIDFromString(*tokenID)
	if err != nil {
		panic(err)
	}

	client := hedera.ClientForTestnet() // Testnet

	additionTime := time.Minute * 2 // 2 minutes

	transactionID := hedera.NewTransactionIDWithValidStart(executor, time.Now().Add(additionTime))

	frozen, err := hedera.NewTokenAssociateTransaction().
		SetTransactionID(transactionID).
		SetAccountID(bridgeAccount).
		SetTokenIDs(token).
		FreezeWith(client)

	if err != nil {
		panic(err)
	}
	bytes, err := frozen.ToBytes()
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(bytes))
}
