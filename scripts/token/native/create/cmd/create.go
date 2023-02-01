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
	"github.com/limechain/hedera-eth-bridge-validator/scripts/token/native/create"
	"strings"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	bridgeID := flag.String("bridgeID", "0.0", "Bridge account ID")
	memberPrKeys := flag.String("memberPrKeys", "", "The count of the members")
	flag.Parse()
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}
	if *bridgeID == "0.0" {
		panic("Bridge id was not provided")
	}

	tokenName := flag.String("name", "Hedera Native Generic Token", "token name")
	tokenSymbol := flag.String("symbol", "HNT", "token symbol")
	decimals := flag.Uint("decimals", 8, "decimals")
	setSupplyKey := flag.Bool("setSupplyKey", true, "Sets supply key to be the deployer")

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	if *network != "testnet" && *setSupplyKey {
		var confirmation string
		fmt.Printf("Network is set to [%s] and setSupplyKey is set to [%v]. Are you sure you what to proceed?\n", *network, *setSupplyKey)
		fmt.Println("Y/N:")
		fmt.Scanln(&confirmation)
		if confirmation != "Y" {
			panic("Exiting")
		}
	}

	membersSlice := strings.Split(*memberPrKeys, ",")

	var custodianKey []hedera.PrivateKey
	for i := 0; i < len(membersSlice); i++ {
		privateKeyFromStr, err := hedera.PrivateKeyFromString(membersSlice[i])
		if err != nil {
			panic(err)
		}
		custodianKey = append(custodianKey, privateKeyFromStr)
	}

	tokenId, err := create.CreateNativeFungibleToken(
		client,
		client.GetOperatorAccountID(),
		*tokenName,
		*tokenSymbol,
		*decimals,
		100000000000000,
		hedera.HbarFrom(20, "hbar"),
		*setSupplyKey,
	)
	if err != nil {
		panic(err)
	}

	bridgeIDFromString, err := hedera.AccountIDFromString(*bridgeID)
	if err != nil {
		panic(err)
	}

	receipt, err := associate.TokenToAccountWithCustodianKey(client, *tokenId, bridgeIDFromString, custodianKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("Token ID:", tokenId)
	fmt.Println("Associate transaction status:", receipt.Status)
}
