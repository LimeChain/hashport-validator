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

package service

import "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"

// Scheduled interface is implemented by the Scheduled Service provides business logic for execution of Scheduled Transactions
type Scheduled interface {
	// Execute submits a scheduled transaction and executes provided functions when necessary
	Execute(id, nativeAsset string, transfers []transfer.Hedera, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string))
}
