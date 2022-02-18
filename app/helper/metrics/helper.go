/*
 * Copyright 2022 LimeChain Ltd.
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

package metrics

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
	"strings"
)

func PrepareIdForPrometheus(id string) string {
	for symbolToReplace, repetitions := range constants.PrometheusNotAllowedSymbolsWithRepetitions {
		id = strings.Replace(id, symbolToReplace, constants.NotAllowedSymbolsReplacement, repetitions)
	}

	return id
}

func ConstructNameForMetric(sourceNetworkId, targetNetworkId uint64, tokenType, transactionId, metricTarget string) (string, error) {
	errMsg := "Network id %v is missing in id to name mapping."
	sourceNetworkName, exist := constants.NetworksById[sourceNetworkId]
	if !exist {
		return "", errors.New(fmt.Sprintf(errMsg, sourceNetworkId))
	}
	targetNetworkName, exist := constants.NetworksById[targetNetworkId]
	if !exist {
		return "", errors.New(fmt.Sprintf(errMsg, targetNetworkId))
	}

	transactionId = PrepareIdForPrometheus(transactionId)

	return fmt.Sprintf("%s_%s_to_%s_%s_%s", tokenType, sourceNetworkName, targetNetworkName, transactionId, metricTarget), nil
}

// Success Rate Metrics //

func CreateUserGetHisTokensIfNotExists(sourceChainId, targetChainId uint64, asset string, transferID string, prometheusService service.Prometheus, logger *log.Entry) prometheus.Gauge {
	if !prometheusService.GetIsMonitoringEnabled() {
		return nil
	}

	gauge, err := prometheusService.CreateSuccessRateGaugeIfNotExists(
		transferID,
		sourceChainId,
		targetChainId,
		asset,
		constants.UserGetHisTokensNameSuffix,
		constants.UserGetHisTokensHelp)

	if err != nil {
		logger.Errorf("[%s] - Failed to create gauge metric for [%s]. Error: [%s]", transferID, constants.UserGetHisTokensNameSuffix, err)
	}
	return gauge
}

func SetUserGetHisTokens(sourceChainId, targetChainId uint64, asset string, transferID string, prometheusService service.Prometheus, logger *log.Entry) {
	if !prometheusService.GetIsMonitoringEnabled() {
		return
	}
	gauge := CreateUserGetHisTokensIfNotExists(sourceChainId, targetChainId, asset, transferID, prometheusService, logger)

	logger.Infof("[%s] - Setting value to 1.0 for metric [%v]", transferID, constants.UserGetHisTokensNameSuffix)
	gauge.Set(1.0)
}

func CreateFeeTransferredIfNotExists(sourceChainId, targetChainId uint64, asset string, transferID string, prometheusService service.Prometheus, logger *log.Entry) prometheus.Gauge {
	if !prometheusService.GetIsMonitoringEnabled() {
		return nil
	}

	gauge, err := prometheusService.CreateSuccessRateGaugeIfNotExists(
		transferID,
		sourceChainId,
		targetChainId,
		asset,
		constants.FeeTransferredNameSuffix,
		constants.FeeTransferredHelp)

	if err != nil {
		logger.Errorf("[%s] - Failed to create gauge metric for [%s]. Error: [%s]", transferID, constants.FeeTransferredNameSuffix, err)
	}
	return gauge
}

func SetFeeTransferred(sourceChainId, targetChainId uint64, asset string, transferID string, prometheusService service.Prometheus, logger *log.Entry) {
	if !prometheusService.GetIsMonitoringEnabled() {
		return
	}
	gauge := CreateFeeTransferredIfNotExists(sourceChainId, targetChainId, asset, transferID, prometheusService, logger)

	logger.Infof("[%s] - Setting value to 1.0 for metric [%v]", transferID, constants.FeeTransferredNameSuffix)
	gauge.Set(1.0)
}

func CreateMajorityReachedIfNotExists(sourceChainId uint64, targetChainId uint64, asset string, transferID string, prometheusService service.Prometheus, logger *log.Entry) prometheus.Gauge {
	if !prometheusService.GetIsMonitoringEnabled() {
		return nil
	}

	gauge, err := prometheusService.CreateSuccessRateGaugeIfNotExists(
		transferID,
		sourceChainId,
		targetChainId,
		asset,
		constants.MajorityReachedNameSuffix,
		constants.MajorityReachedHelp)

	if err != nil {
		logger.Errorf("[%s] - Failed to create gauge metric for [%s]. Error: [%s]", transferID, constants.MajorityReachedNameSuffix, err)
	}

	return gauge
}

func SetMajorityReached(sourceChainId, targetChainId uint64, asset string, transferID string, prometheusService service.Prometheus, logger *log.Entry) {
	if !prometheusService.GetIsMonitoringEnabled() {
		return
	}
	gauge := CreateMajorityReachedIfNotExists(sourceChainId, targetChainId, asset, transferID, prometheusService, logger)

	logger.Infof("[%s] - Setting value to 1.0 for metric [%v]", transferID, constants.MajorityReachedNameSuffix)
	gauge.Set(1.0)
}

func AssetAddressToMetricName(assetAddress string) string {
	replace := PrepareIdForPrometheus(assetAddress)
	result := fmt.Sprintf("%s%s", constants.AssetMetricsNamePrefix, replace)
	return result
}

func ConvertToHbar(amount int) float64 {
	hbar := hedera.HbarFromTinybar(int64(amount))
	return hbar.As(hedera.HbarUnits.Hbar)
}

func ConvertBasedOnDecimal(value *big.Int, decimal uint8) (*float64, error) {
	if decimal < 1 {
		return nil, errors.New(fmt.Sprintf(`Failed to calc with decimal: [%d].`, decimal))
	}
	parseValue, e := strconv.ParseFloat(constants.CreateDecimalPrefix+strings.Repeat(constants.CreateDecimalRepeat, int(decimal)), 64)
	if e != nil {
		return nil, e
	}
	res, _ := new(big.Float).Set(new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(parseValue))).Float64()
	return &res, nil
}
