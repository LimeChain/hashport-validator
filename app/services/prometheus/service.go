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
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Service struct {
	mu           sync.RWMutex
	logger       *log.Entry
	gauges       map[string]prometheus.Gauge
	counters     map[string]prometheus.Counter
	assetsConfig config.Assets
}

func NewService(assetsConfig config.Assets) *Service {

	return &Service{
		logger:       config.GetLoggerFor("Prometheus Service"),
		gauges:       map[string]prometheus.Gauge{},
		counters:     map[string]prometheus.Counter{},
		assetsConfig: assetsConfig,
	}
}

//func (s Service) A

func (s *Service) CreateAndRegisterGaugeMetric(name string, help string) prometheus.Gauge {
	s.mu.Lock()
	defer s.mu.Unlock()

	if gauge, exist := s.gauges[name]; exist {
		return gauge
	}

	s.logger.Infof("Creating Gauge Metric '%v' ...", name)
	opts := prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}
	gauge := prometheus.NewGauge(opts)
	s.logger.Infof("Gauge Metric '%v' successfully created!", name)

	s.logger.Infof("Registering Gauge Metric '%v' ...", name)
	prometheus.MustRegister(gauge)
	s.logger.Infof("Gauge Metric '%v' successfully registed!", name)

	s.gauges[name] = gauge

	return gauge
}

func (s *Service) GetGauge(name string) prometheus.Gauge {
	s.mu.Lock()
	defer s.mu.Unlock()

	gauge := s.gauges[name]
	return gauge
}

func (s *Service) CreateAndRegisterCounterMetric(name string, help string) prometheus.Counter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if counter, exist := s.counters[name]; exist {
		return counter
	}

	s.logger.Infof("Creating Counter Metric '%v' ...", name)
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}
	counter := prometheus.NewCounter(opts)
	s.logger.Infof("Counter Metric '%v' successfully created!", name)

	s.logger.Infof("Registering Counter Metric '%v' ...", name)
	prometheus.MustRegister(counter)
	s.logger.Infof("Counter Metric '%v' successfully registed!", name)

	s.counters[name] = counter

	return counter
}

func (s *Service) ConstructNameForSuccessRateMetric(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error) {
	tokenType := constants.Wrapped
	isNativeAsset := s.assetsConfig.IsNative(int64(sourceNetworkId), asset)
	if isNativeAsset {
		tokenType = constants.Native
	}

	return metrics.ConstructNameForMetric(sourceNetworkId, targetNetworkId, tokenType, transactionId, metricTarget)
}

func (s *Service) GetCounter(name string) prometheus.Counter {
	s.mu.Lock()
	defer s.mu.Unlock()

	counter := s.counters[name]
	return counter
}

func (s *Service) IsNative(networkId int64, asset string) bool {
	return s.assetsConfig.IsNative(networkId, asset)
}

func (s *Service) NativeToWrapped(nativeAsset string, nativeNetworkId, targetNetworkId int64) string {
	return s.assetsConfig.NativeToWrapped(nativeAsset, nativeNetworkId, targetNetworkId)
}

func (s *Service) WrappedToNative(wrappedAsset string, wrappedNetworkId int64) *config.NativeAsset {
	nativeAsset := s.assetsConfig.WrappedToNative(wrappedAsset, wrappedNetworkId)
	return nativeAsset
}
