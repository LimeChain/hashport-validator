package pricing

import (
	"github.com/shopspring/decimal"
	"math/big"
)

type TokenPriceInfo struct {
	UsdPrice         decimal.Decimal
	MinAmountWithFee *big.Int
}
