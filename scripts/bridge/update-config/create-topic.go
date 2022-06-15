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

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	members := flag.Int("members", 1, "The count of the members")

	flag.Parse()
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)
	pK, err := hedera.PrivateKeyFromString(*privateKey)
	if err != nil {
		panic(err)
	}
	var memberKeys []hedera.PrivateKey
	memberKeys = append(memberKeys, pK)

	fmt.Println("Private keys array:", memberKeys)

	topicKey := hedera.KeyListWithThreshold(uint(*members))
	for i := 0; i < *members; i++ {
		topicKey.Add(memberKeys[i].PublicKey())
	}

	adminPublicKey := pK.PublicKey()
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
}
