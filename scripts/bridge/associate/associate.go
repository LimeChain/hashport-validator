package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func main() {
	executorAccountID := flag.String("executorAccountID", "0.0.30779785", "Hedera Executor Account ID")
	bridgeAccountID := flag.String("bridgeAccountID", "0.0.15678172", "Bridge account ID")
	tokenID := flag.String("tokenID", "0.0.34957240", "Token ID")
	network := flag.String("network", "testnet", "Hedera network")
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

	client, err := hedera.ClientForName(*network) // Mainnet
	if err != nil {
		panic(err)
	}

	additionTime := time.Minute * 2 // 1 minutes

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
