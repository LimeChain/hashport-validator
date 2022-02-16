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

package service

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus interface {
	// CreateGaugeIfNotExists creates new Gauge Metric and registers it in Prometheus if not exists
	CreateGaugeIfNotExists(opts prometheus.GaugeOpts) prometheus.Gauge
	// CreateSuccessRateGaugeIfNotExists creates new Gauge Metric for Success Rate and registers it in Prometheus if not exists
	CreateSuccessRateGaugeIfNotExists(transactionId string, sourceChainId, targetChainId uint64, asset, metricNameSuffix, metricHelp string) (prometheus.Gauge, error)
	// GetGauge retrieves Gauge by name with flag for existence
	GetGauge(name string) prometheus.Gauge
	// DeleteGauge unregisters and deletes Gauge with the passed name
	DeleteGauge(name string)
	// CreateCounterIfNotExists creates new Counter Metric and registers it in Prometheus
	CreateCounterIfNotExists(opts prometheus.CounterOpts) prometheus.Counter
	// GetCounter retrieves Counter by name with flag for existence
	GetCounter(name string) prometheus.Counter
	// DeleteCounter unregisters and deletes Counter with the passed name
	DeleteCounter(name string)
	// ConstructMetricName constructing name for metric
	ConstructMetricName(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error)
	// GetIsMonitoringEnabled returns if the monitoring is enabled
	GetIsMonitoringEnabled() bool
}
