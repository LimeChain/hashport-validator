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

type HederaNodeClient struct {
	client *hedera.Client
}

var (
	// TODO: remove
	hederaNodeID, _ = hedera.AccountIDFromString("0.0.4")
	hederaNodeIDs   = []hedera.AccountID{hederaNodeID}
)

func NewNodeClient(config config.Client) *HederaNodeClient {
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

	return &HederaNodeClient{client}
}

func (hc *HederaNodeClient) GetClient() *hedera.Client {
	return hc.client
}

func (hc *HederaNodeClient) SubmitTopicConsensusMessage(topicId hedera.TopicID, message []byte) (*hedera.TransactionID, error) {
	txResponse, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicId).
		SetMessage(message).
		Execute(hc.client)

	if err != nil {
		return nil, err
	}

	return hc.checkTransactionReceipt(txResponse)
}

func (hc *HederaNodeClient) SubmitScheduledTransaction(tinybarAmount int64, recipient, payerAccountID hedera.AccountID, nonce string) (*hedera.TransactionID, error) {
	receiveAmount := hedera.HbarFromTinybar(tinybarAmount)
	subtractedAmount := hedera.HbarFromTinybar(-tinybarAmount)

	// todo: create TransactionID with nonce and set it in TransferTransaction
	txnId := hedera.TransactionID{
		AccountID: payerAccountID,
	}

	transferTransaction, err := hedera.NewTransferTransaction().
		SetTransactionID(txnId).
		AddHbarTransfer(recipient, receiveAmount).
		AddHbarTransfer(payerAccountID, subtractedAmount).
		SetNodeAccountIDs(hederaNodeIDs).
		FreezeWith(hc.GetClient())

	if err != nil {
		return nil, err
	}

	signedTransaction, err := transferTransaction.
		SignWithOperator(hc.GetClient())

	if err != nil {
		return nil, err
	}

	scheduledTx := hedera.NewScheduleCreateTransaction().
		SetTransaction(&signedTransaction.Transaction).
		// todo: we can set some kind of memo to showcase bridge transfers (e.g. "Bridge ETH -> Hedera Transfer")
		SetMemo(nonce). // todo: replace when TransactionID nonce is introduced
		SetPayerAccountID(payerAccountID)

	scheduledTx = scheduledTx.SetTransactionID(hedera.TransactionIDGenerate(hc.client.GetOperatorAccountID()))
	txResponse, err := scheduledTx.Execute(hc.GetClient())
	if err != nil {
		return nil, err
	}

	return hc.checkTransactionReceipt(txResponse)
}

func (hc *HederaNodeClient) checkTransactionReceipt(txResponse hedera.TransactionResponse) (*hedera.TransactionID, error) {
	receipt, err := txResponse.GetReceipt(hc.client)
	if err != nil {
		return nil, err
	}

	if receipt.Status != hedera.StatusSuccess {
		return nil, errors.New(fmt.Sprintf("Transaction [%s] failed with status [%s]", txResponse.TransactionID.String(), receipt.Status))
	}

	return &txResponse.TransactionID, err
}
