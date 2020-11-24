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

	estimatedFee, err := getFee()
	if err != nil {
		return false, err
	}

	if transferFee.Cmp(estimatedFee) >= 0 {
		return true, nil
	}

	return false, nil
}
