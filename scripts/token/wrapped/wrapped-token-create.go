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
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	// The bridge account, which will be added as treasury to the new account
	bridgeID := flag.String("bridgeID", "0.0", "Bridge account ID")
	// The admin key
	adminKey := flag.String("adminKey", "", "Admin Key")
	// The desired threshold of n/m keys required for supply key
	threshold := flag.Uint("threshold", 1, "Threshold Key")
	// A list of keys that will be added as supply key to the newly created token, with a threshold of the provided
	supplyKeys := flag.String("supplyKeys", "", "Supply keys")
	// Keys that are key to the bridge ID, which need to sign the transaction
	memberPrKeys := flag.String("memberPrKeys", "", "The count of the members")
	// Generate supplyKeys from members privateKeys
	generateSupplyKeysFromMemberPrKeys := flag.Bool("generateSupplyKeysFromMemberPrKeys", false, "Flag to generate the supplyKeys (public keys) from members private keys.")

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
	if *adminKey == "" {
		panic("admin key not provided")
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	membersSlice := strings.Split(*memberPrKeys, ",")

	adminPublicKey, err := hedera.PublicKeyFromString(*adminKey)
	if err != nil {
		panic(err)
	}

	var custodianKey []hedera.PrivateKey
	for i := 0; i < len(membersSlice); i++ {
		privateKeyFromStr, err := hedera.PrivateKeyFromString(membersSlice[i])
		if err != nil {
			panic(err)
		}
		custodianKey = append(custodianKey, privateKeyFromStr)
	}

	supplyKey := hedera.KeyListWithThreshold(*threshold)
	if generateSupplyKeysFromMemberPrKeys != nil && *generateSupplyKeysFromMemberPrKeys == true {
		for _, prKey := range custodianKey {
			key := prKey.PublicKey()
			supplyKey.Add(key)
		}
	} else {
		supplyKeysSlice := strings.Split(*supplyKeys, ",")

		for _, sk := range supplyKeysSlice {
			key, err := hedera.PublicKeyFromString(sk)
			if err != nil {
				panic(fmt.Sprintf("failed to parse supply key [%s]. error [%s]", sk, err))
			}
			supplyKey.Add(key)
		}
	}

	bridgeIDFromString, err := hedera.AccountIDFromString(*bridgeID)
	if err != nil {
		panic(err)
	}

	tokenId := createBridgeAccountToken(client, bridgeIDFromString, adminPublicKey, supplyKey, custodianKey)

	fmt.Println("Token ID:", tokenId)
}

func createBridgeAccountToken(client *hedera.Client, bridgeAccount hedera.AccountID, adminKey hedera.PublicKey, supplyKey *hedera.KeyList, custodianKey []hedera.PrivateKey) *hedera.TokenID {
	freezeTokenTX, err := hedera.NewTokenCreateTransaction().
		SetTreasuryAccountID(bridgeAccount).
		SetAdminKey(adminKey).
		SetSupplyKey(supplyKey).
		SetTokenName("e2e-test-token").
		SetTokenSymbol("ett").
		SetInitialSupply(100000000000000).
		SetDecimals(8).
		FreezeWith(client)

	if err != nil {
		panic(err)
	}

	// add all keys
	for i := 0; i < len(custodianKey); i++ {
		freezeTokenTX = freezeTokenTX.Sign(custodianKey[i])
	}
	createTx, err := freezeTokenTX.Execute(client)
	if err != nil {
		panic(err)
	}
	receipt, err := createTx.GetReceipt(client)
	if err != nil {
		panic(err)
	}

	return receipt.TokenID
}
