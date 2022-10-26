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
	validateParams(privateKeys, transaction)

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

	switch tx := deserialized.(type) {
	case hedera.TransferTransaction:
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
	case hedera.TopicUpdateTransaction:
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
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
	case hedera.TokenCreateTransaction:
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
	case hedera.TokenMintTransaction:
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
	case hedera.TokenAssociateTransaction:
		ref := &tx
		for _, key := range keys {
			ref = ref.Sign(key)
		}
		bytes, err := ref.ToBytes()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(bytes))
	case hedera.TopicMessageSubmitTransaction:
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

func validateParams(privateKeys *string, transaction *string) {
	if *privateKeys == "" {
		panic("no private keys provided")
	}
	if *transaction == "" {
		panic("transaction has not been provided")
	}
}
