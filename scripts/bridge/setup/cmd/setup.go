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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/bridge/setup"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	members := flag.Int("members", 1, "The count of the members")
	adminKey := flag.String("adminKey", "", "The admin key")
	topicThreshold := flag.Uint("topicThreshold", 1, "Topic member keys sign threshold")
	flag.Parse()

	// pass empty array of private keys since we are using this script for new accounts
	hederaPrivateKeys := make([]hedera.PrivateKey, 0)
	result := setup.Deploy(privateKey, accountID, adminKey, network, members, hederaPrivateKeys, *topicThreshold)
	if result.Error != nil {
		panic(result.Error)
	}
}
