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
	"github.com/hashgraph/hedera-sdk-go/v2"
	client "github.com/limechain/hedera-eth-bridge-validator/scripts"
)

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
	switch deserialized.(type) {
	case hedera.TransferTransaction:
		tx := deserialized.(hedera.TransferTransaction)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.TopicUpdateTransaction:
		tx := deserialized.(hedera.TopicUpdateTransaction)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.TokenUpdateTransaction:
		tx := deserialized.(hedera.TokenUpdateTransaction)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.AccountUpdateTransaction:
		tx := deserialized.(hedera.AccountUpdateTransaction)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.TokenCreateTransaction:
		tx := deserialized.(hedera.TokenCreateTransaction)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.TokenMintTransaction:
		tx := deserialized.(hedera.TokenMintTransaction)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.TokenAssociateTransaction:
		tx := deserialized.(hedera.TokenAssociateTransaction)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
		break
	case hedera.TopicMessageSubmitTransaction:
		tx := deserialized.(hedera.TopicMessageSubmitTransaction)
		fmt.Println(tx)
		transactionResponse, err = tx.Execute(client)
		break
	default:
		panic("invalid tx type provided")
	}

	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("TransactionID: [%s]", transactionResponse.TransactionID))

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
