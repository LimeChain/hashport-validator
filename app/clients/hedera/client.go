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

package hedera

import (
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Node struct holding the hedera.Client. Used to interact with Hedera consensus nodes
type Node struct {
	client   *hedera.Client
	maxRetry int
	logger   *log.Entry
}

// NewNodeClient creates new instance of hedera.Client based on the provided client configuration
func NewNodeClient(cfg config.Hedera) *Node {
	var client *hedera.Client
	switch cfg.Network {
	case "mainnet":
		client = hedera.ClientForMainnet()
	case "testnet":
		client = hedera.ClientForTestnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	default:
		log.Fatalf("Invalid Client Network provided: [%s]", cfg.Network)
	}
	if len(cfg.Rpc) > 0 {
		log.Debugf("Setting provided RPC nodes for [%s].", cfg.Network)
		err := client.SetNetwork(cfg.Rpc)
		if err != nil {
			log.Fatalf("Could not set rpc nodes [%s]. Error: [%s]", cfg.Rpc, err)
		}
	} else {
		log.Debugf("Setting default node rpc urls for [%s].", cfg.Network)
	}

	accID, err := hedera.AccountIDFromString(cfg.Operator.AccountId)
	if err != nil {
		log.Fatalf("Invalid Operator AccountId provided: [%s]", cfg.Operator.AccountId)
	}

	privateKey, err := hedera.PrivateKeyFromString(cfg.Operator.PrivateKey)
	if err != nil {
		log.Fatalf("Invalid Operator PrivateKey provided: [%s]", cfg.Operator.PrivateKey)
	}

	client.SetOperator(accID, privateKey)

	return &Node{
		client:   client,
		maxRetry: cfg.MaxRetry,
		logger:   config.GetLoggerFor("Hedera Node Client"),
	}
}

// GetClient returns the hedera.Client
func (hc Node) GetClient() *hedera.Client {
	return hc.client
}

// SubmitScheduledTokenMintTransaction creates a token mint transaction and submits it as a scheduled mint transaction
func (hc Node) SubmitScheduledTokenMintTransaction(tokenID hedera.TokenID, amount int64, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	tokenMintTx := hedera.NewTokenMintTransaction().
		SetTokenID(tokenID).
		SetAmount(uint64(amount)).
		SetMaxRetry(hc.maxRetry)

	tx, err := tokenMintTx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	hc.logger.Debugf("[%s] - Signing transaction with ID: [%s] and Node Account IDs: %v", memo, tx.GetTransactionID().String(), tx.GetNodeAccountIDs())
	signedTransaction, err := tx.
		SignWithOperator(hc.GetClient())
	if err != nil {
		return nil, err
	}

	return hc.submitScheduledTransaction(signedTransaction, payerAccountID, memo)
}

// SubmitScheduledTokenBurnTransaction creates a token burn transaction and submits it as a scheduled burn transaction
func (hc Node) SubmitScheduledTokenBurnTransaction(tokenID hedera.TokenID, amount int64, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	tokenBurnTx := hedera.NewTokenBurnTransaction().
		SetTokenID(tokenID).
		SetAmount(uint64(amount)).
		SetMaxRetry(hc.maxRetry)
	tx, err := tokenBurnTx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	hc.logger.Debugf("[%s] - Signing transaction with ID: [%s] and Node Account IDs: %v", memo, tx.GetTransactionID().String(), tx.GetNodeAccountIDs())
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
	tx, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicId).
		SetMessage(message).
		SetMaxRetry(hc.maxRetry).
		FreezeWith(hc.client)

	if err != nil {
		return nil, err
	}

	hc.logger.Debugf("Submit Topic Consensus Message, with transaction ID [%s] and Node Account IDs: %v",
		tx.GetTransactionID(),
		tx.GetNodeAccountIDs(),
	)

	response, err := tx.Execute(hc.GetClient())

	if err != nil {
		return nil, err
	}

	_, err = hc.checkTransactionReceipt(response)

	return &response.TransactionID, err

}

// SubmitScheduleSign submits a ScheduleSign transaction for a given ScheduleID
func (hc Node) SubmitScheduleSign(scheduleID hedera.ScheduleID) (*hedera.TransactionResponse, error) {
	tx, err := hedera.NewScheduleSignTransaction().
		SetScheduleID(scheduleID).
		SetMaxRetry(hc.maxRetry).
		FreezeWith(hc.GetClient())

	if err != nil {
		return nil, err
	}

	hc.logger.Debugf("Submit Schedule Sign, with transaction ID [%s], schedule ID: [%s] and Node Account IDs: %v",
		tx.GetTransactionID(),
		tx.GetScheduleID(),
		tx.GetNodeAccountIDs(),
	)
	response, err := tx.Execute(hc.GetClient())

	return &response, err
}

// SubmitScheduledTokenTransferTransaction creates a token transfer transaction and submits it as a scheduled transaction
func (hc Node) SubmitScheduledTokenTransferTransaction(
	tokenID hedera.TokenID,
	transfers []transfer.Hedera,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {
	transferTransaction := hedera.NewTransferTransaction().
		SetMaxRetry(hc.maxRetry)

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
	transferTransaction := hedera.NewTransferTransaction().
		SetMaxRetry(hc.maxRetry)

	for _, t := range transfers {
		transferTransaction.AddHbarTransfer(t.AccountID, hedera.HbarFromTinybar(t.Amount))
	}

	return hc.submitScheduledTransferTransaction(payerAccountID, memo, transferTransaction)
}

func (hc Node) SubmitScheduledNftTransferTransaction(
	nftID hedera.NftID,
	payerAccount hedera.AccountID,
	sender hedera.AccountID,
	receiving hedera.AccountID,
	memo string, approved bool) (*hedera.TransactionResponse, error) {

	transferTransaction := hedera.
		NewTransferTransaction().
		AddApprovedNftTransfer(nftID, sender, receiving, approved)

	return hc.submitScheduledTransferTransaction(payerAccount, memo, transferTransaction)
}

func (hc Node) TransactionReceiptQuery(transactionID hedera.TransactionID, nodeAccIds []hedera.AccountID) (hedera.TransactionReceipt, error) {
	return hedera.NewTransactionReceiptQuery().
		SetTransactionID(transactionID).
		SetNodeAccountIDs(nodeAccIds).
		SetMaxRetry(hc.maxRetry).
		Execute(hc.GetClient())
}

func (hc Node) SubmitScheduledNftApproveTransaction(
	payer hedera.AccountID,
	memo string,
	nftId hedera.NftID,
	owner hedera.AccountID,
	spender hedera.AccountID) (*hedera.TransactionResponse, error) {
	tx := hedera.NewAccountAllowanceApproveTransaction().
		ApproveTokenNftAllowance(nftId, owner, spender)

	return hc.submitScheduledAllowTransaction(payer, memo, tx)
}

func (hc Node) submitScheduledAllowTransaction(payer hedera.AccountID, memo string, tx *hedera.AccountAllowanceApproveTransaction) (*hedera.TransactionResponse, error) {
	tx, err := tx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	signedTx, err := tx.SignWithOperator(hc.GetClient())
	if err != nil {
		return nil, err
	}

	return hc.submitScheduledTransaction(signedTx, payer, memo)
}

// submitScheduledTransferTransaction freezes the input transaction, signs with operator and submits it
func (hc Node) submitScheduledTransferTransaction(payerAccountID hedera.AccountID, memo string, tx *hedera.TransferTransaction) (*hedera.TransactionResponse, error) {
	tx, err := tx.FreezeWith(hc.GetClient())
	if err != nil {
		return nil, err
	}

	hc.logger.Debugf("[%s] - Signing transaction with ID: [%s] and Node Account IDs: %v", memo, tx.GetTransactionID().String(), tx.GetNodeAccountIDs())
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
		return nil, fmt.Errorf("Transaction [%s] failed with status [%s]", txResponse.TransactionID.String(), receipt.Status)
	}

	return &receipt, err
}
