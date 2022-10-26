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
	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/token/associate"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	tokenID := flag.String("tokenID", "0.0", "Bridge account ID")
	flag.Parse()
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}
	if *tokenID == "0.0" {
		panic("Token id was not provided")
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	tokenIDFromString, err := hedera.TokenIDFromString(*tokenID)
	if err != nil {
		panic(err)
	}
	receipt, err := associate.TokenToAccount(client, tokenIDFromString, client.GetOperatorAccountID())
	if err != nil {
		panic(err)
	}
	fmt.Println("Associate transaction status:", receipt.Status)
}
