package service

import (
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/stretchr/testify/mock"
)

type MockReadOnlyService struct {
	mock.Mock
}

func (m *MockReadOnlyService) FindTransfer(transferID string, fetch func() (*mirror_node.Response, error), save func(transactionID, scheduleID, status string) error) {
	m.Called(transferID, fetch, save)
}
