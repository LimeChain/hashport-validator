/*
 * Copyright 2022 LimeChain Ltd.
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
	"encoding/hex"
	"flag"
	"fmt"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"
)

type Transactioner interface {
	GetTransactionID() hedera.TransactionID
}

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	transaction := flag.String("transaction", "", "Hedera to-be-submitted Transaction")
	flag.Parse()
	validateParams(transaction, privateKey, accountID)

	client := client.Init(*privateKey, *accountID, *network)
	decoded, err := hex.DecodeString(*transaction)
	if err != nil {
		panic(err)
	}

	deserialized, err := hedera.TransactionFromBytes(decoded)
	if err != nil {
		panic(fmt.Sprintf("failed to parse transaction. err [%s]", err))
	}

	var transactionResponse hedera.TransactionResponse
	switch tx := deserialized.(type) {
	case hedera.TransferTransaction:
		waitForTransactionStart(&tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TopicUpdateTransaction:
		waitForTransactionStart(&tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TokenUpdateTransaction:
		waitForTransactionStart(&tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.AccountUpdateTransaction:
		waitForTransactionStart(&tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TokenCreateTransaction:
		waitForTransactionStart(&tx)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TokenMintTransaction:
		waitForTransactionStart(&tx)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TokenAssociateTransaction:
		waitForTransactionStart(&tx)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TopicMessageSubmitTransaction:
		waitForTransactionStart(&tx)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
	case hedera.TokenBurnTransaction:
		waitForTransactionStart(&tx)
		transactionResponse, err = tx.Execute(client)
	default:
		panic("invalid tx type provided")
	}

	if err != nil {
		panic(err)
	}
	fmt.Printf("TransactionID: [%s]\n", transactionResponse.TransactionID)

	receipt, err := transactionResponse.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Println(receipt)
}

func validateParams(transaction *string, privateKey *string, accountID *string) {
	if *transaction == "" {
		panic("transaction has not been provided")
	}
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}
}

func waitForTransactionStart(tx Transactioner) {
	validStart := tx.GetTransactionID().ValidStart
	waitTime := time.Until(*validStart)
	fmt.Printf("Transaction will be excuted after: %v\n", waitTime)
	time.Sleep(waitTime)
}
