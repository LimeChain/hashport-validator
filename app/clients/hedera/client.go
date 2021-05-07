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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/transaction"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Node struct holding the hedera.Client. Used to interact with Hedera consensus nodes
type Node struct {
	client *hedera.Client
}

// NewNodeClient creates new instance of hedera.Client based on the provided client configuration
func NewNodeClient(config config.Hedera) *Node {
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
	memo string) (transaction.Response, error) {
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
	memo string) (transaction.Response, error) {
	transferTransaction := hedera.NewTransferTransaction()

	for _, t := range transfers {
		transferTransaction.AddHbarTransfer(t.AccountID, hedera.HbarFromTinybar(t.Amount))
	}

	return hc.submitScheduledTransferTransaction(payerAccountID, memo, transferTransaction)
}

// submitScheduledTransferTransaction freezes the input transaction, signs with operator and submits to HCS
func (hc Node) submitScheduledTransferTransaction(payerAccountID hedera.AccountID, memo string, tx *hedera.TransferTransaction) (transaction.Response, error) {
	tx, err := tx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	signedTransaction, err := tx.
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
		SetScheduleMemo(memo)

	response, err := scheduledTx.Execute(hc.GetClient())
	return hederahelper.NewWrappedTransactionResponse(response), err
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
