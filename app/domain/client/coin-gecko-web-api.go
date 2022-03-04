package client

import "github.com/shopspring/decimal"

type CoinGeckoWebAPI interface {
	GetUsdPrices(idsByNetworkAndAddress map[uint64]map[string]string) (pricesByNetworkAndAddress map[uint64]map[string]decimal.Decimal, err error)
}
