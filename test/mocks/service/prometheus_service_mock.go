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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
)

type MockPrometheusService struct {
	mock.Mock
}

// CreateAndRegisterGaugeMetricIfNotExists creates new Gauge Metric and registers it in Prometheus if not exists
func (mps *MockPrometheusService) CreateAndRegisterGaugeMetricIfNotExists(name string, help string, labels prometheus.Labels) prometheus.Gauge {
	args := mps.Called(name, help, labels)
	result := args.Get(0).(prometheus.Gauge)
	return result
}

// CreateAndRegisterSuccessRateGaugeMetricIfNotExists creates new Gauge Metric for Success Rate and registers it in Prometheus if not exists
func (mps *MockPrometheusService) CreateAndRegisterSuccessRateGaugeMetricIfNotExists(transactionId string, sourceChainId int64, targetChainId int64, asset, metricNameSuffix, metricHelp string) (prometheus.Gauge, error) {
	args := mps.Called(transactionId, sourceChainId, targetChainId, asset, metricNameSuffix, metricHelp)
	return args.Get(0).(prometheus.Gauge), args.Error(1)
}

// GetGauge retrieves Gauge by name with flag for existence
func (mps *MockPrometheusService) GetGauge(name string) prometheus.Gauge {
	args := mps.Called(name)
	result := args.Get(0).(prometheus.Gauge)
	return result
}

// UnregisterAndDeleteGauge unregisters and deletes Gauge with the passed name
func (mps *MockPrometheusService) UnregisterAndDeleteGauge(name string) {
	_ = mps.Called(name)
}

// CreateAndRegisterCounterMetricIfNotExists creates new Counter Metric and registers it in Prometheus
func (mps *MockPrometheusService) CreateAndRegisterCounterMetricIfNotExists(name string, help string, labels prometheus.Labels) prometheus.Counter {
	args := mps.Called(name, help, labels)
	result := args.Get(0).(prometheus.Counter)
	return result
}

// GetCounter retrieves Counter by name with flag for existence
func (mps *MockPrometheusService) GetCounter(name string) prometheus.Counter {
	args := mps.Called(name)
	result := args.Get(0).(prometheus.Counter)
	return result
}

// ConstructMetricName constructing name for metric
func (mps *MockPrometheusService) ConstructMetricName(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error) {
	args := mps.Called(sourceNetworkId, targetNetworkId, asset, transactionId, metricTarget)
	return args.Get(0).(string), args.Error(1)
}

// UnregisterAndDeleteCounter unregisters and deletes Counter with the passed name
func (mps *MockPrometheusService) UnregisterAndDeleteCounter(name string) {
	_ = mps.Called(name)
}

// GetIsMonitoringEnabled returns if the monitoring is enabled
func (mps *MockPrometheusService) GetIsMonitoringEnabled() bool {
	args := mps.Called()
	return args.Get(0).(bool)
}
