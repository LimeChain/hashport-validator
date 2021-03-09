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
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/stretchr/testify/mock"
)

type MockHederaMirrorClient struct {
	mock.Mock
}

func (m *MockHederaMirrorClient) GetMessagesForTopicAfterTimestamp(topicId hedera.TopicID, timestamp int64) (*mirror_node.Messages, error) {
	args := m.Called(topicId, timestamp)

	if args.Get(1) == nil {
		return args.Get(0).(*mirror_node.Messages), nil
	}
	return args.Get(0).(*mirror_node.Messages), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetAccountCreditTransactionsAfterTimestamp(accountId hedera.AccountID, milestoneTimestamp int64) (*mirror_node.Response, error) {
	args := m.Called(accountId, milestoneTimestamp)

	if args.Get(1) == nil {
		return args.Get(0).(*mirror_node.Response), nil
	}
	return args.Get(0).(*mirror_node.Response), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetAccountTransaction(transactionID string) (*mirror_node.Response, error) {
	args := m.Called(transactionID)

	if args.Get(1) == nil {
		return args.Get(0).(*mirror_node.Response), nil
	}
	return args.Get(0).(*mirror_node.Response), args.Get(1).(error)
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
