package provider

type ExchangeRateProvider interface {
	GetEthVsHbarRate() (float64, error)
}
