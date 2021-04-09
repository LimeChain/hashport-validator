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
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Node struct holding the hedera.Client. Used to interact with Hedera consensus nodes
type Node struct {
	client *hedera.Client
}

// NewNodeClient creates new instance of hedera.Client based on the provided client configuration
func NewNodeClient(config config.Client) *Node {
	var client *hedera.Client
	switch config.NetworkType {
	case "mainnet":
		client = hedera.ClientForMainnet()
	case "testnet":
		client = hedera.ClientForTestnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	default:
		log.Fatalf("Invalid Client NetworkType provided: [%s]", config.NetworkType)
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

func (hc Node) SubmitScheduleSign(scheduleID hedera.ScheduleID) (*hedera.TransactionResponse, error) {
	response, err := hedera.NewScheduleSignTransaction().
		SetScheduleID(scheduleID).
		Execute(hc.GetClient())

	return &response, err
}

// SubmitScheduledTokenTransferTransaction creates a token transfer transaction and submits a scheduled transfer transaction
func (hc Node) SubmitScheduledTokenTransferTransaction(
	tinybarAmount int64,
	tokenID hedera.TokenID,
	recipient,
	sender,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {

	transferTransaction := hedera.NewTransferTransaction().
		AddTokenTransfer(tokenID, recipient, tinybarAmount).
		AddTokenTransfer(tokenID, sender, -tinybarAmount)

	return hc.submitScheduledTransferTransaction(payerAccountID, memo, transferTransaction)
}

// SubmitScheduledHbarTransferTransaction creates a hbar transfer transaction and submits a scheduled transfer transaction
func (hc Node) SubmitScheduledHbarTransferTransaction(tinybarAmount int64,
	recipient,
	sender,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {
	receiveAmount := hedera.HbarFromTinybar(tinybarAmount)
	subtractedAmount := hedera.HbarFromTinybar(-tinybarAmount)

	transferTransaction := hedera.NewTransferTransaction().
		AddHbarTransfer(recipient, receiveAmount).
		AddHbarTransfer(sender, subtractedAmount)

	return hc.submitScheduledTransferTransaction(payerAccountID, memo, transferTransaction)
}

// submitScheduledTransferTransaction freezes the input transaction, signs with operator and submits to HCS
func (hc Node) submitScheduledTransferTransaction(payerAccountID hedera.AccountID, memo string, transaction *hedera.TransferTransaction) (*hedera.TransactionResponse, error) {
	transaction, err := transaction.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	signedTransaction, err := transaction.
		SignWithOperator(hc.GetClient())
	if err != nil {
		return nil, err
	}

	scheduledTx, err := hedera.NewScheduleCreateTransaction().
		SetScheduledTransaction(signedTransaction)
	if err != nil {
		return nil, err
	}
	scheduledTx = scheduledTx.
		SetPayerAccountID(payerAccountID).
		SetTransactionMemo(memo)

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
