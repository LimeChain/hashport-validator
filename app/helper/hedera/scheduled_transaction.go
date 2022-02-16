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
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"sync"
)

// AwaitMultipleScheduledTransactions is meant to be used when you need to wait all the scheduled transactions to be mined.
func AwaitMultipleScheduledTransactions(
	outParams *OutParams,
	sourceChainId uint64,
	targetChainId uint64,
	asset string,
	transferID string,
	callback func(sourceChainId, targetChainId uint64, nativeAsset string, transferID string, isTransferSuccessful bool)) {

	if outParams.waitGroup == nil {
		panic(fmt.Sprintf("[%v] Await group is nil. [AwaitMultipleScheduledTransactions]", transferID))
	}
	if outParams.outTransactionsResults == nil {
		panic(fmt.Sprintf("[%v] Out transaction results are nil. [AwaitMultipleScheduledTransactions]", transferID))
	}

	outParams.waitGroup.Wait()

	isTransferSuccessful := true
	for _, result := range *outParams.outTransactionsResults {
		if result != nil && *result == false {
			isTransferSuccessful = false
			break
		}
	}

	callback(sourceChainId, targetChainId, asset, transferID, isTransferSuccessful)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/* Struct for encapsulating the logic needed for out parameters used in  'AwaitMultipleScheduledTransactions' */
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type OutParams struct {
	waitGroup              *sync.WaitGroup
	outTransactionsResults *[]*bool
}

// FeeOutParams Helper struct for AwaitMultipleScheduledTransactions
type FeeOutParams struct {
	*OutParams
	currentIndex             int
	countOfAllSplitTransfers int
}

// NewFeeOutParams is used to initialize out parameters needed for awaiting fee transfers to be mined.
// Helper func for AwaitMultipleScheduledTransactions.
func NewFeeOutParams(countOfAllSplitTransfers int) *FeeOutParams {
	params := new(FeeOutParams)
	params.OutParams = new(OutParams)

	countOfTransferResults := countOfAllSplitTransfers - 1
	if countOfTransferResults <= 0 {
		countOfTransferResults = 1
	}
	transferResults := make([]*bool, countOfTransferResults)
	params.outTransactionsResults = &transferResults
	params.waitGroup = new(sync.WaitGroup)
	params.waitGroup.Add(countOfAllSplitTransfers)
	params.currentIndex = 0
	params.countOfAllSplitTransfers = countOfAllSplitTransfers

	return params
}

// HandleResultForAwaitedTransfer is used to handle result from mined scheduled transfer for Fees.
// Helper func to collect results for AwaitMultipleScheduledTransactions.
func (params *FeeOutParams) HandleResultForAwaitedTransfer(result *bool, hasReceiver bool, splitTransfer []model.Hedera) {
	defer params.waitGroup.Done()

	if hasReceiver {
		if len(splitTransfer) > 1 {
			if params.countOfAllSplitTransfers > 1 {
				// Appending the result if the split transfers are more than 1 and contain other transfer than the one for the users
				updatedTransferResults := append(*params.outTransactionsResults, result)
				params.outTransactionsResults = &updatedTransferResults
			} else {
				// Setting the result if the split transfer contains other transfer than the one for the user
				(*params.outTransactionsResults)[params.currentIndex] = result
				params.currentIndex += 1
			}
		}
	} else {
		(*params.outTransactionsResults)[params.currentIndex] = result
		params.currentIndex += 1
	}
}

// UserOutParams Helper struct for AwaitMultipleScheduledTransactions
type UserOutParams struct {
	*OutParams
}

// NewUserOutParams is used to initialize out parameters needed for awaiting user transfer to be mined.
// Helper func for AwaitMultipleScheduledTransactions.
func NewUserOutParams() *UserOutParams {
	params := new(UserOutParams)
	params.OutParams = new(OutParams)
	transferResults := make([]*bool, 1)
	params.outTransactionsResults = &transferResults
	params.waitGroup = new(sync.WaitGroup)
	params.waitGroup.Add(1)

	return params
}

// HandleResultForAwaitedTransfer is used to handle result from mined scheduled transfer for User.
// Helper func to collect results for AwaitMultipleScheduledTransactions.
func (params *UserOutParams) HandleResultForAwaitedTransfer(result *bool, hasReceiver bool) {
	if hasReceiver {
		(*params.outTransactionsResults)[0] = result
		params.waitGroup.Done()
	}
}
