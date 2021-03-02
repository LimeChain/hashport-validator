package message

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(message *message.TransactionMessage) error {
	args := m.Called(message)
	return args.Get(0).(error)
}

func (m *MockMessageRepository) GetTransaction(txId, signature, hash string) (*message.TransactionMessage, error) {
	args := m.Called(txId, signature, hash)
	return args.Get(0).(*message.TransactionMessage), args.Get(1).(error)
}

func (m *MockMessageRepository) GetTransactions(txId string, txHash string) ([]message.TransactionMessage, error) {
	args := m.Called(txId, txHash)
	return args.Get(0).([]message.TransactionMessage), args.Get(1).(error)
}
