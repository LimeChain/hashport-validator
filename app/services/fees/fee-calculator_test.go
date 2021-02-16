package fees

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testConfigAddress     = "../../../config/application.yml"
	exchangeRate          = 0.0000764
	validGasPrice         = "2"
	invalidValue          = "someinvalidvalue"
	transferFee           = "600000000"
	tooSmallTransferFee   = "2700"
	transferAmount        = "10000000000000000"
	invalidTransferAmount = "100"
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

	valid, err := feeCalculator.ValidateExecutionFee(tooSmallTransferFee, invalidTransferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), INSANE)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInsufficientFee(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee(tooSmallTransferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), INSUFFICIENT_FEE)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferFee(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee(invalidValue, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), INVALID_TRANSFER_FEE)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidGasPrice(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, invalidValue)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), INVALID_GAS_PRICE)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferAmount(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, invalidValue, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), INVALID_TRANSFER_AMOUNT)
	assert.False(t, valid)
}

func TestFeeCalculatorWithInvalidRateProvider(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(float64(0), errors.New("This error should be returned"))

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), RATE_PROVIDER_FAILURE)
	assert.False(t, valid)
}
