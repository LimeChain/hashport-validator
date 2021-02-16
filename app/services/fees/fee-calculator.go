package fees

import (
	"errors"
	"github.com/ethereum/go-ethereum/params"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/provider"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
)

type FeeCalculator struct {
	rateProvider  provider.ExchangeRateProvider
	configuration config.Hedera
}

func NewFeeCalculator(rateProvider provider.ExchangeRateProvider, configuration config.Hedera) *FeeCalculator {
	return &FeeCalculator{
		rateProvider:  rateProvider,
		configuration: configuration,
	}
}

func (fc FeeCalculator) ValidateExecutionFee(strTransferFee string, transferAmount string, gasPrice string) (bool, error) {
	bigTransferAmount, err := helper.ToBigInt(transferAmount)
	if err != nil {
		return false, errors.New(INVALID_TRANSFER_AMOUNT)
	}

	serviceFeePercent := new(big.Int).SetUint64(fc.configuration.Client.ServiceFeePercent)
	bigServiceFee := new(big.Int).Mul(bigTransferAmount, serviceFeePercent)
	bigServiceFee = new(big.Int).Div(bigServiceFee, new(big.Int).SetInt64(100))

	bigTxFee, err := helper.ToBigInt(strTransferFee)
	if err != nil {
		return false, errors.New(INVALID_TRANSFER_FEE)
	}

	estimatedFee := getFee(bigTxFee, bigServiceFee)

	if bigTransferAmount.Cmp(estimatedFee) < 0 {
		return false, errors.New(INSANE)
	}

	bigGasPrice, err := helper.ToBigInt(gasPrice)
	if err != nil {
		return false, errors.New(INVALID_GAS_PRICE)
	}

	exchangeRate, err := fc.rateProvider.GetEthVsHbarRate()
	if err != nil {
		return false, err
	}

	estimatedGas := new(big.Int).SetUint64(fc.getEstimatedGas())

	bigGasPrice = gweiToWei(bigGasPrice)

	weiTxFee := calculateWeiTxFee(bigGasPrice, estimatedGas)

	ethTxFee := weiToEther(weiTxFee)

	hbarTxFee := etherToHbar(ethTxFee, exchangeRate)

	tinyBarTxFee := hbarToTinyBar(hbarTxFee)

	decimalTxFee := decimal.NewFromBigInt(bigTxFee, 0)
	if tinyBarTxFee.Cmp(decimalTxFee) >= 0 {
		return false, errors.New(INSUFFICIENT_FEE)
	}

	return true, nil
}

func (fc FeeCalculator) getEstimatedGas() uint64 {
	majorityValidatorsCount := len(fc.configuration.Handler.ConsensusMessage.Addresses)/2 + 1
	estimatedGas := fc.configuration.Client.BaseGasUsage + uint64(majorityValidatorsCount)*fc.configuration.Client.GasPerValidator
	return estimatedGas
}

func calculateWeiTxFee(gasPrice *big.Int, estimatedGas *big.Int) *big.Int {
	return new(big.Int).Mul(gasPrice, estimatedGas)
}

func etherToHbar(ethTxFee decimal.Decimal, exchangeRate float64) decimal.Decimal {
	return ethTxFee.Div(decimal.NewFromFloat(exchangeRate))
}

func gweiToWei(gwei *big.Int) *big.Int {
	return new(big.Int).Mul(gwei, big.NewInt(params.GWei))
}

func weiToEther(wei *big.Int) decimal.Decimal {
	divisor := decimal.NewFromInt(params.Ether)
	ether := decimal.NewFromBigInt(wei, 0)
	ether = ether.Div(divisor)
	return ether
}

func hbarToTinyBar(hbar decimal.Decimal) decimal.Decimal {
	exp := math.Pow(10, 8)
	multiplier := decimal.NewFromFloat(exp)
	hbar = hbar.Mul(multiplier)
	return hbar
}

func getFee(transferFee *big.Int, serviceFee *big.Int) *big.Int {
	return new(big.Int).Add(transferFee, serviceFee)
}
