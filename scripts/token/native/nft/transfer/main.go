package main

import (
	"flag"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "", "Hedera Network Type")
	serialNum := flag.Int("serialNum", 0, "Hedera NFT Serial number")
	tokenID := flag.String("tokenID", "0.0", "Hedera NFT Token ID")
	receiver := flag.String("receiver", "0.0", "Hedera Receiver")
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
	if *receiver == "0.0" {
		panic("invalid receiver provided")
	}
	if *serialNum == 0 {
		panic("invalid serial num provided")
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	tokenIDFromString, err := hedera.TokenIDFromString(*tokenID)
	if err != nil {
		panic(err)
	}
	receiverAcc, err := hedera.AccountIDFromString(*receiver)
	if err != nil {
		panic(err)
	}

	// Extract into iterable function
	nftID, err := hedera.NftIDFromString(fmt.Sprintf("%d@%s", serialNum, tokenIDFromString.String()))
	res, err := hedera.NewTransferTransaction().
		AddNftTransfer(nftID, client.GetOperatorAccountID(), receiverAcc).
		Execute(client)
	if err != nil {
		panic(err)
	}
	rec, err := res.GetReceipt(client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s - %s\n", rec.Status, res.TransactionID)
}
