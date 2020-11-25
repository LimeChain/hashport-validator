package helper

import (
	"errors"
	"fmt"
	"math/big"
)

func ToBigInt(value string) (*big.Int, error) {
	amount := new(big.Int)
	amount, ok := amount.SetString(value, 10)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Failed to parse amount [%s] to big integer.", amount))
	}

	return amount, nil
}
