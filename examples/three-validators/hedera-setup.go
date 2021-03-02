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
	"fmt"

	"github.com/hashgraph/hedera-sdk-go"
)

func main() {
	client := previewClient()
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
	txId, err := hedera.NewTopicCreateTransaction().
		SetSubmitKey(topicKey).
		Execute(client)
	if err != nil {
		panic(err)
	}
	topicReceipt, err := txId.GetReceipt(client)
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
func previewClient() *hedera.Client {
	client := hedera.ClientForPreviewnet()
	// Set your account ID for PreviewNet
	accId, _ := hedera.AccountIDFromString("0.0.6526")
	// Set your Private Key for PreviewNet
	pK, _ := hedera.PrivateKeyFromString("302e020100300506032b6570042204203425f0b39dc6d80345b08b65fb4ec6296871ecb6244e9dcc258c482106ca4414")
	client.SetOperator(accId, pK)
	return client
}
