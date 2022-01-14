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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	mirrorNode         client.MirrorNode
	transferRepository repository.Transfer
	pollingInterval    time.Duration
	logger             *log.Entry
	gauges             map[string]prometheus.Gauge
}

func NewService(
	mirrorNode client.MirrorNode,
	transferRepository repository.Transfer,
	pollingInterval time.Duration) *Service {

	return &Service{
		mirrorNode:         mirrorNode,
		transferRepository: transferRepository,
		pollingInterval:    pollingInterval,
		logger:             config.GetLoggerFor("Prometheus Service"),
		gauges:             map[string]prometheus.Gauge{},
	}
}

func (s Service) NewGaugeMetric(name string, help string) prometheus.Gauge {
	opts := prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}
	gauge := prometheus.NewGauge(opts)
	s.gauges[name] = gauge

	return gauge
}

func (s Service) RegisterGaugeMetric(gauge prometheus.Gauge) {
	prometheus.MustRegister(gauge)
}

func (s Service) GetGauge(name string) prometheus.Gauge {
	gauge := s.gauges[name]
	return gauge
}
