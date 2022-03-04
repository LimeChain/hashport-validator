package constants

import "math/big"

const (
	FeeMaxPercentage = 100000
	FeeMinPercentage = 0
)

var (
	FeeMaxPercentageBigInt = big.NewInt(FeeMaxPercentage)
)
