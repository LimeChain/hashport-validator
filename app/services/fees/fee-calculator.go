package fees

import (
	"errors"
	"fmt"
	"math/big"
)

func getFee() (*big.Int, error) {
	return new(big.Int), nil
}

func ValidateExecutionFee(strTransferFee string) (bool, error) {
	transferFee := new(big.Int)
	transferFee, ok := transferFee.SetString(strTransferFee, 10)
	if !ok {
		return false, errors.New(fmt.Sprintf("Failed to parse fee: [%s]", strTransferFee))
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
