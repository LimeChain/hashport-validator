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
	"log"
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	mirrorNodeClient "github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"
)

type AccountBalance struct {
	Balance uint64         `json:"balance"`
	Tokens  []TokenBalance `json:"tokens"`
}

type TokenBalance struct {
	TokenID string `json:"token_id"`
	Balance uint64 `json:"balance"`
}

const (
	HederaMainnetNetworkId = 295
	HederaTestnetNetworkId = 296
)

func getAccountBalanceFromMirrorNode(mirrorClient mirrorNodeClient.MirrorNode, accountID string) (map[string]int, error) {
	account, err := mirrorClient.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting account balance from mirror node: %w", err)
	}
	hederaTokenBalancesByAddress := make(map[string]int)
	for _, token := range account.Balance.Tokens {
		hederaTokenBalancesByAddress[token.TokenID] = token.Balance
	}

	return hederaTokenBalancesByAddress, nil
}

func transferTokens(client *hedera.Client, mirrorClient mirrorNodeClient.MirrorNode, senderAccountID hedera.AccountID, recipientAccountId hedera.AccountID, tokenIds []hedera.TokenID, hbarAmount *hedera.Hbar) (*hedera.TransactionReceipt, error) {
	accountBalance, err := getAccountBalanceFromMirrorNode(mirrorClient, senderAccountID.String())
	if err != nil {
		return nil, fmt.Errorf("error retrieving account balance: %w", err)
	}

	fmt.Println("Preparing transaction")

	transferTransaction := hedera.NewTransferTransaction()

	if hbarAmount != nil {
		transferTransaction.
			AddHbarTransfer(senderAccountID, hbarAmount.Negated()).
			AddHbarTransfer(recipientAccountId, *hbarAmount)
	}

	for _, tokenID := range tokenIds {
		if balance, ok := accountBalance[tokenID.String()]; ok {
			transferTransaction.AddTokenTransfer(tokenID, senderAccountID, -int64(balance)).
				AddTokenTransfer(tokenID, recipientAccountId, int64(balance))
		} else {
			log.Printf("No balance found for token ID: %v", tokenID)
		}
	}

	txResponse, err := transferTransaction.Execute(client)

	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt: %w", err)
	}

	receipt, err := txResponse.GetReceipt(client)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt: %w", err)
	}

	if receipt.Status != hedera.StatusSuccess {
		return nil, fmt.Errorf("transaction failed with status: %v", receipt.Status)
	}

	return &receipt, nil
}

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	senderAccountId := flag.String("senderAccountId", "0.0", "Sender Account ID")
	recipientAccountId := flag.String("recipientAccountId", "", "Recipient account ID")
	network := flag.String("network", "", "Hedera Network Type")
	tokenIDs := flag.String("tokenIds", "", "Comma-separated list of token IDs")
	hbarAmount := flag.Float64("hbarAmount", 0, "Amount of HBAR to transfer (optional)")

	flag.Parse()
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *senderAccountId == "0.0" {
		panic("Account id was not provided")
	}
	if *tokenIDs == "" {
		panic("Token ids was not provided")
	}
	if *recipientAccountId == "" {
		panic("Recipient id was not provided")
	}
	if *network == "" {
		panic("Network was not provided")
	}

	parsedSenderAccountId, err := hedera.AccountIDFromString(*senderAccountId)
	if err != nil {
		log.Fatalf("Error parsing sender account ID: %v", err)
	}

	parsedRecipientAccountId, err := hedera.AccountIDFromString(*recipientAccountId)
	if err != nil {
		log.Fatalf("Error parsing recipient account ID: %v", err)
	}

	client := client.Init(*privateKey, *senderAccountId, *network)

	tokenIDStrings := strings.Split(*tokenIDs, ",")
	var hederaTokenIDs []hedera.TokenID
	for _, tokenIDStr := range tokenIDStrings {
		tokenID, err := hedera.TokenIDFromString(tokenIDStr)
		if err != nil {
			log.Fatalf("Error parsing token ID '%s': %v", tokenIDStr, err)
		}
		hederaTokenIDs = append(hederaTokenIDs, tokenID)
	}

	var hbarAmountHedera *hedera.Hbar
	if *hbarAmount > 0 {
		hbarValue := hedera.HbarFrom(*hbarAmount, hedera.HbarUnits.Hbar)
		hbarAmountHedera = &hbarValue
	}

	var hederaNetworkId uint64
	if *network != "testnet" {
		var confirmation string
		fmt.Printf("Network is set to [%s]. Are you sure you what to proceed?\n", *network)
		fmt.Println("Y/N:")
		fmt.Scanln(&confirmation)
		if confirmation != "Y" {
			panic("Exiting")
		}
		hederaNetworkId = HederaMainnetNetworkId
	} else {
		hederaNetworkId = HederaTestnetNetworkId
	}

	mirrorNodeConfigByNetwork := map[uint64]config.MirrorNode{
		HederaMainnetNetworkId: {
			ClientAddress: "mainnet-public.mirrornode.hedera.com/:443",
			ApiAddress:    "https://mainnet-public.mirrornode.hedera.com/api/v1/",
		},
		HederaTestnetNetworkId: {
			ClientAddress: "hcs.testnet.mirrornode.hedera.com:5600",
			ApiAddress:    "https://testnet.mirrornode.hedera.com/api/v1/",
		},
	}

	mirrorClient := mirror_node.NewClient(mirrorNodeConfigByNetwork[hederaNetworkId])

	_, err = transferTokens(client, mirrorClient, parsedSenderAccountId, parsedRecipientAccountId, hederaTokenIDs, hbarAmountHedera)

	if err != nil {
		log.Fatalf("Error transferring tokens and HBAR: %v", err)
	}

	fmt.Println("Tokens and HBAR successfully transferred")
}
