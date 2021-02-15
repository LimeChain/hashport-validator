package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetByTransactionId(transactionId string) (*transaction.Transaction, error) {
	args := m.Called(transactionId)
	return args.Get(0).(*transaction.Transaction), args.Get(0).(error)
}

func (m *MockTransactionRepository) GetInitialAndSignatureSubmittedTx() ([]*transaction.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]*transaction.Transaction), args.Get(0).(error)

}

func (m *MockTransactionRepository) Create(ct *proto.CryptoTransferMessage) error {
	args := m.Called(ct)
	return args.Get(0).(error)
}

func (m *MockTransactionRepository) UpdateStatusCompleted(txId string) error {
	args := m.Called(txId)
	return args.Get(0).(error)

}

func (m *MockTransactionRepository) UpdateStatusInsufficientFee(txId string) error {
	args := m.Called(txId)
	return args.Get(0).(error)

}

func (m *MockTransactionRepository) UpdateStatusSignatureProvided(txId string) error {
	args := m.Called(txId)
	return args.Get(0).(error)

}

func (m *MockTransactionRepository) UpdateStatusSignatureFailed(txId string) error {
	args := m.Called(txId)
	return args.Get(0).(error)

}

func (m *MockTransactionRepository) UpdateStatusEthTxSubmitted(txId string, hash string) error {
	args := m.Called(txId, hash)
	return args.Get(0).(error)

}

func (m *MockTransactionRepository) UpdateStatusEthTxReverted(txId string) error {
	args := m.Called(txId)
	return args.Get(0).(error)

}

func (m *MockTransactionRepository) UpdateStatusSignatureSubmitted(txId string, submissionTxId string, signature string) error {
	args := m.Called(txId, submissionTxId, signature)
	return args.Get(0).(error)
}
