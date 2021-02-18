/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fees

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	exchangeRate  = 0.00007
	validGasPrice = "130"
	smallGasPrice = "1"

	invalidValue = "someinvalidvalue"

	transferAmount        = "100000000000"
	invalidTransferAmount = "100"

	transferFee         = "60000000000"
	tooSmallTransferFee = "2"
	tooBigTransferFee   = "100000000001"
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
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestFeeCalculatorSanityCheckWorks(t *testing.T) {
	mocks.Setup()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(tooBigTransferFee, invalidTransferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNotCalled(t, "GetEthVsHbarRate")
	assert.NotNil(t, err)
	assert.Equal(t, InsufficientFee, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInsufficientFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(tooSmallTransferFee, transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.NotNil(t, err)
	assert.Equal(t, InsufficientFee, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferFee(t *testing.T) {
	mocks.Setup()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(invalidValue, transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNotCalled(t, "GetEthVsHbarRate")
	assert.NotNil(t, err)
	assert.Equal(t, InvalidTransferFee, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidGasPrice(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, invalidValue)
	mocks.MExchangeRateProvider.AssertNotCalled(t, "GetEthVsHbarRate")
	assert.NotNil(t, err)
	assert.Equal(t, InvalidGasPrice, err)
	assert.False(t, valid)
}

func TestFeeCalculatorFailsWithInvalidTransferAmount(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, invalidValue, validGasPrice)
	mocks.MExchangeRateProvider.AssertNotCalled(t, "GetEthVsHbarRate")
	assert.NotNil(t, err)
	assert.Equal(t, InvalidTransferAmount, err)
	assert.False(t, valid)
}

func TestFeeCalculatorWithInvalidRateProvider(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(float64(0), RateProviderFailure)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.NotNil(t, err)
	assert.Equal(t, RateProviderFailure, err)
	assert.False(t, valid)
}

func TestFeeCalculatorWithManyValidators(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)
	config := validHederaConfig()

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, config)

	valid, err := feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.Nil(t, err)
	assert.True(t, valid)

	config = addMoreValidatorsTo(config, 7)
	feeCalculator = NewFeeCalculator(mocks.MExchangeRateProvider, config)

	valid, err = feeCalculator.ValidateExecutionFee(transferFee, transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 2)
	assert.NotNil(t, err)
	assert.Equal(t, InsufficientFee, err)
	assert.False(t, valid)
}

func TestFeeCalculatorWithZeroServiceFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)
	config := validHederaConfig()
	config.Client.ServiceFeePercent = 0

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, config)

	valid, err := feeCalculator.ValidateExecutionFee(transferAmount, transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestFeeCalculatorConsidersServiceFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)
	config := validHederaConfig()
	config.Client.BaseGasUsage = 1
	config.Client.GasPerValidator = 1
	config.Client.ServiceFeePercent = 0

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, config)

	// Based on the mocked information above + exchange rate. This is the lowest possible transaction cost provided
	lowerEnd := "4286"

	valid, err := feeCalculator.ValidateExecutionFee(lowerEnd, transferAmount, smallGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.Nil(t, err)
	assert.True(t, valid)

	insufficientFee := "4285"

	valid, err = feeCalculator.ValidateExecutionFee(insufficientFee, transferAmount, smallGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 2)
	assert.NotNil(t, err)
	assert.Equal(t, InsufficientFee, err)
	assert.False(t, valid)
}

func TestFeeCalculatorWithZeroTransferFee(t *testing.T) {
	mocks.Setup()
	mocks.MExchangeRateProvider.On("GetEthVsHbarRate").Return(exchangeRate, nil)

	feeCalculator := NewFeeCalculator(mocks.MExchangeRateProvider, validHederaConfig())

	valid, err := feeCalculator.ValidateExecutionFee("0", transferAmount, validGasPrice)
	mocks.MExchangeRateProvider.AssertNumberOfCalls(t, "GetEthVsHbarRate", 1)
	assert.NotNil(t, err)
	assert.Equal(t, InsufficientFee, err)
	assert.False(t, valid)
}
