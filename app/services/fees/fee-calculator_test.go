package fees

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testConfigAddress = "../../../config/application.yml"
	exchangeRate      = 0.0000764
)

var (
	hederaConfig          = config.LoadConfigTest(testConfigAddress).Hedera
	transferAmount        = "10000000000000000"
	invalidTransferAmount = "100"
)

func TestFeeCalculatorHappyPath(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)

	valid, err := feeCalculator.ValidateExecutionFee("600000000", transferAmount, "2")
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestFeeCalculatorSanityCheckWorks(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee("2700", invalidTransferAmount, "1")
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInsufficientFee(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee("2700", transferAmount, "1")
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferFee(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee("someinvalidvalue", "1000000000000", "1")
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidGasPrice(t *testing.T) {
	mocks.Setup()
	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, hederaConfig)
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	valid, err := feeCalculator.ValidateExecutionFee("27000000", "1000000000000", "someinvalidvalue")
	assert.NotNil(t, err)
	assert.False(t, valid)
}
