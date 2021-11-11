/*
 * Copyright 2021 LimeChain Ltd.
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

package hedera

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

var mainnetNodes = map[string]hedera.AccountID{
	"35.237.200.180:50211": {Account: 3},
	//"35.186.191.247:50211": {Account: 4},
	"35.192.2.25:50211":    {Account: 5},
	"35.199.161.108:50211": {Account: 6},
	"35.203.82.240:50211":  {Account: 7},
	"35.236.5.219:50211":   {Account: 8},
	"35.197.192.225:50211": {Account: 9},
	//"35.242.233.154:50211": {Account: 10},
	"35.240.118.96:50211":  {Account: 11},
	"35.204.86.32:50211":   {Account: 12},
	"35.234.132.107:50211": {Account: 13},
	"35.236.2.27:50211":    {Account: 14},
	"35.228.11.53:50211":   {Account: 15},
	"34.91.181.183:50211":  {Account: 16},
	"34.86.212.247:50211":  {Account: 17},
	"172.105.247.67:50211": {Account: 18},
	"34.89.87.138:50211":   {Account: 19},
	"34.82.78.255:50211":   {Account: 20},
}

// Node struct holding the hedera.Client. Used to interact with Hedera consensus nodes
type Node struct {
	client *hedera.Client
}

// NewNodeClient creates new instance of hedera.Client based on the provided client configuration
func NewNodeClient(config config.Hedera) *Node {
	var client *hedera.Client
	switch config.Network {
	case "mainnet":
		client = hedera.ClientForMainnet()
		err := client.SetNetwork(mainnetNodes)
		if err != nil {
			log.Fatalf("Could not set mainnet nodes. Error [%s]", err)
		}
	case "testnet":
		client = hedera.ClientForTestnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	default:
		log.Fatalf("Invalid Client Network provided: [%s]", config.Network)
	}

	accID, err := hedera.AccountIDFromString(config.Operator.AccountId)
	if err != nil {
		log.Fatalf("Invalid Operator AccountId provided: [%s]", config.Operator.AccountId)
	}

	privateKey, err := hedera.PrivateKeyFromString(config.Operator.PrivateKey)
	if err != nil {
		log.Fatalf("Invalid Operator PrivateKey provided: [%s]", config.Operator.PrivateKey)
	}

	client.SetOperator(accID, privateKey)

	return &Node{client}
}

// GetClient returns the hedera.Client
func (hc Node) GetClient() *hedera.Client {
	return hc.client
}

// SubmitScheduledTokenMintTransaction creates a token mint transaction and submits it as a scheduled mint transaction
func (hc Node) SubmitScheduledTokenMintTransaction(tokenID hedera.TokenID, amount int64, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	tokenMintTx := hedera.NewTokenMintTransaction().SetTokenID(tokenID).SetAmount(uint64(amount))

	tx, err := tokenMintTx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	signedTransaction, err := tx.
		SignWithOperator(hc.GetClient())
	if err != nil {
		return nil, err
	}

	return hc.submitScheduledTransaction(signedTransaction, payerAccountID, memo)
}

// SubmitScheduledTokenBurnTransaction creates a token burn transaction and submits it as a scheduled burn transaction
func (hc Node) SubmitScheduledTokenBurnTransaction(tokenID hedera.TokenID, amount int64, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	tokenBurnTx := hedera.NewTokenBurnTransaction().SetTokenID(tokenID).SetAmount(uint64(amount))
	tx, err := tokenBurnTx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	signedTransaction, err := tx.
		SignWithOperator(hc.GetClient())
	if err != nil {
		return nil, err
	}

	return hc.submitScheduledTransaction(signedTransaction, payerAccountID, memo)
}

// SubmitTopicConsensusMessage submits the provided message bytes to the
// specified HCS `topicId`
func (hc Node) SubmitTopicConsensusMessage(topicId hedera.TopicID, message []byte) (*hedera.TransactionID, error) {
	txResponse, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicId).
		SetMessage(message).
		Execute(hc.client)

	if err != nil {
		return nil, err
	}

	_, err = hc.checkTransactionReceipt(txResponse)

	return &txResponse.TransactionID, err
}

// SubmitScheduleSign submits a ScheduleSign transaction for a given ScheduleID
func (hc Node) SubmitScheduleSign(scheduleID hedera.ScheduleID) (*hedera.TransactionResponse, error) {
	response, err := hedera.NewScheduleSignTransaction().
		SetScheduleID(scheduleID).
		Execute(hc.GetClient())

	return &response, err
}

// SubmitScheduledTokenTransferTransaction creates a token transfer transaction and submits it as a scheduled transaction
func (hc Node) SubmitScheduledTokenTransferTransaction(
	tokenID hedera.TokenID,
	transfers []transfer.Hedera,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {
	transferTransaction := hedera.NewTransferTransaction()

	for _, t := range transfers {
		transferTransaction.AddTokenTransfer(tokenID, t.AccountID, t.Amount)
	}

	return hc.submitScheduledTransferTransaction(payerAccountID, memo, transferTransaction)
}

// SubmitScheduledHbarTransferTransaction creates an hbar transfer transaction and submits it as a scheduled transaction
func (hc Node) SubmitScheduledHbarTransferTransaction(
	transfers []transfer.Hedera,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {
	transferTransaction := hedera.NewTransferTransaction()

	for _, t := range transfers {
		transferTransaction.AddHbarTransfer(t.AccountID, hedera.HbarFromTinybar(t.Amount))
	}

	return hc.submitScheduledTransferTransaction(payerAccountID, memo, transferTransaction)
}

// submitScheduledTransferTransaction freezes the input transaction, signs with operator and submits it
func (hc Node) submitScheduledTransferTransaction(payerAccountID hedera.AccountID, memo string, tx *hedera.TransferTransaction) (*hedera.TransactionResponse, error) {
	tx, err := tx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	signedTransaction, err := tx.
		SignWithOperator(hc.GetClient())
	if err != nil {
		return nil, err
	}

	return hc.submitScheduledTransaction(signedTransaction, payerAccountID, memo)
}

func (hc Node) submitScheduledTransaction(signedTransaction hedera.ITransaction, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	scheduledTx, err := hedera.NewScheduleCreateTransaction().
		SetScheduledTransaction(signedTransaction)
	if err != nil {
		return nil, err
	}
	scheduledTx = scheduledTx.
		SetPayerAccountID(payerAccountID).
		SetScheduleMemo(memo)

	response, err := scheduledTx.Execute(hc.GetClient())
	return &response, err
}

func (hc Node) checkTransactionReceipt(txResponse hedera.TransactionResponse) (*hedera.TransactionReceipt, error) {
	receipt, err := txResponse.GetReceipt(hc.client)
	if err != nil {
		return nil, err
	}

	if receipt.Status != hedera.StatusSuccess {
		return nil, errors.New(fmt.Sprintf("Transaction [%s] failed with status [%s]", txResponse.TransactionID.String(), receipt.Status))
	}

	return &receipt, err
}
