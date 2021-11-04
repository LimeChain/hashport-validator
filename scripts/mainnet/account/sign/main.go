package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"strings"
)

func main() {
	// Transaction bytes in hex
	transaction := flag.String("transaction", "", "Hedera to-be-signed Transaction")
	// Keys that are key to the bridge ID, which need to sign the transaction
	privateKeys := flag.String("privateKeys", "", "The private keys")

	flag.Parse()
	if *privateKeys == "" {
		panic("no private keys provided")
	}
	if *transaction == "" {
		panic("transaction has not been provided")
	}

	prKeysSlice := strings.Split(*privateKeys, ",")
	var keys []hedera.PrivateKey
	for i := 0; i < len(prKeysSlice); i++ {
		privateKeyFromStr, err := hedera.PrivateKeyFromString(prKeysSlice[i])
		if err != nil {
			panic(err)
		}
		keys = append(keys, privateKeyFromStr)
	}

	decoded, err := hex.DecodeString(*transaction)
	if err != nil {
		panic(err)
	}

	deserialized, err := hedera.TransactionFromBytes(decoded)
	if err != nil {
		panic(fmt.Sprintf("failed to parse transaction. err [%s]", err))
	}

	switch deserialized.(type) {
	case hedera.TopicUpdateTransaction:
		tx := deserialized.(hedera.TopicUpdateTransaction)
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
		break
	case hedera.TokenUpdateTransaction:
		tx := deserialized.(hedera.TokenUpdateTransaction)
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
		break
	case hedera.AccountUpdateTransaction:
		tx := deserialized.(hedera.AccountUpdateTransaction)
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
	default:
		panic("invalid tx type provided")
	}
}
