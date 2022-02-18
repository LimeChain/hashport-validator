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
	"flag"
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
	client "github.com/limechain/hedera-eth-bridge-validator/scripts"
)

var balance = hedera.NewHbar(100)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	members := flag.Int("members", 1, "The count of the members")
	adminKey := flag.String("adminKey", "", "The admin key")
	flag.Parse()
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}
	if *adminKey == "" {
		panic("admin key not provided")
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	var memberKeys []hedera.PrivateKey
	for i := 0; i < *members; i++ {
		privKey, err := cryptoCreate(client)
		if err != nil {
			panic(err)
		}
		memberKeys = append(memberKeys, privKey)
	}

	fmt.Println("Private keys array:", memberKeys)

	topicKey := hedera.KeyListWithThreshold(1)
	for i := 0; i < *members; i++ {
		topicKey.Add(memberKeys[i].PublicKey())
	}
	adminPublicKey, err := hedera.PublicKeyFromString(*adminKey)
	if err != nil {
		panic(err)
	}

	txID, err := hedera.NewTopicCreateTransaction().
		SetAdminKey(adminPublicKey).
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
	fmt.Println("--------------------------")

	custodialKey := hedera.KeyListWithThreshold(uint(*members))
	for i := 0; i < *members; i++ {
		custodialKey.Add(memberKeys[i].PublicKey())
	}

	// Creating Bridge threshold account
	bridgeAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		Execute(client)
	if err != nil {
		panic(err)
	}

	bridgeAccountReceipt, err := bridgeAccount.GetReceipt(client)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bridge Account: %v\n", bridgeAccountReceipt.AccountID)
	fmt.Println("--------------------------")

	// Creating Scheduled transaction payer theshhold account
	scheduledTxPayerAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		SetInitialBalance(balance).
		Execute(client)
	if err != nil {
		panic(err)
	}
	scheduledTxPayerAccountReceipt, err := scheduledTxPayerAccount.GetReceipt(client)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Scheduled Tx Payer Account: %v\n", scheduledTxPayerAccountReceipt.AccountID)
	fmt.Printf("Balance: %v\n HBars", balance)
	fmt.Println("---Executed Successfully---")
}

func cryptoCreate(client *hedera.Client) (hedera.PrivateKey, error) {
	privateKey, _ := hedera.GeneratePrivateKey()
	fmt.Printf("Hedera Private Key: %v\n", privateKey.String())
	fmt.Printf("Hederea Public Key: %v\n", privateKey.PublicKey().String())
	publicKey := privateKey.PublicKey()
	newAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(publicKey).
		SetInitialBalance(balance).
		Execute(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	receipt, err := newAccount.GetReceipt(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	fmt.Printf("AccountID: %v\n", receipt.AccountID)
	fmt.Printf("Balance: %v\n HBars", balance)
	fmt.Println("--------------------------")
	return privateKey, nil
}
