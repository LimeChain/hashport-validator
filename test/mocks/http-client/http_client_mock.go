package http_client

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	args := m.Called(url)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].(*http.Response), nil
	}
	return args[0].(*http.Response), args[1].(error)
}
