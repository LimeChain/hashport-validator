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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var precision = new(big.Int).SetInt64(100000)

type Calculator struct {
	rateProvider  clients.ExchangeRate
	configuration config.Hedera
	bridge        bridge.ContractService
}

func NewCalculator(rateProvider clients.ExchangeRate, configuration config.Hedera, bridge bridge.ContractService) *Calculator {
	return &Calculator{
		rateProvider:  rateProvider,
		configuration: configuration,
		bridge:        bridge,
	}
}

func (fc Calculator) ValidateExecutionFee(transferFee string, transferAmount string, gasPriceGwei string) (bool, error) {
	bigTransferAmount, err := helper.ToBigInt(transferAmount)
	if err != nil {
		return false, InvalidTransferAmount
	}

	bigTxFee, err := helper.ToBigInt(transferFee)
	if err != nil {
		return false, InvalidTransferFee
	}

	// Value of the serviceFeePercent in percentage. Range 0% to 99.999% multiplied my 1000
	bigServiceFee := new(big.Int).Mul(new(big.Int).Sub(bigTransferAmount, bigTxFee), fc.bridge.GetServiceFee())
	bigServiceFee = new(big.Int).Div(bigServiceFee, precision)
	estimatedFee := getFee(bigTxFee, bigServiceFee)

	if bigTransferAmount.Cmp(estimatedFee) < 0 {
		return false, InsufficientFee
	}

	bigGasPriceGWei, err := helper.ToBigInt(gasPriceGwei)
	if err != nil {
		return false, InvalidGasPrice
	}

	tinyBarTxFee, err := fc.getEstimatedTxFee(bigGasPriceGWei)
	if err != nil {
		return false, err
	}

	floatTxFee := new(big.Float).SetInt(estimatedFee)

	if tinyBarTxFee.Cmp(floatTxFee) >= 0 {
		return false, InsufficientFee
	}

	return true, nil
}

func (fc Calculator) GetEstimatedTxFee(gasPriceGwei string) (string, error) {
	bigGasPriceGWei, err := helper.ToBigInt(gasPriceGwei)
	if err != nil {
		return "", InvalidGasPrice
	}

	bigEstimatedTxFee, err := fc.getEstimatedTxFee(bigGasPriceGWei)
	if err != nil {
		return "", err
	}

	return bigEstimatedTxFee.String(), nil
}

func (fc Calculator) getEstimatedTxFee(gasPriceGwei *big.Int) (*big.Float, error) {
	exchangeRate, err := fc.rateProvider.GetEthVsHbarRate()
	if err != nil {
		return nil, err
	}

	estimatedGas := new(big.Int).SetUint64(fc.getEstimatedGas())
	bigGasPriceWei := gweiToWei(gasPriceGwei)
	weiTxFee := calculateWeiTxFee(bigGasPriceWei, estimatedGas)

	return weiToTinyBar(weiTxFee, exchangeRate), nil
}

func weiToTinyBar(weiTxFee *big.Int, exchangeRate float64) *big.Float {
	bigExchangeRate := big.NewFloat(exchangeRate)
	multiplicationRatio := big.NewFloat(1e-10)
	bigWeiTxFee := new(big.Float).SetInt(weiTxFee)
	ratioTxFee := new(big.Float).Mul(bigWeiTxFee, multiplicationRatio)
	return new(big.Float).Quo(ratioTxFee, bigExchangeRate)
}

func (fc Calculator) getEstimatedGas() uint64 {
	majorityValidatorsCount := len(fc.bridge.GetMembers())/2 + 1
	estimatedGas := fc.configuration.Client.BaseGasUsage + uint64(majorityValidatorsCount)*fc.configuration.Client.GasPerValidator
	return estimatedGas
}

func calculateWeiTxFee(gasPrice *big.Int, estimatedGas *big.Int) *big.Int {
	return new(big.Int).Mul(gasPrice, estimatedGas)
}

func gweiToWei(gwei *big.Int) *big.Int {
	return new(big.Int).Mul(gwei, big.NewInt(params.GWei))
}

func getFee(transferFee *big.Int, serviceFee *big.Int) *big.Int {
	return new(big.Int).Add(transferFee, serviceFee)
}
