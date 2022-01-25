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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
)

type MockPrometheusService struct {
	mock.Mock
}

// CreateAndRegisterGaugeMetric creates new Gauge Metric and registers it in Prometheus
func (mps *MockPrometheusService) CreateAndRegisterGaugeMetric(name string, help string) prometheus.Gauge {
	args := mps.Called(name, help)
	result := args.Get(0).(prometheus.Gauge)
	return result
}

// GetGauge retrieves Gauge by name with flag for existence
func (mps *MockPrometheusService) GetGauge(name string) prometheus.Gauge {
	args := mps.Called(name)
	result := args.Get(0).(prometheus.Gauge)
	return result
}

// CreateAndRegisterCounterMetric creates new Counter Metric and registers it in Prometheus
func (mps *MockPrometheusService) CreateAndRegisterCounterMetric(name string, help string) prometheus.Counter {
	args := mps.Called(name, help)
	result := args.Get(0).(prometheus.Counter)
	return result
}

// GetCounter retrieves Counter by name with flag for existence
func (mps *MockPrometheusService) GetCounter(name string) prometheus.Counter {
	args := mps.Called(name)
	result := args.Get(0).(prometheus.Counter)
	return result
}

// ConstructNameForSuccessRateMetric constructing name for success rate metric
func (mps *MockPrometheusService) ConstructNameForSuccessRateMetric(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error) {
	args := mps.Called(sourceNetworkId, targetNetworkId, asset, transactionId, metricTarget)
	return args.Get(0).(string), args.Error(1)
}

// IsNative Returns true if the asset is native to the passed network
func (mps *MockPrometheusService) IsNative(networkId int64, asset string) bool {
	args := mps.Called(networkId, asset)
	return args.Get(0).(bool)
}

// NativeToWrapped Getting Wrapped token from Native token for the given network/chain ids
func (mps *MockPrometheusService) NativeToWrapped(nativeAsset string, nativeChainId, targetChainId int64) string {
	args := mps.Called(nativeAsset, nativeChainId, targetChainId)
	return args.Get(0).(string)
}

// WrappedToNative Getting Native token from Wrapped token for the given network/chain ids
func (mps *MockPrometheusService) WrappedToNative(wrappedAsset string, wrappedChainId int64) *config.NativeAsset {
	args := mps.Called(wrappedAsset, wrappedChainId)
	return args.Get(0).(*config.NativeAsset)
}
