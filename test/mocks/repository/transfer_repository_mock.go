package repository

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/stretchr/testify/mock"
)

type MockTransferRepository struct {
	mock.Mock
}

func (m *MockTransferRepository) GetByTransactionId(txId string) (*entity.Transfer, error) {
	panic("implement me")
}

func (m *MockTransferRepository) GetWithFee(txId string) (*entity.Transfer, error) {
	panic("implement me")
}

func (m *MockTransferRepository) GetWithPreloads(txId string) (*entity.Transfer, error) {
	panic("implement me")
}

func (m *MockTransferRepository) GetUnprocessedTransfers() ([]*entity.Transfer, error) {
	panic("implement me")
}

func (m *MockTransferRepository) Create(ct *transfer.Transfer) (*entity.Transfer, error) {
	args := m.Called(ct)
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transfer), nil
	}
	return nil, args.Get(1).(error)
}

func (m *MockTransferRepository) SaveRecoveredTxn(ct *transfer.Transfer) error {
	panic("implement me")
}

func (m *MockTransferRepository) UpdateStatusCompleted(txId string) error {
	args := m.Called(txId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockTransferRepository) UpdateStatusSignatureSubmitted(txId string) error {
	panic("implement me")
}

func (m *MockTransferRepository) UpdateStatusSignatureMined(txId string) error {
	panic("implement me")
}

func (m *MockTransferRepository) UpdateStatusSignatureFailed(txId string) error {
	panic("implement me")
}
