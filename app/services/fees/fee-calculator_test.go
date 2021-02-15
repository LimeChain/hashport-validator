package fees

import (
	exchangerate "github.com/limechain/hedera-eth-bridge-validator/app/clients/exchange-rate"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testConfigAddress = "../../../config/application.yml"
)

var (
	provider      = exchangerate.NewExchangeRateProvider("hedera-hashgraph", "eth") // TODO: Mock returning value
	hederaConfig  = config.LoadConfigTest(testConfigAddress).Hedera
	feeCalculator = NewFeeCalculator(provider, hederaConfig)
)

func TestFeeCalculatorHappyPath(t *testing.T) {
	valid, err := feeCalculator.ValidateExecutionFee("2500", "100000", "2")
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestFeeCalculatorSanityCheckWorks(t *testing.T) {
	valid, err := feeCalculator.ValidateExecutionFee("270000000000", "10000", "1")
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInsufficientFee(t *testing.T) {
	valid, err := feeCalculator.ValidateExecutionFee("270000000", "1000000000000", "1")
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferFee(t *testing.T) {
	valid, err := feeCalculator.ValidateExecutionFee("someinvalidvalue", "1000000000000", "1")
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidGasPrice(t *testing.T) {
	valid, err := feeCalculator.ValidateExecutionFee("27000000", "1000000000000", "someinvalidvalue")
	assert.NotNil(t, err)
	assert.False(t, valid)
}
