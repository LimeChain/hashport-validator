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

package prometheus

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

type Service struct {
	mu                  sync.RWMutex
	logger              *log.Entry
	gauges              map[string]prometheus.Gauge
	counters            map[string]prometheus.Counter
	isMonitoringEnabled bool
	assetsConfig        config.Assets
}

func NewService(assetsConfig config.Assets, isMonitoringEnabled bool) *Service {

	return &Service{
		logger:              config.GetLoggerFor("Prometheus Service"),
		gauges:              map[string]prometheus.Gauge{},
		counters:            map[string]prometheus.Counter{},
		isMonitoringEnabled: isMonitoringEnabled,
		assetsConfig:        assetsConfig,
	}
}

func (s *Service) CreateAndRegisterGaugeMetric(name string, help string, labels prometheus.Labels) prometheus.Gauge {
	if !s.isMonitoringEnabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if gauge, exist := s.gauges[name]; exist {
		return gauge
	}

	s.logger.Infof("Creating Gauge Metric '%v' ...", name)
	opts := prometheus.GaugeOpts{
		Name:        name,
		Help:        help,
		ConstLabels: labels,
	}
	gauge := prometheus.NewGauge(opts)
	s.logger.Infof("Gauge Metric '%v' successfully created! Labels: %s", name, labels)

	s.logger.Infof("Registering Gauge Metric '%v' ...", name)
	prometheus.MustRegister(gauge)
	s.logger.Infof("Gauge Metric '%v' successfully registed!", name)

	s.gauges[name] = gauge

	return gauge
}

func (s *Service) CreateAndRegisterGaugeMetricForSuccessRate(transactionId string, sourceChainId int64, targetChainId int64, asset, metricType, metricHelp string) (prometheus.Gauge, error) {
	if !s.isMonitoringEnabled {
		return nil, errors.New("monitoring is disabled.")
	}

	metricName, err := s.ConstructMetricName(
		uint64(sourceChainId),
		uint64(targetChainId),
		asset,
		transactionId,
		metricType,
	)

	if err != nil {
		return nil, err
	}

	gauge := s.CreateAndRegisterGaugeMetric(metricName, metricHelp, prometheus.Labels{
		"source_network_id": strconv.FormatInt(sourceChainId, 10),
		"target_network_id": strconv.FormatInt(targetChainId, 10),
		"asset":             asset,
		"transaction_id":    transactionId,
		"metric_type":       metricType,
	})

	return gauge, nil
}

func (s *Service) GetGauge(name string) prometheus.Gauge {
	if !s.isMonitoringEnabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	gauge := s.gauges[name]
	return gauge
}

func (s *Service) UnregisterGauge(name string) {
	if !s.isMonitoringEnabled {
		return
	}

	s.logger.Infof("Unregistering Gauge Metric '%v' ...", name)
	gauge := s.GetGauge(name)
	prometheus.Unregister(gauge)
	delete(s.gauges, name)
	s.logger.Infof("Gauge Metric '%v' successfully unregisted!", name)
}

func (s *Service) CreateAndRegisterCounterMetric(name string, help string, labels prometheus.Labels) prometheus.Counter {
	if !s.isMonitoringEnabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if counter, exist := s.counters[name]; exist {
		return counter
	}

	s.logger.Infof("Creating Counter Metric '%v' ...", name)
	opts := prometheus.CounterOpts{
		Name:        name,
		Help:        help,
		ConstLabels: labels,
	}
	counter := prometheus.NewCounter(opts)
	s.logger.Infof("Counter Metric '%v' successfully created!", name)

	s.logger.Infof("Registering Counter Metric '%v' ...", name)
	prometheus.MustRegister(counter)
	s.logger.Infof("Counter Metric '%v' successfully registed!", name)

	s.counters[name] = counter

	return counter
}

func (s *Service) ConstructMetricName(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricType string) (string, error) {
	tokenType := constants.Wrapped
	isNativeAsset := s.assetsConfig.IsNative(int64(sourceNetworkId), asset)
	if isNativeAsset {
		tokenType = constants.Native
	}

	return metrics.ConstructNameForMetric(sourceNetworkId, targetNetworkId, tokenType, transactionId, metricType)
}

func (s *Service) GetCounter(name string) prometheus.Counter {
	if !s.isMonitoringEnabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	counter := s.counters[name]
	return counter
}

func (s *Service) UnregisterCounter(name string) {
	if !s.isMonitoringEnabled {
		return
	}

	s.logger.Infof("Unregistering Counter Metric '%v' ...", name)
	counter := s.GetCounter(name)
	prometheus.Unregister(counter)
	delete(s.counters, name)
	s.logger.Infof("Counter Metric '%v' successfully unregisted!", name)
}

func (s *Service) GetIsMonitoringEnabled() bool {
	return s.isMonitoringEnabled
}
