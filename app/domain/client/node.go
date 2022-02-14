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

package client

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
)

type HederaNode interface {
	// GetClient returns the underlying Hedera SDK client
	GetClient() *hedera.Client
	// SubmitTopicConsensusMessage submits the provided message bytes to the
	// specified HCS `topicId`
	SubmitTopicConsensusMessage(topicId hedera.TopicID, message []byte) (*hedera.TransactionID, error)
	// SubmitScheduledTokenTransferTransaction creates a token transfer transaction and submits it as a scheduled transaction
	SubmitScheduledTokenTransferTransaction(tokenID hedera.TokenID, transfers []transfer.Hedera, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error)
	// SubmitScheduledHbarTransferTransaction creates an hbar transfer transaction and submits it as a scheduled transfer transaction
	SubmitScheduledHbarTransferTransaction(transfers []transfer.Hedera, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error)
	// SubmitScheduleSign submits a ScheduleSign transaction for a given ScheduleID
	SubmitScheduleSign(scheduleID hedera.ScheduleID) (*hedera.TransactionResponse, error)
	// SubmitScheduledTokenMintTransaction creates a token mint transaction and submits it as a scheduled mint transaction
	SubmitScheduledTokenMintTransaction(tokenID hedera.TokenID, amount int64, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error)
	// SubmitScheduledTokenBurnTransaction creates a token burn transaction and submits it as a scheduled burn transaction
	SubmitScheduledTokenBurnTransaction(id hedera.TokenID, amount int64, account hedera.AccountID, memo string) (*hedera.TransactionResponse, error)
}
