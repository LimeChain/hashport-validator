package mocks

var MExchangeRateProvider *MockExchangeRateProvider

func Setup() {
	MExchangeRateProvider = &MockExchangeRateProvider{}
}
