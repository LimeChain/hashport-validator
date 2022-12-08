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

package service

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
)

// Scheduled interface is implemented by the Scheduled Service
// Provides business logic for execution of Scheduled Transactions
type Scheduled interface {
	// ExecuteScheduledTransferTransaction submits a scheduled transfer transaction and executes provided functions when necessary
	ExecuteScheduledTransferTransaction(id, asset string, transfers []transfer.Hedera, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string))
	// ExecuteScheduledMintTransaction submits a scheduled mint transaction and executes provided functions when necessary
	ExecuteScheduledMintTransaction(id, asset string, amount int64, status *chan string, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string))
	// ExecuteScheduledBurnTransaction submits a scheduled burn transaction and executes provided functions when necessary
	ExecuteScheduledBurnTransaction(id, asset string, amount int64, status *chan string, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string))
	// ExecuteScheduledNftTransferTransaction submits a scheduled nft transfer transaction and executes provided functions when necessary
	ExecuteScheduledNftTransferTransaction(id string, nftID hedera.NftID, sender hedera.AccountID, receiving hedera.AccountID, approved bool, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string))
	// ExecuteScheduledNftAllowTransaction submits a scheduled NFT allow transaction and executes provided functions when necessary
	ExecuteScheduledNftAllowTransaction(
		id string, nftID hedera.NftID, owner hedera.AccountID, spender hedera.AccountID,
		onExecutionSuccess func(txId, scheduleId string), onExecutionFail, onSuccess, onFail func(txId string))
}
