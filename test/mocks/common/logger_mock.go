package common

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
	Entries []string
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
	m.Entries = append(m.Entries, fmt.Sprintf(format, args...))
}
