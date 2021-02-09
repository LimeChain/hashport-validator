package fees

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"math/big"
)

func getFee() (*big.Int, error) {

	return new(big.Int), nil
}

func ValidateExecutionFee(strTransferFee string) (bool, error) {
	transferFee, err := helper.ToBigInt(strTransferFee)
	if err != nil {
		return false, err
	}

	serviceFee, err := retrieveServiceFee()
	if err != nil {
		return false, err
	}

	estimatedFee, err := getFee()
	if err != nil {
		return false, err
	}

	estimation := estimatedFee.Mul(estimatedFee, serviceFee)
	estimation = estimation.Div(estimation, new(big.Int).SetInt64(100))

	if transferFee.Cmp(estimation) >= 0 {
		return true, nil
	}

	return false, nil
}

func retrieveServiceFee() (*big.Int, error) {
	return new(big.Int).SetInt64(10), nil
}
