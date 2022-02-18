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
	"github.com/stretchr/testify/mock"
)

type MockScheduledService struct {
	mock.Mock
}

func (mss *MockScheduledService) ExecuteScheduledNftTransferTransaction(id string, nftID hedera.NftID, sender hedera.AccountID, receiving hedera.AccountID, onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	panic("implement me")
}

func (mss *MockScheduledService) ExecuteScheduledBurnTransaction(id, asset string, amount int64, status *chan string, onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	mss.Called(id, asset, amount)
}

func (mss *MockScheduledService) ExecuteScheduledMintTransaction(id, asset string, amount int64, status *chan string, onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	mss.Called(id, asset, amount)
}

func (mss *MockScheduledService) ExecuteScheduledTransferTransaction(id, nativeAsset string, transfers []transfer.Hedera, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	mss.Called(id, nativeAsset, transfers)
}
