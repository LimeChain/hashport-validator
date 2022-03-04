package client

import "github.com/shopspring/decimal"

type HederaWebAPI interface {
	GetHBARUsdPrice() (price decimal.Decimal, err error)
}
