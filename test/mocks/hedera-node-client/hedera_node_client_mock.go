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
	"github.com/hashgraph/hedera-sdk-go"
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
