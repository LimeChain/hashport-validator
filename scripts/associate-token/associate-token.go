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
	"strconv"
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountId", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	bridgeID := flag.String("bridgeID", "0.0", "Bridge account ID")
	tokenID := flag.String("tokenID", "0.0", "Bridge account ID")
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
	if *tokenID == "0.0" {
		panic("Token id was not provided")
	}

	fmt.Println("-----------Start-----------")
	client := initClient(*privateKey, *accountID, *network)

	bridgeIDFromString, err := hedera.AccountIDFromString(*bridgeID)
	if err != nil {
		panic(err)
	}
	tokenIDFromString := TokenIDFromString(*tokenID)

	receipt := associateTokenToAccount(client, tokenIDFromString, bridgeIDFromString)
	fmt.Println("Associate transaction status:", receipt.Status)
}
func associateTokenToAccount(client *hedera.Client, token hedera.TokenID, bridgeID hedera.AccountID) hedera.TransactionReceipt {
	res, err := hedera.
		NewTokenAssociateTransaction().
		SetAccountID(client.GetOperatorAccountID()).
		SetTokenIDs(token).
		Execute(client)
	if err != nil {
		fmt.Println(err)
	}

	receipt, err := res.GetReceipt(client)
	if err != nil {
		fmt.Println(err)
	}
	return receipt
}
func initClient(privateKey, accountID, network string) *hedera.Client {
	var client *hedera.Client

	if network == "previewnet" {
		client = hedera.ClientForPreviewnet()
	} else if network == "testnet" {
		client = hedera.ClientForTestnet()
	} else if network == "mainnet" {
		client = hedera.ClientForMainnet()
	} else {
		panic("Unknown Network Type!")
	}
	accID, err := hedera.AccountIDFromString(accountID)
	if err != nil {
		panic(err)
	}
	pK, err := hedera.PrivateKeyFromString(privateKey)
	if err != nil {
		panic(err)
	}
	client.SetOperator(accID, pK)
	return client
}
func TokenIDFromString(tokenId string) hedera.TokenID {
	args := strings.Split(tokenId, ".")
	shard, _ := strconv.ParseUint(args[0], 10, 64)
	realm, _ := strconv.ParseUint(args[1], 10, 64)
	token, _ := strconv.ParseUint(args[2], 10, 64)
	return hedera.TokenID{
		Shard: shard,
		Realm: realm,
		Token: token,
	}
}
