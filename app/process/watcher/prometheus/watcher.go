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
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type Watcher struct {
	dashboardPolling          time.Duration
	mirrorNode                client.MirrorNode
	EVMClients                map[int64]client.EVM
	configuration             config.Config
	prometheusService         service.Prometheus
	enableMonitoring          bool
	payerAccountBalanceGauge  prometheus.Gauge
	bridgeAccountBalanceGauge prometheus.Gauge
	operatorBalanceGauge      prometheus.Gauge
	// A mapping, storing all assets <> metric name
	supplyAssetsMetrics    map[string]string
	balanceAssetsMetrics   map[string]string
	bridgeAccAssetsMetrics map[string]string
}

func NewWatcher(
	dashboardPolling time.Duration,
	mirrorNode client.MirrorNode,
	configuration config.Config,
	enableMonitoring bool,
	prometheusService service.Prometheus,
	EVMClients map[int64]client.EVM,
) *Watcher {

	var (
		payerAccountBalanceGauge  prometheus.Gauge
		bridgeAccountBalanceGauge prometheus.Gauge
		operatorBalanceGauge      prometheus.Gauge
		// A mapping, storing all assets <> metric name
		supplyAssetsMetrics    = make(map[string]string)
		balanceAssetsMetrics   = make(map[string]string)
		bridgeAccAssetsMetrics = make(map[string]string)
	)

	if enableMonitoring && prometheusService != nil {
		payerAccountBalanceGauge = prometheusService.CreateAndRegisterGaugeMetric(constants.FeeAccountAmountGaugeName, constants.FeeAccountAmountGaugeHelp)
		bridgeAccountBalanceGauge = prometheusService.CreateAndRegisterGaugeMetric(constants.BridgeAccountAmountGaugeName, constants.BridgeAccountAmountGaugeHelp)
		operatorBalanceGauge = prometheusService.CreateAndRegisterGaugeMetric(constants.OperatorAccountAmountName, constants.OperatorAccountAmountHelp)
	}

	return &Watcher{
		dashboardPolling:          dashboardPolling,
		mirrorNode:                mirrorNode,
		EVMClients:                EVMClients,
		configuration:             configuration,
		prometheusService:         prometheusService,
		enableMonitoring:          enableMonitoring,
		payerAccountBalanceGauge:  payerAccountBalanceGauge,
		bridgeAccountBalanceGauge: bridgeAccountBalanceGauge,
		operatorBalanceGauge:      operatorBalanceGauge,
		supplyAssetsMetrics:       supplyAssetsMetrics,
		balanceAssetsMetrics:      balanceAssetsMetrics,
		bridgeAccAssetsMetrics:    bridgeAccAssetsMetrics,
	}
}

func (pw Watcher) Watch(q qi.Queue) {
	// there will be no handler, so the q is to implement the interface
	go pw.beginWatching()
}

func (pw Watcher) beginWatching() {
	//The queue will be not used
	pw.registerAssetsMetrics()
	pw.setMetrics()
}

func (pw Watcher) registerAssetsMetrics() {
	fungibleNetworkAssets := pw.configuration.Bridge.Assets.GetFungibleNetworkAssets()
	for chainId, assetArr := range fungibleNetworkAssets {
		for _, asset := range assetArr {
			if chainId == 0 { // Hedera
				if asset != constants.Hbar {
					res, e := pw.mirrorNode.GetToken(asset)
					if e != nil {
						panic(e)
					}
					pw.registerHederaAssetsSupplyMetrics(asset, res)
					pw.registerBridgeAccAssetsMetrics(asset)
				} else { // HBAR
					pw.registerHbarSuppyMetric(asset)
				}
			} else { // EVM
				evm := pw.EVMClients[chainId].GetClient()
				wrappedInstance, e := wtoken.NewWtoken(common.HexToAddress(asset), evm)
				if e != nil {
					panic(e)
				}
				name, e := wrappedInstance.Name(&bind.CallOpts{})
				if e != nil {
					panic(e)
				}
				address := pw.configuration.Bridge.EVMs[chainId].RouterContractAddress

				pw.registerEvmAssetSupplyMetric(asset, name)
				pw.registerEvmAssetBalanceMetric(asset, name, address)
			}
		}
	}
}

func (pw Watcher) registerHederaAssetsSupplyMetrics(asset string, res *model.TokenResponse) {
	name := fmt.Sprintf("%s%s",
		constants.SupplyAssetMetricNamePrefix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s",
		constants.SupplyAssetMetricsHelpPrefix,
		res.Name)
	pw.initAndRegAssetMetric(asset, pw.supplyAssetsMetrics, name, help)
}

func (pw Watcher) registerHbarSuppyMetric(asset string) {
	name := fmt.Sprintf("%s%s",
		constants.SupplyAssetMetricNamePrefix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s",
		constants.SupplyAssetMetricsHelpPrefix,
		asset)
	pw.initAndRegAssetMetric(asset, pw.supplyAssetsMetrics, name, help)
}

func (pw Watcher) registerBridgeAccAssetsMetrics(asset string) {
	name := fmt.Sprintf("%s%s",
		constants.BridgeAccAssetMetricsNamePrefix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s",
		constants.BridgeAccAssetMetricsNameHelp,
		asset)
	pw.initAndRegAssetMetric(asset, pw.bridgeAccAssetsMetrics, name, help)
}

func (pw Watcher) registerEvmAssetSupplyMetric(asset string, name string) {
	assetMetricName := fmt.Sprintf("%s%s",
		constants.SupplyAssetMetricNamePrefix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s", constants.SupplyAssetMetricsHelpPrefix, name)
	pw.initAndRegAssetMetric(asset, pw.supplyAssetsMetrics, assetMetricName, help)
}

func (pw Watcher) registerEvmAssetBalanceMetric(asset string, name string, address string) {
	metricName := fmt.Sprintf("%s%s",
		constants.BalanceAssetMetricNamePrefix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s%s%s",
		constants.BalanceAssetMetricHelpPrefix,
		name,
		constants.AssetMetricHelpSuffix,
		address)
	pw.initAndRegAssetMetric(asset, pw.balanceAssetsMetrics, metricName, help)
}

func tokenIDtoMetricName(id string) string {
	replace := strings.Replace(id, constants.DotSymbol, constants.ReplaceDotSymbol, constants.DotSymbolRep)
	result := fmt.Sprintf("%s%s", constants.AssetMetricsNamePrefix, replace)
	return result
}

func (pw Watcher) initAndRegAssetMetric(asset string, metricsMap map[string]string, name string, help string) {
	metricsMap[asset] = name
	pw.initAndRegGauge(name, help)
}

func (pw Watcher) initAndRegGauge(name string, help string) {
	pw.prometheusService.CreateAndRegisterGaugeMetric(name, help)
	log.Infof("Registered metric with name [%s] help [%s]", name, help)
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

		pw.setAssetsMetrics()
		pw.setBridgeAccAssetsMetrics(bridgeAccount)

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

func (pw Watcher) setAssetsMetrics() {
	fungibleNetworkAssets := pw.configuration.Bridge.Assets.GetFungibleNetworkAssets()
	for chainId, assetArr := range fungibleNetworkAssets {
		for _, asset := range assetArr {
			if chainId == 0 { // Hedera
				if asset != constants.Hbar {
					token, e := pw.mirrorNode.GetToken(asset)
					if e != nil {
						panic(e)
					}
					pw.setHederaAssetSupplyMetric(asset, token)
				} else { // HBAR
					pw.setHederaNetworkSupply(asset)
				}
			} else { // EVM
				evm := pw.EVMClients[chainId].GetClient()
				address := common.HexToAddress(pw.configuration.Bridge.EVMs[chainId].RouterContractAddress)

				wrappedInstance, e := wtoken.NewWtoken(common.HexToAddress(asset), evm)
				if e != nil {
					panic(e)
				}

				pw.setEvmAssetSupplyMetric(wrappedInstance, asset)
				pw.setEvmAssetBalanceMetric(wrappedInstance, asset, address)
			}
		}
	}
}

func (pw Watcher) setBridgeAccAssetsMetrics(bridgeAcc *model.AccountsResponse) {
	for _, token := range bridgeAcc.Balance.Tokens {
		metric := pw.prometheusService.GetGauge(pw.bridgeAccAssetsMetrics[token.TokenID])
		if metric == nil {
			log.Infof("Skip metrics for Bridge Account asset with ID [%s]", token.TokenID)
		} else {
			metric.Set(float64(token.Balance))
			log.Infof("The Bridge Account asset with ID [%s] has Balance = %f", token.TokenID, float64(token.Balance))
		}
	}
}

func (pw Watcher) setHederaAssetSupplyMetric(asset string, token *model.TokenResponse) {
	totalSupplyMetric := pw.prometheusService.GetGauge(pw.supplyAssetsMetrics[asset])
	pw.setHederaAssetMetric(totalSupplyMetric, token.TotalSupply, asset)
}

func (pw Watcher) setHederaNetworkSupply(asset string) {
	supply, e := pw.mirrorNode.GetNetworkSupply()
	if e != nil {
		panic(e)
	}
	totalSupplyMetric := pw.prometheusService.GetGauge(pw.supplyAssetsMetrics[asset])
	pw.setHederaAssetMetric(totalSupplyMetric, supply.TotalSupply, asset)
}

func (pw Watcher) setHederaAssetMetric(metric prometheus.Gauge, value string, asset string) {
	parseValue, e := strconv.ParseFloat(value, 64)
	if e != nil {
		panic(e)
	}
	metric.Set(parseValue)
	log.Infof("The Asset with ID [%s] has Total Supply = %f", asset, parseValue)
}

func (pw Watcher) setEvmAssetSupplyMetric(wrappedInstance *wtoken.Wtoken, asset string) {
	totalSupply, e := wrappedInstance.TotalSupply(&bind.CallOpts{})
	if e != nil {
		panic(e)
	}
	totalSupplyMetric := pw.prometheusService.GetGauge(pw.supplyAssetsMetrics[asset])
	pw.setEvmAssetMetric(totalSupplyMetric, asset, "Total Supply", totalSupply)
}

func (pw Watcher) setEvmAssetBalanceMetric(wrappedInstance *wtoken.Wtoken, asset string, address common.Address) {
	balance, e := wrappedInstance.BalanceOf(&bind.CallOpts{}, address)
	if e != nil {
		panic(e)
	}
	balanceMetric := pw.prometheusService.GetGauge(pw.balanceAssetsMetrics[asset])
	pw.setEvmAssetMetric(balanceMetric, asset, "Balance of", balance)
}

func (pw Watcher) setEvmAssetMetric(metric prometheus.Gauge, asset string, metricName string, value *big.Int) {
	parseValue, _ := new(big.Float).SetInt(value).Float64()
	metric.Set(parseValue)
	log.Infof("The Asset with ID [%s] has %s = %f", asset, metricName, parseValue)
}
