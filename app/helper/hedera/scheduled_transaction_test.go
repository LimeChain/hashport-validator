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
	"github.com/hashgraph/hedera-sdk-go/v2"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var (
	outParams     *OutParams
	sourceChainId = uint64(0)
	targetChainId = uint64(1)
	asset         = "asset"
	transferId    = "txId"
	callback      func(sourceChainId, targetChainId uint64, nativeAsset string, transferID string, isTransferSuccessful bool)
	feeOutParams  = &FeeOutParams{
		OutParams: &OutParams{
			waitGroup: &sync.WaitGroup{},
		},
		currentIndex:             0,
		countOfAllSplitTransfers: 2,
	}
	acc1, _      = hedera.AccountIDFromString("0.0.1")
	acc2, _      = hedera.AccountIDFromString("0.0.2")
	twoTransfers = []model.Hedera{
		{
			AccountID: acc1,
			Amount:    -2500,
		},
		{
			AccountID: acc2,
			Amount:    2500,
		},
	}
	singleTransfer = []model.Hedera{
		{
			AccountID: acc1,
			Amount:    5000,
		},
	}
	expectedUserOutParams = &UserOutParams{
		OutParams: &OutParams{},
	}
)

func setup() {
	outResults := make([]*bool, 0, 0)
	outParams = &OutParams{
		waitGroup:              &sync.WaitGroup{},
		outTransactionsResults: &outResults,
	}
	callback = func(sourceChainId, targetChainId uint64, nativeAsset string, transferID string, isTransferSuccessful bool) {
		return
	}

	outTransactionResults := make([]*bool, 0, 0)
	outTransactionResults = append(outTransactionResults, nil)
	outTransactionResults = append(outTransactionResults, nil)
	feeOutParams.outTransactionsResults = &outTransactionResults

	expectedUserOutParams.OutParams = new(OutParams)
	transferResults := make([]*bool, 1)
	expectedUserOutParams.outTransactionsResults = &transferResults
	expectedUserOutParams.waitGroup = new(sync.WaitGroup)
}

func Test_AwaitMultipleScheduledTransactions_WithUnsuccessfulTransfer(t *testing.T) {
	setup()
	callback = func(sourceChainId, targetChainId uint64, nativeAsset string, transferID string, isTransferSuccessful bool) {
		assert.False(t, isTransferSuccessful)
	}
	outTransactionResults := make([]*bool, 0, 0)
	f := false
	outTransactionResults = append(outTransactionResults, &f)
	outParams.outTransactionsResults = &outTransactionResults

	AwaitMultipleScheduledTransactions(
		outParams,
		sourceChainId,
		targetChainId,
		asset,
		transferId,
		callback)
}

func Test_AwaitMultipleScheduledTransactions_WithNilWaitGroup(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	setup()
	outParams.waitGroup = nil

	AwaitMultipleScheduledTransactions(
		outParams,
		sourceChainId,
		targetChainId,
		asset,
		transferId,
		callback)
}

func Test_AwaitMultipleScheduledTransactions_WithNilOutTransactionResults(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	setup()
	outParams.outTransactionsResults = nil

	AwaitMultipleScheduledTransactions(
		outParams,
		sourceChainId,
		targetChainId,
		asset,
		transferId,
		callback)
}

func Test_NewFeeOutParams(t *testing.T) {
	setup()
	expected := &FeeOutParams{
		OutParams:                outParams,
		currentIndex:             0,
		countOfAllSplitTransfers: 1,
	}

	expected.OutParams.waitGroup.Add(1)
	results := make([]*bool, 0, 0)
	results = append(results, nil)
	expected.OutParams.outTransactionsResults = &results

	actual := NewFeeOutParams(1)
	assert.Equal(t, expected, actual)
}

func Test_FeeOutParams_HandleResultForAwaitedTransfer_WithoutReceiver(t *testing.T) {
	setup()

	tr := true
	expectedFeeOutParams := *feeOutParams
	expectedFeeOutParams.currentIndex += 1
	(*expectedFeeOutParams.outTransactionsResults)[expectedFeeOutParams.currentIndex] = &tr

	feeOutParams.waitGroup.Add(1)
	feeOutParams.HandleResultForAwaitedTransfer(&tr, false, twoTransfers)

	actualFeeOutParams := *feeOutParams
	assert.Equal(t, expectedFeeOutParams, actualFeeOutParams)
}

func Test_FeeOutParams_HandleResultForAwaitedTransfer_WithReceiver(t *testing.T) {
	setup()

	expectedFeeOutParams := *feeOutParams
	tr := true
	(*expectedFeeOutParams.outTransactionsResults)[expectedFeeOutParams.currentIndex] = &tr

	feeOutParams.waitGroup.Add(1)
	feeOutParams.HandleResultForAwaitedTransfer(&tr, true, twoTransfers)

	actualFeeOutParams := *feeOutParams
	assert.Equal(t, expectedFeeOutParams, actualFeeOutParams)
}

func Test_FeeOutParams_HandleResultForAwaitedTransfer_WithOneSplitTransferAndReceiver(t *testing.T) {
	setup()

	expectedFeeOutParams := *feeOutParams
	tr := true

	feeOutParams.waitGroup.Add(1)
	feeOutParams.HandleResultForAwaitedTransfer(&tr, true, singleTransfer)

	actualFeeOutParams := *feeOutParams
	assert.Equal(t, expectedFeeOutParams, actualFeeOutParams)
}

func Test_NewUserOutParams(t *testing.T) {
	setup()

	expectedUserOutParams.waitGroup.Add(1)
	actual := NewUserOutParams()

	assert.Equal(t, expectedUserOutParams, actual)
}

func Test_UserOutParams_HandleResultForAwaitedTransfer_WithReceiver(t *testing.T) {
	setup()

	actualUserOutParams := *expectedUserOutParams
	expectedUserOutParams.waitGroup = new(sync.WaitGroup)
	tr := true
	(*expectedUserOutParams.outTransactionsResults)[0] = &tr

	actualUserOutParams.waitGroup.Add(1)
	actualUserOutParams.HandleResultForAwaitedTransfer(&tr, true)
	assert.Equal(t, *expectedUserOutParams, actualUserOutParams)
}

func Test_UserOutParams_HandleResultForAwaitedTransfer_WithoutReceiver(t *testing.T) {
	setup()

	actualUserOutParams := *expectedUserOutParams
	tr := true

	actualUserOutParams.HandleResultForAwaitedTransfer(&tr, false)
	assert.Equal(t, *expectedUserOutParams, actualUserOutParams)
}
