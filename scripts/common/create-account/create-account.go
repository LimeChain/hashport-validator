/*
 * Copyright 2024 LimeChain Ltd.
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
	"log"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Creator Private Key")
	senderAccountId := flag.String("senderAccountId", "0.0", "Creator Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	hbarAmount := flag.Float64("hbarAmount", 0, "Amount of HBAR to transfer on creation")

	flag.Parse()
	if *privateKey == "0x0" {
		panic("Hedera Topic Member's Supply Public Keys weren't provided")
	}
	if *senderAccountId == "0.0" {
		panic("Executor Account Id wasn't provided")
	}

	if *network != "mainnet" && *network != "testnet" {
		panic("Invalid network type")
	}

	fmt.Println("-----------Creating Account-----------")
	// Initialize the client
	// Parse the private key
	operatorPrivateKey, err := hedera.PrivateKeyFromString(*privateKey)
	if err != nil {
		log.Fatalf("Error parsing private key: %v", err)
	}

	operatorAccountID, err := hedera.AccountIDFromString(*senderAccountId)
	if err != nil {
		log.Fatalf("Error parsing account ID: %v", err)
	}

	client := hedera.ClientForTestnet()
	if *network == "mainnet" {
		client = hedera.ClientForMainnet()
	}
	client.SetOperator(operatorAccountID, operatorPrivateKey)

	// Generate a new key pair for the new account
	newPrivateKey, err := hedera.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("Error generating new private key: %v", err)
	}
	newPublicKey := newPrivateKey.PublicKey()

	// Create the new account transaction
	transactionResponse, err := hedera.NewAccountCreateTransaction().
		SetKey(newPrivateKey.PublicKey()).
		SetMaxAutomaticTokenAssociations(100).
		SetInitialBalance(hedera.HbarFrom(*hbarAmount, hedera.HbarUnits.Hbar)).
		Execute(client)

	if err != nil {
		log.Fatalf("Error executing account create transaction: %v", err)
	}

	// Get the receipt of the transaction
	receipt, err := transactionResponse.GetReceipt(client)
	if err != nil {
		log.Fatalf("Error getting transaction receipt: %v", err)
	}

	// Get the new account ID
	newAccountID := receipt.AccountID

	fmt.Println("-----------Account Created-----------")
	fmt.Printf("Transaction ID: %v\n", transactionResponse.TransactionID)
	fmt.Printf("New Account ID: %v\n", newAccountID)
	fmt.Printf("PRIVATE KEY: %v\n", newPrivateKey)
	fmt.Printf("PUBLIC KEY: %v\n", newPublicKey)

}
