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

package setup

import (
	"errors"
	"fmt"

	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

var balance = hedera.NewHbar(100)

type DeployResult struct {
	MembersPrivateKeys []hedera.PrivateKey
	MembersPublicKeys  []hedera.PublicKey
	MembersAccountIDs  []hedera.AccountID
	TopicId            *hedera.TopicID
	ConfigTopicId      *hedera.TopicID
	BridgeAccountID    *hedera.AccountID
	PayerAccountID     *hedera.AccountID
	Error              error
}

func Deploy(privateKey *string, accountID *string, adminKey *string, network *string, members *int, previousPrivateKeys []hedera.PrivateKey, treshold uint) DeployResult {
	result := DeployResult{}

	err := ValidateArguments(privateKey, accountID, adminKey)
	if err != nil {
		result.Error = err
		return result
	}

	fmt.Println("-----------Start-----------")
	client := client.Init(*privateKey, *accountID, *network)

	result.MembersPrivateKeys = make([]hedera.PrivateKey, 0, *members)
	result.MembersAccountIDs = make([]hedera.AccountID, 0, *members)

	var privKey hedera.PrivateKey
	for i := 0; i < *members; i++ {
		if i < len(previousPrivateKeys) {
			privKey, err = cryptoCreate(client, &result, &previousPrivateKeys[i])
		} else {
			privKey, err = cryptoCreate(client, &result, nil)
		}
		if err != nil {
			result.Error = fmt.Errorf("failed to create member Private Key. Err: [%s]", err)
			return result
		}

		result.MembersPrivateKeys = append(result.MembersPrivateKeys, privKey)
	}
	fmt.Println("Members Private keys array:", result.MembersPrivateKeys)

	result.MembersPublicKeys = make([]hedera.PublicKey, 0, *members)

	// treshold is 1 because each validator should be able to post to the gossip topic
	topicKey := hedera.KeyListWithThreshold(1)
	for i := 0; i < *members; i++ {
		pubKey := result.MembersPrivateKeys[i].PublicKey()
		result.MembersPublicKeys = append(result.MembersPublicKeys, pubKey)
		topicKey.Add(pubKey)
	}
	fmt.Println("Members Public keys array:", result.MembersPublicKeys)

	adminPublicKey, err := hedera.PublicKeyFromString(*adminKey)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse admin Public Key. Err: [%s]", err)
		return result
	}

	txID, err := hedera.NewTopicCreateTransaction().
		SetAdminKey(adminPublicKey).
		SetSubmitKey(topicKey).
		Execute(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to create topic. Err: [%s]", err)
		return result
	}

	topicReceipt, err := txID.GetReceipt(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to get topic receipt. Err: [%s]", err)
		return result
	}
	result.TopicId = topicReceipt.TopicID
	fmt.Printf("TopicID: %v\n", topicReceipt.TopicID)
	fmt.Println("--------------------------")

	configTxID, err := hedera.NewTopicCreateTransaction().
		SetAdminKey(adminPublicKey).
		SetSubmitKey(topicKey).
		Execute(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to create topic. Err: [%s]", err)
		return result
	}

	configTopicReceipt, err := configTxID.GetReceipt(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to get topic receipt. Err: [%s]", err)
		return result
	}
	result.ConfigTopicId = configTopicReceipt.TopicID
	fmt.Printf("Config TopicID: %v\n", topicReceipt.TopicID)
	fmt.Println("--------------------------")

	custodialKey := hedera.KeyListWithThreshold(treshold)
	for i := 0; i < *members; i++ {
		custodialKey.Add(result.MembersPublicKeys[i])
	}

	// Creating Bridge threshold account
	bridgeAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		Execute(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to create bridge account. Err: [%s]", err)
		return result
	}

	bridgeAccountReceipt, err := bridgeAccount.GetReceipt(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to get bridge account receipt. Err: [%s]", err)
		return result
	}
	result.BridgeAccountID = bridgeAccountReceipt.AccountID
	fmt.Printf("Bridge Account: %v\n", bridgeAccountReceipt.AccountID)
	fmt.Println("--------------------------")

	// Creating Scheduled transaction payer threshold account
	scheduledTxPayerAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(custodialKey).
		SetInitialBalance(balance).
		Execute(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to create payer account. Err: [%s]", err)
		return result
	}
	scheduledTxPayerAccountReceipt, err := scheduledTxPayerAccount.GetReceipt(client)
	if err != nil {
		result.Error = fmt.Errorf("failed to get payer account receipt. Err: [%s]", err)
		return result
	}
	result.PayerAccountID = scheduledTxPayerAccountReceipt.AccountID
	fmt.Printf("Scheduled Tx Payer Account: %v\n", scheduledTxPayerAccountReceipt.AccountID)
	fmt.Printf("Balance: %v\n HBars", balance)
	fmt.Println("---Executed Successfully---")

	return result
}

func ValidateArguments(privateKey *string, accountID *string, adminKey *string) error {
	if *privateKey == "0x0" {
		return errors.New("private key was not provided")
	}
	if *accountID == "0.0" {
		return errors.New("account id was not provided")
	}
	if *adminKey == "" {
		return errors.New("admin key not provided")
	}

	return nil
}

func cryptoCreate(client *hedera.Client, result *DeployResult, previousPrivateKey *hedera.PrivateKey) (hedera.PrivateKey, error) {
	privateKey, _ := hedera.PrivateKeyGenerateEd25519()
	if previousPrivateKey != nil {
		privateKey = *previousPrivateKey
	}
	fmt.Printf("Hedera Private Key: %v\n", privateKey.String())
	fmt.Printf("Hedera Public Key: %v\n", privateKey.PublicKey().String())
	publicKey := privateKey.PublicKey()
	newAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(publicKey).
		SetInitialBalance(balance).
		Execute(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	receipt, err := newAccount.GetReceipt(client)
	if err != nil {
		return hedera.PrivateKey{}, err
	}
	fmt.Printf("AccountID: %v\n", receipt.AccountID)
	result.MembersAccountIDs = append(result.MembersAccountIDs, *receipt.AccountID)
	fmt.Printf("Balance: %v\n HBars", balance)
	fmt.Println("--------------------------")
	return privateKey, nil
}
