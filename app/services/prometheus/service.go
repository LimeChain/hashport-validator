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
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger *log.Entry
	gauges map[string]prometheus.Gauge
}

func NewService() *Service {

	return &Service{
		logger: config.GetLoggerFor("Prometheus Service"),
		gauges: map[string]prometheus.Gauge{},
	}
}

func (s Service) CreateAndRegisterGaugeMetric(name string, help string) prometheus.Gauge {

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

	s.logger.Infof("Registring Gauge Metric '%v' ...", name)
	prometheus.MustRegister(gauge)
	s.logger.Infof("Gauge Metric '%v' successfully registed!", name)

	s.gauges[name] = gauge

	return gauge
}

func (s Service) GetGauge(name string) prometheus.Gauge {
	gauge := s.gauges[name]
	return gauge
}
