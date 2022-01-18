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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

type Watcher struct {
	dashboardPolling          time.Duration
	client                    client.MirrorNode
	bridgeConfig              config.Bridge
	prometheusService         service.Prometheus
	enableMonitoring          bool
	payerAccountBalanceGauge  prometheus.Gauge
	bridgeAccountBalanceGauge prometheus.Gauge
}

func NewWatcher(
	dashboardPolling time.Duration,
	client client.MirrorNode,
	bridgeConfig config.Bridge,
	enableMonitoring bool,
	prometheusService service.Prometheus) *Watcher {

	var (
		payerAccountBalanceGauge  prometheus.Gauge
		bridgeAccountBalanceGauge prometheus.Gauge
	)

	if enableMonitoring && prometheusService != nil {
		payerAccountBalanceGauge = prometheusService.NewGaugeMetric(constants.FeeAccountAmountGaugeName, constants.FeeAccountAmountGaugeHelp)
		bridgeAccountBalanceGauge = prometheusService.NewGaugeMetric(constants.BridgeAccountAmountGaugeName, constants.BridgeAccountAmountGaugeHelp)
	}

	return &Watcher{
		dashboardPolling:          dashboardPolling,
		client:                    client,
		bridgeConfig:              bridgeConfig,
		prometheusService:         prometheusService,
		enableMonitoring:          enableMonitoring,
		payerAccountBalanceGauge:  payerAccountBalanceGauge,
		bridgeAccountBalanceGauge: bridgeAccountBalanceGauge,
	}
}

func (pw Watcher) Watch(q qi.Queue) {
	// there will be no handler, so the q is to implement the interface
	go pw.beginWatching()
}

func (pw Watcher) beginWatching() {
	//The queue will be not used
	dashboardPolling := pw.dashboardPolling
	node := pw.client
	bridgeConfig := pw.bridgeConfig
	pw.setMetrics(
		node,
		bridgeConfig,
		dashboardPolling)
}

func (pw Watcher) setMetrics(node client.MirrorNode, bridgeConfig config.Bridge, dashboardPolling time.Duration) {
	if !pw.enableMonitoring {
		return
	}

	for {

		pw.payerAccountBalanceGauge.Set(pw.getAccountBalance(node, bridgeConfig.Hedera.PayerAccount))
		pw.bridgeAccountBalanceGauge.Set(pw.getAccountBalance(node, bridgeConfig.Hedera.BridgeAccount))

		log.Infoln("Dashboard Polling interval: ", dashboardPolling)
		time.Sleep(dashboardPolling)
	}
}

func (pw Watcher) getAccountBalance(node client.MirrorNode, accountId string) float64 {
	account, e := node.GetAccount(accountId)
	if e != nil {
		panic(e)
	}
	accountBalance := float64(account.Balance.Balance)
	tinyBarBalance := float64(hedera.NewHbar(accountBalance).AsTinybar())
	log.Infof("The Account with ID [%s] has balance AsTinybar = %b", accountId, tinyBarBalance)
	return tinyBarBalance
}
