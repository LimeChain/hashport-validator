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
	clientScript "github.com/limechain/hedera-eth-bridge-validator/scripts/client"
	"io/ioutil"
	"time"
)

func main() {
	executorAccountID := flag.String("executorAccountID", "", "Hedera Account Id")
	topicID := flag.String("topicID", "", "Hedera Topic Id")
	network := flag.String("network", "", "Hedera Network Type")
	configPath := flag.String("configPath", "", "Path to the 'bridge.yaml' config file")
	nodeAccountID := flag.String("nodeAccountID", "0.0.3", "Node account id on which to process the transaction.")
	validStartMinutes := flag.Int("validStartMinutes", 2, "Valid minutes for which the transaction needs to be signed and submitted after.")
	flag.Parse()
	validatePrepareUpdateConfigParams(executorAccountID, topicID, network, configPath, validStartMinutes)

	content, topicIdParsed, executor, nodeAccount := parseParams(configPath, topicID, executorAccountID, nodeAccountID)

	client := clientScript.GetClientForNetwork(*network)
	additionTime := time.Minute * time.Duration(*validStartMinutes)
	transactionID := hedera.NewTransactionIDWithValidStart(executor, time.Now().Add(additionTime))
	frozenTx, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicIdParsed).
		SetMaxChunks(30).
		SetMessage(content).
		SetMaxChunks(60).
		SetTransactionID(transactionID).
		SetNodeAccountIDs([]hedera.AccountID{nodeAccount}).
		FreezeWith(client)
	if err != nil {
		panic(err)
	}

	bytes, err := frozenTx.ToBytes()
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(bytes))
}

func parseParams(configPath *string, topicId *string, executorId *string, nodeAccountId *string) ([]byte, hedera.TopicID, hedera.AccountID, hedera.AccountID) {
	content, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	contentLength := len(content)
	if contentLength == 0 {
		panic("config file is empty")
	}
	topicIdParsed, err := hedera.TopicIDFromString(*topicId)
	if err != nil {
		panic(err)
	}
	executor, err := hedera.AccountIDFromString(*executorId)
	if err != nil {
		panic(err)
	}
	nodeAccount, err := hedera.AccountIDFromString(*nodeAccountId)
	if err != nil {
		panic(fmt.Sprintf("Invalid Node Account Id. Err: %s", err))
	}
	return content, topicIdParsed, executor, nodeAccount
}

func validatePrepareUpdateConfigParams(executorId *string, topicId *string, network *string, configPath *string, validStartMinutes *int) {
	if *executorId == "0.0" {
		panic("Executor id was not provided")
	}
	if *topicId == "" {
		panic("topic Id not provided")
	}
	if *network == "" {
		panic("network not provided")
	}
	if *configPath == "" {
		panic("configPath not provided")
	}
}
