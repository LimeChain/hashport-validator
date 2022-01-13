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

// NewGaugeMetric creates new Gauge Metric
func (mps *MockPrometheusService) NewGaugeMetric(name string, help string) prometheus.Gauge {
	args := mps.Called(name, help)
	result := args.Get(0).(prometheus.Gauge)
	return result
}

// RegisterGaugeMetric registering new Gauge Metric
func (mps *MockPrometheusService) RegisterGaugeMetric(gauge prometheus.Gauge) {
	mps.Called(gauge)
}

// GetGauge retrieves Gauge by name with flag for existence
func (mps *MockPrometheusService) GetGauge(name string) prometheus.Gauge {
	args := mps.Called(name)
	result := args.Get(0).(prometheus.Gauge)
	return result
}
