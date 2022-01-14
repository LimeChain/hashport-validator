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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
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
	mirrorNode                client.MirrorNode
	configuration             config.Config
	prometheusService         service.Prometheus
	enableMonitoring          bool
	payerAccountBalanceGauge  prometheus.Gauge
	bridgeAccountBalanceGauge prometheus.Gauge
	operatorBalanceGauge      prometheus.Gauge
}

func NewWatcher(
	dashboardPolling time.Duration,
	mirrorNode client.MirrorNode,
	configuration config.Config,
	enableMonitoring bool,
	prometheusService service.Prometheus) *Watcher {

	var (
		payerAccountBalanceGauge  prometheus.Gauge
		bridgeAccountBalanceGauge prometheus.Gauge
		operatorBalanceGauge      prometheus.Gauge
	)

	if enableMonitoring && prometheusService != nil {
		payerAccountBalanceGauge = prometheusService.GetGauge(constants.FeeAccountAmountGaugeName)
		bridgeAccountBalanceGauge = prometheusService.GetGauge(constants.BridgeAccountAmountGaugeName)
		operatorBalanceGauge = prometheusService.GetGauge(constants.OperatorAccountAmountName)
	}

	return &Watcher{
		dashboardPolling:          dashboardPolling,
		mirrorNode:                mirrorNode,
		configuration:             configuration,
		prometheusService:         prometheusService,
		enableMonitoring:          enableMonitoring,
		payerAccountBalanceGauge:  payerAccountBalanceGauge,
		bridgeAccountBalanceGauge: bridgeAccountBalanceGauge,
		operatorBalanceGauge:      operatorBalanceGauge,
	}
}

func (pw Watcher) Watch(q qi.Queue) {
	// there will be no handler, so the q is to implement the interface
	go pw.beginWatching()
}

func (pw Watcher) beginWatching() {
	//The queue will be not used
	pw.setMetrics()
}

func (pw Watcher) setMetrics() {
	if !pw.enableMonitoring {
		return
	}

	for {
		payerAccount := pw.getAccount(pw.configuration.Bridge.Hedera.PayerAccount)
		bridgeAccount := pw.getAccount(pw.configuration.Bridge.Hedera.BridgeAccount)
		operatorAccount := pw.getAccount(pw.configuration.Node.Clients.Hedera.Operator.AccountId)

		pw.payerAccountBalanceGauge.Set(pw.getAccountBalance(payerAccount))
		pw.bridgeAccountBalanceGauge.Set(pw.getAccountBalance(bridgeAccount))
		pw.operatorBalanceGauge.Set(pw.getAccountBalance(operatorAccount))

		log.Infoln("Dashboard Polling interval: ", pw.dashboardPolling)
		time.Sleep(pw.dashboardPolling)
	}
}

func (pw Watcher) getAccount(accountId string) *model.AccountsResponse {
	account, e := pw.mirrorNode.GetAccount(accountId)
	if e != nil {
		panic(e)
	}
	return account
}
func (pw Watcher) getAccountBalance(account *model.AccountsResponse) float64 {
	balance := float64(account.Balance.Balance)
	log.Infof("The Account with ID [%s] has balance = %f", account.Account, balance)

	return balance
}
