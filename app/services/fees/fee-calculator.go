package fees

import (
	"github.com/ethereum/go-ethereum/params"
	exchangerate "github.com/limechain/hedera-eth-bridge-validator/app/clients/exchange-rate"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"math/big"
)

func getFee(transferFee *big.Int, serviceFee *big.Int) (*big.Int, error) {
	return new(big.Int).Add(transferFee, serviceFee), nil
}

func ValidateExecutionFee(strTransferFee string, transferAmount uint64, exchangeRateProvider *exchangerate.ExchangeRateProvider, gasPrice string) (bool, error) {
	hederaConfiguration := config.LoadConfig().Hedera

	serviceFeePercent := hederaConfiguration.Client.ServiceFeePercent

	// Sanity Check
	bigTransferAmount := new(big.Int).SetUint64(transferAmount)
	bigServiceFee := new(big.Int).SetUint64(transferAmount * serviceFeePercent / 100)

	// Transaction Fee to big.Int type
	bigTxFee, err := helper.ToBigInt(strTransferFee)
	if err != nil {
		return false, err
	}

	// Get the estimated fee
	estimatedFee, err := getFee(bigTxFee, bigServiceFee)
	if err != nil {
		return false, err
	}

	// Report Invalid Transfer Amount
	if bigTransferAmount.Cmp(estimatedFee) < 0 {
		return false, nil
	}

	// Get Gas Price from Memo
	bigGasPrice, err := helper.ToBigInt(gasPrice)
	if err != nil {
		return false, err
	}

	// Estimate Gas
	majorityValidatorsCount := len(hederaConfiguration.Handler.ConsensusMessage.Addresses)/2 + 1
	estimatedGas := hederaConfiguration.Client.BaseGasUsage + uint64(majorityValidatorsCount)*hederaConfiguration.Client.GasPerValidator

	// Get Exchange Rate from External APIs (CoinGecko, etc.)
	exchangeRate, err := exchangeRateProvider.GetEthVsHbarRate()
	if err != nil {
		return false, err
	}

	// Convert GWei to Wei
	bigGasPrice = new(big.Int).Mul(bigGasPrice, big.NewInt(params.GWei))

	// Calculate HBar Transaction Fee
	hBarTxFee := new(big.Float).SetFloat64(float64(bigGasPrice.Uint64()*estimatedGas) / exchangeRate)
	txFeeFloat := new(big.Float).SetInt(bigTxFee)

	// Report Insufficiency
	if txFeeFloat.Cmp(hBarTxFee) < 0 {
		return false, err
	}

	return true, nil
}
