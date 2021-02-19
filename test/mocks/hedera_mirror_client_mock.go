package mocks

import (
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/stretchr/testify/mock"
)

type MockHederaMirrorClient struct {
	mock.Mock
}

func (m *MockHederaMirrorClient) GetSuccessfulAccountCreditTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp int64) (*transaction.HederaTransactions, error) {
	args := m.Called(accountId, milestoneTimestamp)

	if args.Get(1) == nil {
		return args.Get(0).(*transaction.HederaTransactions), nil
	}
	return args.Get(0).(*transaction.HederaTransactions), args.Get(1).(error)
}

func (m *MockHederaMirrorClient) GetAccountTransaction(transactionID string) (*transaction.HederaTransactions, error) {
	args := m.Called(transactionID)

	if args.Get(1) == nil {
		return args.Get(0).(*transaction.HederaTransactions), nil
	}
	return args.Get(0).(*transaction.HederaTransactions), args.Get(1).(error)
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
