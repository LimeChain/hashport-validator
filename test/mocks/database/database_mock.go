package database

import (
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
