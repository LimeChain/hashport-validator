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
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
	client "github.com/limechain/hedera-eth-bridge-validator/scripts"
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

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	membersSlice := strings.Split(*memberPrKeys, " ")

	var custodianKey []hedera.PrivateKey
	for i := 0; i < len(membersSlice); i++ {
		privateKeyFromStr, err := hedera.PrivateKeyFromString(membersSlice[i])
		if err != nil {
			panic(err)
		}
		custodianKey = append(custodianKey, privateKeyFromStr)
	}

	custodialKey := hedera.KeyListWithThreshold(uint(len(membersSlice)))
	for _, m := range membersSlice {
		key, _ := hedera.PrivateKeyFromString(m)
		custodialKey.Add(key.PublicKey())
	}

	bridgeIDFromString, err := hedera.AccountIDFromString(*bridgeID)
	if err != nil {
		panic(err)
	}

	tokenId := createBridgeAccountToken(client, bridgeIDFromString, custodialKey, custodianKey)

	fmt.Println("Token ID:", tokenId)
}
func createBridgeAccountToken(client *hedera.Client, bridgeAccount hedera.AccountID, supplyKey *hedera.KeyList, custodianKey []hedera.PrivateKey) *hedera.TokenID {
	freezeTokenTX, err := hedera.NewTokenCreateTransaction().
		SetTreasuryAccountID(bridgeAccount).
		SetSupplyKey(supplyKey).
		SetTokenName("e2e-test-token").
		SetTokenSymbol("ett").
		SetInitialSupply(100000000000000).
		SetDecimals(8).
		SetMaxTransactionFee(hedera.HbarFrom(20, "hbar")).
		FreezeWith(client)

	if err != nil {
		fmt.Println(err)
	}

	// add all keys
	for i := 0; i < len(custodianKey); i++ {
		freezeTokenTX = freezeTokenTX.Sign(custodianKey[i])
	}
	createTx, err := freezeTokenTX.Execute(client)
	if err != nil {
		fmt.Println(err)
	}
	receipt, err := createTx.GetReceipt(client)
	if err != nil {
		fmt.Println(err)
	}

	return receipt.TokenID
}
