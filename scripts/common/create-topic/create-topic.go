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
	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func main() {
	supplyKeys := flag.String("supplyKeys", "0x0", "Hedera Supply Public Keys, separated by comma")
	network := flag.String("network", "", "Hedera Network Type")
	executorAccountId := flag.String("executorAccountID", "", "Executor's Hedera Account Id")
	executorPK := flag.String("executorPrivateKey", "", "Executor's Hedera Private Key")
	keyThreshold := flag.Uint("keyThreshold", 1, "Topic's keys threshold.")

	flag.Parse()
	if *supplyKeys == "0x0" {
		panic("Hedera Topic Member's Supply Public Keys weren't provided")
	}
	if *executorAccountId == "" {
		panic("Executor Account Id wasn't provided")
	}
	if *executorPK == "" {
		panic("Executor Private Key wasn't provided")
	}

	supplyKeysParsed := strings.Split(*supplyKeys, ",")

	fmt.Println("-----------Start-----------")
	hederaClient := client.Init(*executorPK, *executorAccountId, *network)
	executorPrivateKey, err := hedera.PrivateKeyFromString(*executorPK)
	if err != nil {
		panic(err)
	}

	topicKey := hedera.KeyListWithThreshold(*keyThreshold)
	for _, key := range supplyKeysParsed {
		pK, err := hedera.PublicKeyFromString(key)
		if err != nil {
			panic(err)
		}
		topicKey.Add(pK)
	}

	adminPublicKey := executorPrivateKey.PublicKey()
	txID, err := hedera.NewTopicCreateTransaction().
		SetAdminKey(adminPublicKey).
		SetSubmitKey(topicKey).
		Execute(hederaClient)
	if err != nil {
		panic(err)
	}

	topicReceipt, err := txID.GetReceipt(hederaClient)
	if err != nil {
		panic(err)
	}

	fmt.Printf("TopicID: %v\n", topicReceipt.TopicID)
	fmt.Println("--------------------------")
}
