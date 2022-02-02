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
	"sync"
)

// AwaitMultipleMinedScheduledTransactions is meant to be used when you need to wait all the scheduled transactions to be mined.
func AwaitMultipleMinedScheduledTransactions(
	wg *sync.WaitGroup,
	outTransactionsResults []*bool,
	sourceChainId int64,
	targetChainId int64,
	asset string,
	transferID string,
	callback func(sourceChainId int64, targetChainId int64, nativeAsset string, transferID string, isTransferSuccessful bool)) {

	wg.Wait()

	isTransferSuccessful := true
	for _, result := range outTransactionsResults {
		if result != nil && *result == false {
			isTransferSuccessful = false
			break
		}
	}

	callback(sourceChainId, targetChainId, asset, transferID, isTransferSuccessful)
}
