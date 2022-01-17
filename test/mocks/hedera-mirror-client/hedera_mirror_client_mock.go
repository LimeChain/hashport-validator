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

package hedera_mirror_client

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/stretchr/testify/mock"
)

type MockHederaMirrorClient struct {
	mock.Mock
}

func (m *MockHederaMirrorClient) GetAccountTokenMintTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	panic("implement me")
}

func (m *MockHederaMirrorClient) GetAccountTokenBurnTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	panic("implement me")
}

func (m *MockHederaMirrorClient) GetAccountDebitTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	panic("implement me")
}

func (m *MockHederaMirrorClient) GetAccountCreditTransactionsAfterTimestampString(accountId hedera.AccountID, from string) (*model.Response, error) {
	panic("implement me")
}

func (m *MockHederaMirrorClient) GetScheduledTransaction(transactionID string) (*model.Response, error) {
	panic("implement me")
}

func (m *MockHederaMirrorClient) GetSchedule(scheduleID string) (*model.Schedule, error) {
	panic("implement me")
}

func (m *MockHederaMirrorClient) GetAccountTokenBurnTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*model.Response, error) {
	args := m.Called(accountId, from)

	if args.Get(1) == nil {
		return args.Get(0).(*model.Response), nil
	}
	return args.Get(0).(*model.Response), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetAccountTokenMintTransactionsAfterTimestamp(accountId hedera.AccountID, from int64) (*model.Response, error) {
	args := m.Called(accountId, from)

	if args.Get(1) == nil {
		return args.Get(0).(*model.Response), nil
	}
	return args.Get(0).(*model.Response), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetAccountCreditTransactionsBetween(accountId hedera.AccountID, from, to int64) ([]model.Transaction, error) {
	args := m.Called(accountId, from, to)

	if args.Get(1) == nil {
		return args.Get(0).([]model.Transaction), nil
	}
	return args.Get(0).([]model.Transaction), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetMessagesForTopicBetween(topicId hedera.TopicID, from, to int64) ([]model.Message, error) {
	args := m.Called(topicId, from, to)

	if args.Get(1) == nil {
		return args.Get(0).([]model.Message), nil
	}
	return args.Get(0).([]model.Message), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetMessagesAfterTimestamp(topicId hedera.TopicID, from int64) ([]model.Message, error) {
	args := m.Called(topicId, from)

	if args.Get(1) == nil {
		return args.Get(0).([]model.Message), nil
	}
	return args.Get(0).([]model.Message), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) WaitForTransaction(txId string, onSuccess, onFailure func()) {
	m.Called(txId, onSuccess, onFailure)
}

func (m *MockHederaMirrorClient) GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, milestoneTimestamp int64) (*model.Response, error) {
	args := m.Called(accountId, milestoneTimestamp)

	if args.Get(1) == nil {
		return args.Get(0).(*model.Response), nil
	}
	return args.Get(0).(*model.Response), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetStateProof(transactionID string) ([]byte, error) {
	args := m.Called(transactionID)

	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil
	}
	return args.Get(0).([]byte), args.Get(1).(error)

}

func (m *MockHederaMirrorClient) AccountExists(accountID hedera.AccountID) bool {
	args := m.Called(accountID)
	return args.Get(0).(bool)
}

func (m *MockHederaMirrorClient) GetAccount(accountID string) (*model.AccountsResponse, error) {
	args := m.Called(accountID)

	if args.Get(1) == nil {
		return args.Get(0).(*model.AccountsResponse), nil
	}
	return args.Get(0).(*model.AccountsResponse), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetToken(tokenID string) (*model.TokenResponse, error) {
	args := m.Called(tokenID)

	if args.Get(1) == nil {
		return args.Get(0).(*model.TokenResponse), nil
	}
	return args.Get(0).(*model.TokenResponse), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetNetworkSupply() (*model.NetworkSupplyResponse, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(*model.NetworkSupplyResponse), nil
	}
	return args.Get(0).(*model.NetworkSupplyResponse), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) TopicExists(topicID hedera.TopicID) bool {
	args := m.Called(topicID)
	return args.Get(0).(bool)
}

func (m *MockHederaMirrorClient) GetTransaction(transactionID string) (*model.Response, error) {
	args := m.Called(transactionID)

	if args.Get(1) == nil {
		return args.Get(0).(*model.Response), nil
	}
	return args.Get(0).(*model.Response), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) WaitForScheduledTransaction(txId string, onSuccess, onFailure func()) {
	m.Called(txId /*, onSuccess, onFailure*/)
}
