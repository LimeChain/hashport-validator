package transaction

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/stretchr/testify/mock"
)

type MockTransactionResponse struct {
	mock.Mock
}

func (mtr *MockTransactionResponse) GetReceipt(client *hedera.Client) (hedera.TransactionReceipt, error) {
	args := mtr.Called(client)
	if args.Get(1) == nil {
		return args.Get(0).(hedera.TransactionReceipt), nil
	}
	return args.Get(0).(hedera.TransactionReceipt), args.Get(1).(error)
}

func (mtr *MockTransactionResponse) GetRecord(client *hedera.Client) (hedera.TransactionRecord, error) {
	args := mtr.Called(client)
	if args.Get(1) == nil {
		return args.Get(0).(hedera.TransactionRecord), nil
	}
	return args.Get(0).(hedera.TransactionRecord), args.Get(1).(error)
}

func (mtr *MockTransactionResponse) GetTransactionID() hedera.TransactionID {
	args := mtr.Called()
	return args.Get(0).(hedera.TransactionID)
}
