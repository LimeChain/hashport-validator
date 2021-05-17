package database

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) GetConnection() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDatabase) ConnectWithMigration(config config.Database) *gorm.DB {
	args := m.Called(config)
	return args.Get(0).(*gorm.DB)
}
