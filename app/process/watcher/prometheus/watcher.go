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
	prometheusServices "github.com/limechain/hedera-eth-bridge-validator/app/services/prometheus"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	payerAccountBalance = prometheusServices.NewGaugeMetric(constants.FeeAccountAmountName,
		constants.FeeAccountAmountHelp)
	bridgeAccountBalance = prometheusServices.NewGaugeMetric(constants.BridgeAccountAmountName,
		constants.BridgeAccountAmountHelp)
)

func registerMetrics() {
	//Fee Account Balance
	prometheusServices.RegisterGaugeMetric(payerAccountBalance)
	//Bridge Account Balance
	prometheusServices.RegisterGaugeMetric(bridgeAccountBalance)
}

type Watcher struct {
	dashboardPolling time.Duration
	client           client.MirrorNode
	bridgeConfig     config.Bridge
}

func NewWatcher(dashboardPolling time.Duration, client client.MirrorNode, bridgeConfig config.Bridge) *Watcher {
	return &Watcher{
		dashboardPolling: dashboardPolling,
		client:           client,
		bridgeConfig:     bridgeConfig,
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
	registerMetrics()
	setMetrics(
		node,
		bridgeConfig,
		dashboardPolling)
}

func setMetrics(node client.MirrorNode, bridgeConfig config.Bridge, dashboardPolling time.Duration) {
	for {
		payerAccountBalance.Set(getAccountBalance(node, bridgeConfig.Hedera.PayerAccount))
		bridgeAccountBalance.Set(getAccountBalance(node, bridgeConfig.Hedera.BridgeAccount))

		log.Infoln("Dashboard Polling interval: ", dashboardPolling)
		time.Sleep(dashboardPolling)
	}
}

func getAccountBalance(node client.MirrorNode, accountId string) float64 {
	account, e := node.GetAccount(accountId)
	if e != nil {
		panic(e)
	}
	accountBalance := float64(account.Balance.Balance)
	tinyBarBalance := float64(hedera.NewHbar(accountBalance).AsTinybar())
	log.Infof("The Account with ID [%s] has balance AsTinybar = %b", accountId, tinyBarBalance)
	return tinyBarBalance
}
