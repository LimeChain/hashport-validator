/*
* Copyright 2021 LimeChain Ltd.
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
	"flag"
	"fmt"

	"github.com/hashgraph/hedera-sdk-go"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountId", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	flag.Parse()

	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}
	fmt.Println("-----------Start-----------")
	client := previewClient(*privateKey, *accountID, *network)
	privKey1, err := cryptoCreate(client)
	if err != nil {
		panic(err)
	}
	privKey2, err := cryptoCreate(client)
	if err != nil {
		panic(err)
	}
	privKey3, err := cryptoCreate(client)
	if err != nil {
		panic(err)
	}
	topicKey := hedera.KeyListWithThreshold(1)
	topicKey = topicKey.
		Add(privKey1.PublicKey()).
		Add(privKey2.PublicKey()).
		Add(privKey3.PublicKey())
	txID, err := hedera.NewTopicCreateTransaction().
		SetSubmitKey(topicKey).
		Execute(client)
	if err != nil {
		panic(err)
	}
	topicReceipt, err := txID.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("TopicID: %v\n", topicReceipt.TopicID)
	custodialKey := hedera.KeyListWithThreshold(3)
	custodialKey = custodialKey.Add(privKey1.PublicKey()).Add(privKey2.PublicKey()).Add(privKey3.PublicKey())
	// Creating Bridge theshhold account
	bridgeAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		SetInitialBalance(hedera.NewHbar(100)).
		Execute(client)
	if err != nil {
		panic(err)
	}
	bridgeAccountReceipt, err := bridgeAccount.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bridge Account: %v\n", bridgeAccountReceipt.AccountID)

	// Creating Scheduled transaction payer theshhold account
	scheduledTxPayerAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		SetInitialBalance(hedera.NewHbar(100)).
		Execute(client)
	if err != nil {
		panic(err)
	}
	scheduledTxPayerAccountReceipt, err := scheduledTxPayerAccount.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Scheduled Tx Payer Account: %v\n", scheduledTxPayerAccountReceipt.AccountID)
	fmt.Println("---Executed Successfully---")
}
func cryptoCreate(client *hedera.Client) (hedera.PrivateKey, error) {
	privateKey, _ := hedera.GeneratePrivateKey()
	fmt.Printf("Hedera Private Key: %v\n", privateKey.String())
	publicKey := privateKey.PublicKey()
	newAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(publicKey).
		SetInitialBalance(hedera.NewHbar(100)).
		Execute(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	receipt, err := newAccount.GetReceipt(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	fmt.Printf("AccountID: %v\n", receipt.AccountID)

	balance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(*receipt.AccountID).
		Execute(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	fmt.Printf("Balance = %v\n", balance.Hbars.String())
	fmt.Println("--------------------------")
	return privateKey, nil
}
func previewClient(privateKey, accountID, network string) *hedera.Client {
	var client *hedera.Client

	if network == "previewnet" {
		client = hedera.ClientForPreviewnet()
	} else if network == "testnet" {
		client = hedera.ClientForTestnet()
	} else if network == "mainnet" {
		client = hedera.ClientForMainnet()
	} else {
		panic("Unknown Network Type!")
	}
	accID, err := hedera.AccountIDFromString(accountID)
	if err != nil {
		panic(err)
	}
	pK, err := hedera.PrivateKeyFromString(privateKey)
	if err != nil {
		panic(err)
	}
	client.SetOperator(accID, pK)
	return client
}
