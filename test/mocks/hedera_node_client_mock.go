package mocks

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
