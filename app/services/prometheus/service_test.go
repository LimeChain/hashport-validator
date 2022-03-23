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

package prometheus

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	serviceInstance              *Service
	gauge                        prometheus.Gauge
	counter                      prometheus.Counter
	isMonitoringEnabled          = true
	gaugeOpts                    = prometheus.GaugeOpts{Name: "GaugeName", Help: "GaugeHelp"}
	gaugeSuffix                  = "gauge_suffix"
	counterOpts                  = prometheus.CounterOpts{Name: "CounterName", Help: "CounterHelp"}
	counterSuffix                = "counter_suffix"
	sourceNetworkId              = constants.HederaNetworkId
	sourceNetworkName            = testConstants.Networks[constants.HederaNetworkId].Name
	targetNetworkId              = testConstants.EthereumNetworkId
	targetNetworkName            = testConstants.Networks[testConstants.EthereumNetworkId].Name
	assetAddress                 = constants.Hbar
	transactionId                = "0.0.1234-1234-1234"
	transactionIdWithUnderscores = "0_0_1234_1234_1234"
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService(mocks.MAssetsService, isMonitoringEnabled)

	assert.Equal(t, serviceInstance, actualService)
}

func Test_CreateGaugeIfNotExists(t *testing.T) {
	setup()

	gauge = serviceInstance.CreateGaugeIfNotExists(gaugeOpts)
	defer serviceInstance.DeleteGauge(gaugeOpts.Name)

	assert.NotNil(t, gauge)
}

func Test_ConstructMetricName_Native(t *testing.T) {
	setup()

	mocks.MAssetsService.On("IsNative", sourceNetworkId, assetAddress).Return(true)

	expectedNative := fmt.Sprintf("%v_%v_to_%v_%v_%v", constants.Native, sourceNetworkName, targetNetworkName, transactionIdWithUnderscores, constants.MajorityReachedNameSuffix)
	actual, err := serviceInstance.ConstructMetricName(sourceNetworkId, targetNetworkId, assetAddress, transactionId, constants.MajorityReachedNameSuffix)

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedNative, actual)
}

func Test_ConstructMetricName_Wrapped(t *testing.T) {
	setup()

	mocks.MAssetsService.On("IsNative", targetNetworkId, assetAddress).Return(false)

	expectedNative := fmt.Sprintf("%v_%v_to_%v_%v_%v", constants.Wrapped, targetNetworkName, sourceNetworkName, transactionIdWithUnderscores, constants.MajorityReachedNameSuffix)
	actual, err := serviceInstance.ConstructMetricName(targetNetworkId, sourceNetworkId, assetAddress, transactionId, constants.MajorityReachedNameSuffix)

	assert.Equal(t, err, nil)
	assert.Equal(t, expectedNative, actual)
}

func Test_ConstructMetricName_ShouldThrow(t *testing.T) {
	setup()

	mocks.MAssetsService.On("IsNative", uint64(10), assetAddress).Return(false)

	_, err := serviceInstance.ConstructMetricName(10, sourceNetworkId, assetAddress, transactionId, constants.MajorityReachedNameSuffix)

	expectedError := fmt.Sprintf("Network id %v is missing in id to name mapping.", 10)
	assert.Errorf(t, err, expectedError)
}

func Test_CreateSuccessRateGaugeIfNotExists(t *testing.T) {
	setup()

	mocks.MAssetsService.On("IsNative", sourceNetworkId, assetAddress).Return(true)

	gauge, err := serviceInstance.CreateSuccessRateGaugeIfNotExists(
		transactionId,
		sourceNetworkId,
		targetNetworkId,
		assetAddress,
		gaugeSuffix,
		gaugeOpts.Help)

	fullGaugeName, err2 := serviceInstance.ConstructMetricName(sourceNetworkId, targetNetworkId, assetAddress, transactionId, gaugeSuffix)

	defer serviceInstance.DeleteGauge(fullGaugeName)

	assert.NotNil(t, gauge)
	assert.Nil(t, err)
	assert.Nil(t, err2)
}

func Test_GetGauge(t *testing.T) {
	setup()

	gauge = serviceInstance.CreateGaugeIfNotExists(gaugeOpts)
	defer serviceInstance.DeleteGauge(gaugeOpts.Name)
	gaugeInMapping := serviceInstance.GetGauge(gaugeOpts.Name)

	assert.NotNil(t, gaugeInMapping)
}

func Test_DeleteGauge(t *testing.T) {
	setup()

	gauge = serviceInstance.CreateGaugeIfNotExists(gaugeOpts)
	serviceInstance.DeleteGauge(gaugeOpts.Name)

	gaugeInMapping := serviceInstance.GetGauge(gaugeOpts.Name)

	assert.Nil(t, gaugeInMapping)
}

func Test_CreateCounterIfNotExists(t *testing.T) {
	setup()

	counter = serviceInstance.CreateCounterIfNotExists(counterOpts)
	defer serviceInstance.DeleteCounter(counterOpts.Name)

	assert.NotNil(t, counter)
}

func Test_GetCounter(t *testing.T) {
	setup()

	counter = serviceInstance.CreateCounterIfNotExists(counterOpts)
	defer serviceInstance.DeleteCounter(counterOpts.Name)
	counterInMapping := serviceInstance.GetCounter(counterOpts.Name)

	assert.NotNil(t, counterInMapping)
}

func Test_DeleteCounter(t *testing.T) {
	setup()

	counter = serviceInstance.CreateCounterIfNotExists(counterOpts)
	serviceInstance.DeleteCounter(counterOpts.Name)

	counterInMapping := serviceInstance.GetCounter(counterOpts.Name)

	assert.Nil(t, counterInMapping)
}

func setup() {
	mocks.Setup()
	helper.SetupNetworks()

	serviceInstance = &Service{
		logger:              config.GetLoggerFor("Prometheus Service"),
		gauges:              map[string]prometheus.Gauge{},
		counters:            map[string]prometheus.Counter{},
		assetsService:       mocks.MAssetsService,
		isMonitoringEnabled: isMonitoringEnabled,
	}
}
