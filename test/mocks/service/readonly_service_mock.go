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
	mirrorNodeTransaction "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/stretchr/testify/mock"
)

type MockReadOnlyService struct {
	mock.Mock
}

func (m *MockReadOnlyService) FindNftTransfer(transferID string, tokenID string, serialNum int64, sender string, receiver string, save func(transactionID string, scheduleID string, status string) error) {
	m.Called(transferID, tokenID, serialNum, sender, receiver, save)
}

func (m *MockReadOnlyService) FindTransfer(transferID string, fetch func() (*mirrorNodeTransaction.Response, error), save func(transactionID, scheduleID, status string) error) {
	m.Called(transferID, fetch, save)
}

func (m *MockReadOnlyService) FindAssetTransfer(transferID string, asset string, transfers []transfer.Hedera, fetch func() (*mirrorNodeTransaction.Response, error), save func(transactionID, scheduleID, status string) error) {
	m.Called(transferID, asset, transfers, fetch, save)
}

func (m *MockReadOnlyService) FindScheduledNftAllowanceApprove(t *payload.Transfer, sender hedera.AccountID, save func(transactionID string, scheduleID string, status string) error) {
	m.Called(t, sender, save)
}
