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

	prKey := flag.String("prKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountId", "0.0", "Hedera Account ID")
	flag.Parse()

	client := previewClient(*prKey, *accountID)
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
	fmt.Println("TopicID: ", topicReceipt.TopicID)
	custodialKey := hedera.KeyListWithThreshold(3)
	custodialKey = custodialKey.Add(privKey1.PublicKey()).Add(privKey2.PublicKey()).Add(privKey3.PublicKey())
	custodialAcc, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		SetInitialBalance(hedera.NewHbar(100)).
		Execute(client)
	if err != nil {
		panic(err)
	}
	custodialAccReceipt, err := custodialAcc.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Println("Custodial Account: ", custodialAccReceipt.AccountID)
}
func cryptoCreate(client *hedera.Client) (hedera.PrivateKey, error) {
	privateKey, _ := hedera.GeneratePrivateKey()
	fmt.Println("Hedera Private Key: ", privateKey.String())
	publicKey := privateKey.PublicKey()
	newAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(publicKey).
		SetInitialBalance(hedera.NewHbar(100)).
		Execute(client)
	if err != nil {
		panic(err)
	}
	receipt, err := newAccount.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Println("AccountID: ", receipt.AccountID)
	fmt.Println("--------------------->")
	return privateKey, nil
}
func previewClient(prKey, accountID string) *hedera.Client {
	client := hedera.ClientForPreviewnet()
	// Set your account ID for PreviewNet
	accID, _ := hedera.AccountIDFromString(accountID)
	// Set your Private Key for PreviewNet
	pK, _ := hedera.PrivateKeyFromString(prKey)
	client.SetOperator(accID, pK)
	return client
}
