package main

/*
go run ./scripts/common/unstuck-hedera-evm/unstuck-hedera-evm.go \
	--privateKey "" \
	--transactionId "0.0.1320-1679324186-455906340" \
	--sourceChainId 296 \
	--targetChainId 80001 \
	--targetAsset 0xaA6844fBb7Df9f90FC135f9BCd6F592550e2Fef5 \
	--receiver 0xB075D644d3C46735C8c34AD61a1dEa146950a3F5 \
	--amount 28169014087

*/

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"

	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
)

type Transfer struct {
	TransactionId string
	SourceChainId uint64
	TargetChainId uint64
	TargetAsset   string
	Receiver      string
	Amount        string
}

func main() {
	privateKeys := flag.String("privateKeys", "", "private keys with which the message will be signed")

	transactionId := flag.String("transactionId", "", "0.0.1320-1679317679-373659593")
	sourceChainId := flag.Uint64("sourceChainId", 0, "296")
	targetChainId := flag.Uint64("targetChainId", 0, "80001")
	targetAsset := flag.String("targetAsset", "", "0xaA6844fBb7Df9f90FC135f9BCd6F592550e2Fef5")
	receiver := flag.String("receiver", "", "0xB075D644d3C46735C8c34AD61a1dEa146950a3F5")
	amount := flag.String("amount", "", "84772370487")

	flag.Parse()
	if *privateKeys == "" {
		panic("no private key provided")
	}
	if *transactionId == "" {
		panic("no transactionId provided")
	}
	if *sourceChainId == 0 {
		panic("no sourceChainId key provided")
	}
	if *targetChainId == 0 {
		panic("no targetChainId key provided")
	}
	if *targetAsset == "" {
		panic("no targetAsset provided")
	}
	if *receiver == "" {
		panic("no receiver provided")
	}
	if *amount == "" {
		panic("no amount provided")
	}

	tm := Transfer{
		TransactionId: *transactionId,
		SourceChainId: *sourceChainId,
		TargetChainId: *targetChainId,
		TargetAsset:   *targetAsset,
		Receiver:      *receiver,
		Amount:        *amount,
	}

	authMsgHash, err := auth_message.EncodeFungibleBytesFrom(
		tm.SourceChainId,
		tm.TargetChainId,
		tm.TransactionId,
		tm.TargetAsset,
		tm.Receiver,
		tm.Amount,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(hex.EncodeToString(authMsgHash))

	prKeysSlice := strings.Split(*privateKeys, ",")
	var signers []*signer.Signer

	for i := 0; i < len(prKeysSlice); i++ {
		s := signer.NewEVMSigner(prKeysSlice[i])
		if err != nil {
			panic(err)
		}
		signers = append(signers, s)
	}

	for _, s := range signers {
		signature, err := s.Sign(authMsgHash)
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(signature))
	}
}