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
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
	client "github.com/limechain/hedera-eth-bridge-validator/scripts"
)

func main() {
	membersKeys := flag.String("privateKeys", "0x0", "Members Hedera Private Keys separated by comma")
	network := flag.String("network", "", "Hedera Network Type")
	executorAccountId := flag.String("executorAccountID", "", "Executor's Hedera Account Id")
	executorPK := flag.String("executorPrivateKey", "", "Executor's Hedera Private Key")

	flag.Parse()
	if *membersKeys == "0x0" {
		panic("Members Private Keys weren't provided")
	}
	if *executorAccountId == "" {
		panic("Executor Account Id wasn't provided")
	}
	if *executorPK == "" {
		panic("Executor Private Key wasn't provided")
	}

	membersKeysParsed := strings.Split(*membersKeys, ",")

	fmt.Println("-----------Start-----------")
	client := client.Init(*executorPK, *executorAccountId, *network)
	executorPrivateKey, err := hedera.PrivateKeyFromString(*executorPK)
	if err != nil {
		panic(err)
	}
	var memberKeys []hedera.PrivateKey
	for _, key := range membersKeysParsed {
		pK, err := hedera.PrivateKeyFromString(key)
		if err != nil {
			panic(err)
		}
		memberKeys = append(memberKeys, pK)
	}

	fmt.Println("Private keys array:", memberKeys)

	topicKey := hedera.KeyListWithThreshold(uint(len(memberKeys)))
	for i := 0; i < len(membersKeysParsed); i++ {
		topicKey.Add(memberKeys[i].PublicKey())
	}

	adminPublicKey := executorPrivateKey.PublicKey()
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
