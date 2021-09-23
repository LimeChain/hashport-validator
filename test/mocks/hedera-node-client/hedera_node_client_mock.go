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

package hedera_node_client

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/mock"
)

type MockHederaNodeClient struct {
	mock.Mock
}

func (m *MockHederaNodeClient) GetClient() *hedera.Client {
	args := m.Called()

	return args.Get(0).(*hedera.Client)
}

func (m *MockHederaNodeClient) SubmitTopicConsensusMessage(topicId hedera.TopicID, message []byte) (*hedera.TransactionID, error) {
	args := m.Called(topicId, message)

	if args.Get(1) == nil {
		return args.Get(0).(*hedera.TransactionID), nil
	}
	return args.Get(0).(*hedera.TransactionID), args.Get(1).(error)
}

func (m *MockHederaNodeClient) SubmitScheduledTokenTransferTransaction(
	tokenID hedera.TokenID,
	transfers []transfer.Hedera,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {
	args := m.Called(tokenID, transfers, payerAccountID, memo)

	if args.Get(1) == nil {
		return args.Get(0).(*hedera.TransactionResponse), nil
	}

	return args.Get(0).(*hedera.TransactionResponse), args.Get(1).(error)
}

func (m *MockHederaNodeClient) SubmitScheduledHbarTransferTransaction(
	transfers []transfer.Hedera,
	payerAccountID hedera.AccountID,
	memo string) (*hedera.TransactionResponse, error) {
	args := m.Called(transfers, payerAccountID, memo)

	if args.Get(1) == nil {
		return args.Get(0).(*hedera.TransactionResponse), nil
	}

	return args.Get(0).(*hedera.TransactionResponse), args.Get(1).(error)
}

func (m *MockHederaNodeClient) SubmitScheduleSign(scheduleID hedera.ScheduleID) (*hedera.TransactionResponse, error) {
	args := m.Called(scheduleID)
	if args.Get(1) == nil {
		return args.Get(0).(*hedera.TransactionResponse), nil
	}
	return args.Get(0).(*hedera.TransactionResponse), args.Get(1).(error)
}

func (m *MockHederaNodeClient) SubmitScheduledTokenMintTransaction(tokenID hedera.TokenID, amount int64, payerAccountID hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	args := m.Called(tokenID, amount, payerAccountID, memo)
	if args.Get(1) == nil {
		return args.Get(0).(*hedera.TransactionResponse), nil
	}
	return args.Get(0).(*hedera.TransactionResponse), args.Get(1).(error)
}

func (m *MockHederaNodeClient) SubmitScheduledTokenBurnTransaction(id hedera.TokenID, amount int64, account hedera.AccountID, memo string) (*hedera.TransactionResponse, error) {
	args := m.Called(id, amount, account, memo)
	if args.Get(1) == nil {
		return args.Get(0).(*hedera.TransactionResponse), nil
	}
	return args.Get(0).(*hedera.TransactionResponse), args.Get(1).(error)
}
