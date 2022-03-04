package asset

import "github.com/shopspring/decimal"

type NativeAsset struct {
	MinFeeAmountInUsd *decimal.Decimal
	ChainId           uint64
	Asset             string
	FeePercentage     int64
}

type FungibleAssetInfo struct {
	Name     string
	Symbol   string
	Decimals uint8
}
