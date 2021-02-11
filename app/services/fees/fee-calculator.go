package fees

import (
	exchangerate "github.com/limechain/hedera-eth-bridge-validator/app/clients/exchange-rate"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"math/big"
)

func getFee(transferFee *big.Int, serviceFee *big.Int) (*big.Int, error) {
	return new(big.Int).Add(transferFee, serviceFee), nil
}

func ValidateExecutionFee(strTransferFee string, serviceFee uint64, transferAmount uint64, exchangeRateProvider *exchangerate.ExchangeRateProvider, gasPrice string) (bool, error) {
	exchangeRate, err := exchangeRateProvider.GetEthVsHbarRate()
	if err != nil {
		return false, err
	}

	HBarTxFee := float64(gasPrice*estimatedGas) / exchangeRate // TODO: convert from gwei to wei, because it comes as gwei in the first place
	if HBarTxFee >= TxFee {
		return false, err
	}

	transferFee, err := helper.ToBigInt(strTransferFee)
	if err != nil {
		return false, err
	}

	bigServiceFee := new(big.Int).SetUint64(serviceFee)
	bigTransferAmount := new(big.Int).SetUint64(transferAmount)

	estimatedFee, err := getFee(transferFee, bigServiceFee)
	if err != nil {
		return false, err
	}

	if bigTransferAmount.Cmp(estimatedFee) >= 0 {
		return true, nil
	}

	return false, nil
}
