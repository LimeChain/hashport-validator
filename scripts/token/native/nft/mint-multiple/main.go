package main

import (
	"flag"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	client "github.com/limechain/hedera-eth-bridge-validator/scripts"
	"strings"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	tokenID := flag.String("tokenID", "0.0", "Hedera NFT Token ID")
	metadataArr := flag.String("metadataArr", "", "Array of Metadata, split by `,`")
	flag.Parse()
	if *privateKey == "0x0" {
		panic("Private key was not provided")
	}
	if *accountID == "0.0" {
		panic("Account id was not provided")
	}
	if *tokenID == "0.0" {
		panic("Bridge id was not provided")
	}
	if *metadataArr == "" {
		panic("no metadatas provided")
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	metadataSlice := strings.Split(*metadataArr, ",")

	tokenIDFromString, err := hedera.TokenIDFromString(*tokenID)
	if err != nil {
		panic(err)
	}
	for _, metadata := range metadataSlice {
		receipt := mintNFT(client, tokenIDFromString, []byte(metadata))
		fmt.Println("Mint transaction status:", receipt.Status)
		fmt.Println("Serial numbers: ", receipt.SerialNumbers)
	}
}

func mintNFT(client *hedera.Client, token hedera.TokenID, metadata []byte) hedera.TransactionReceipt {
	associateTX, err := hedera.
		NewTokenMintTransaction().
		SetTokenID(token).
		SetMetadata(metadata).
		Execute(client)
	if err != nil {
		panic(err)
	}

	receipt, err := associateTX.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	return receipt
}
