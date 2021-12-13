package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
)

func main() {
	message := flag.String("message", "", "message to be signed")
	// Private key
	privateKey := flag.String("privateKey", "", "private key with which the message will be signed")

	flag.Parse()
	if *privateKey == "" {
		panic("no private key provided")
	}
	if *message == "" {
		panic("no message provided")
	}

	signer := signer.NewEVMSigner(*privateKey)

	msgBytes, err := hex.DecodeString(*message)
	if err != nil {
		panic(err)
	}

	signature, err := signer.Sign(msgBytes)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(signature))
}
