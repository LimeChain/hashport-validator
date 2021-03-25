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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/stretchr/testify/mock"
)

type MockTransferRepository struct {
	mock.Mock
}

func (m *MockTransferRepository) Create(ct *proto.TransferMessage) (*entity.Transfer, error) {
	args := m.Called(ct)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return args.Get(0).(*entity.Transfer), args.Get(1).(error)
}

func (m *MockTransferRepository) Save(tx *entity.Transfer) error {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusSignatureMined(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateEthTxMined(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusEthTxMsgSubmitted(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusEthTxMsgMined(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusEthTxMsgFailed(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) GetByTransactionId(transactionId string) (*entity.Transfer, error) {
	args := m.Called(transactionId)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return args.Get(0).(*entity.Transfer), args.Get(1).(error)
}

func (m *MockTransferRepository) GetInitialAndSignatureSubmittedTx() ([]*entity.Transfer, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*entity.Transfer), nil
	}
	return args.Get(0).([]*entity.Transfer), args.Get(1).(error)
}

func (m *MockTransferRepository) UpdateStatusInsufficientFee(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusSignatureFailed(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateEthTxSubmitted(txId string, hash string) error {
	args := m.Called(txId, hash)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateEthTxReverted(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) SaveRecoveredTxn(ct *proto.TransferMessage) error {
	args := m.Called(ct)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) GetUnprocessedTransfers() ([]entity.Transfer, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]entity.Transfer), nil
	}
	return args.Get(0).([]entity.Transfer), args.Get(1).(error)
}
