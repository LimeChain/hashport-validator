package fees

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testConfigAddress = "../../../config/application.yml"
	exchangeRate      = 0.0000764
	validGasPrice     = "130"

	invalidValue = "someinvalidvalue"

	transferAmount        = "1000000000000"
	invalidTransferAmount = "100"

	transferFee         = "60000000000"
	tooSmallTransferFee = "2"
	tooBigTransferFee   = transferAmount
)

var (
	hederaConfig = config.LoadConfigTest(testConfigAddress).Hedera
)

func TestFeeCalculatorHappyPath(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestFeeCalculatorSanityCheckWorks(t *testing.T) {
	mocks.Setup()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(tooBigTransferFee, invalidTransferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, Insane)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInsufficientFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(tooSmallTransferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InsufficientFee)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferFee(t *testing.T) {
	mocks.Setup()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(invalidValue, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InvalidTransferFee)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidGasPrice(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, invalidValue)
	assert.NotNil(t, err)
	assert.Equal(t, err, InvalidGasPrice)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferAmount(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, invalidValue, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InvalidTransferAmount)
	assert.False(t, valid)
}

func TestFeeCalculatorWithInvalidRateProvider(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(float64(0), RateProviderFailure)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, RateProviderFailure)
	assert.False(t, valid)
}
