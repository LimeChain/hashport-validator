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

package transaction

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetByTransactionId(transactionId string) (*transaction.Transaction, error) {
	args := m.Called(transactionId)
	if args.Get(1) == nil {
		return args.Get(0).(*transaction.Transaction), nil
	}
	return args.Get(0).(*transaction.Transaction), args.Get(1).(error)
}

func (m *MockTransactionRepository) GetInitialAndSignatureSubmittedTx() ([]*transaction.Transaction, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).([]*transaction.Transaction), nil
	}
	return args.Get(0).([]*transaction.Transaction), args.Get(1).(error)
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
	if args.Get(0) == nil {
		return nil
	}
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
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransactionRepository) Skip(ct *proto.CryptoTransferMessage) error {
	args := m.Called(ct)
	return args.Get(0).(error)
}

func (m *MockTransactionRepository) GetSkippedTransactionsAndMessages() (map[string][]*message.TransactionMessage, error) {
	args := m.Called()
	return args.Get(0).(map[string][]*message.TransactionMessage), args.Get(1).(error)
}
