package fees

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	exchangeRate  = 0.0000764
	validGasPrice = "130"

	invalidValue = "someinvalidvalue"

	transferAmount        = "1000000000000"
	invalidTransferAmount = "100"

	transferFee         = "60000000000"
	tooSmallTransferFee = "2"
	tooBigTransferFee   = transferAmount
)

func validHederaConfig() config.Hedera {
	hederaConfig := config.Hedera{}
	hederaConfig.Client.ServiceFeePercent = 10
	hederaConfig.Client.BaseGasUsage = 130000
	hederaConfig.Client.GasPerValidator = 54000
	hederaConfig.Handler.ConsensusMessage.Addresses = []string{
		"0xsomeaddress",
		"0xsomeaddress2",
		"0xsomeaddress3",
	}
	return hederaConfig
}

func addMoreValidatorsTo(config config.Hedera, additional uint) config.Hedera {
	for additional != 0 {
		config.Handler.ConsensusMessage.Addresses =
			append(config.Handler.ConsensusMessage.Addresses,
				fmt.Sprintf("0xsomeaddress%d", len(config.Handler.ConsensusMessage.Addresses)+1))
		additional--
	}

	return config
}

func TestFeeCalculatorHappyPath(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestFeeCalculatorSanityCheckWorks(t *testing.T) {
	mocks.Setup()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(tooBigTransferFee, invalidTransferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InsufficientFee)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInsufficientFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(tooSmallTransferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InsufficientFee)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferFee(t *testing.T) {
	mocks.Setup()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(invalidValue, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InvalidTransferFee)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidGasPrice(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, invalidValue)
	assert.NotNil(t, err)
	assert.Equal(t, err, InvalidGasPrice)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferAmount(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, invalidValue, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InvalidTransferAmount)
	assert.False(t, valid)
}

func TestFeeCalculatorWithInvalidRateProvider(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(float64(0), RateProviderFailure)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, RateProviderFailure)
	assert.False(t, valid)
}

func TestFeeCalculatorWithManyValidators(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)
	config := validHederaConfig()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, config)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.Nil(t, err)
	assert.True(t, valid)

	config = addMoreValidatorsTo(config, 7)
	feeCalculator = NewFeeCalculator(mocks.MExchangeRateProvider, config)

	valid, err = feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InsufficientFee)
	assert.False(t, valid)
}

func TestFeeCalculatorWithZeroServiceFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)
	config := validHederaConfig()
	config.Client.ServiceFeePercent = 0

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, config)

	valid, err := feeCalculator.ValidateExecutionFee(tooBigTransferFee, transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InsufficientFee)
	assert.False(t, valid)
}

func TestFeeCalculatorWithZeroTransferFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee("0", transferAmount, validGasPrice)
	assert.NotNil(t, err)
	assert.Equal(t, err, InsufficientFee)
	assert.False(t, valid)
}
