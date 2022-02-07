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
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
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

	if prometheusService.GetIsMonitoringEnabled() {
		payerAccountBalanceGauge = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{Name: constants.FeeAccountAmountGaugeName, Help: constants.FeeAccountAmountGaugeHelp})
		bridgeAccountBalanceGauge = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{Name: constants.BridgeAccountAmountGaugeName, Help: constants.BridgeAccountAmountGaugeHelp})
		operatorBalanceGauge = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{Name: constants.OperatorAccountAmountName, Help: constants.OperatorAccountAmountHelp})
	}

	return &Watcher{
		dashboardPolling:          dashboardPolling,
		mirrorNode:                mirrorNode,
		EVMClients:                EVMClients,
		configuration:             configuration,
		prometheusService:         prometheusService,
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
			isNativeAsset := pw.configuration.Bridge.Assets.IsNative(chainId, asset)
			nativeNetworkName, err := pw.getNativeNetworkName(chainId, asset, isNativeAsset)
			if err != nil {
				panic(err)
			}
			if chainId == constants.HederaNetworkId { // Hedera
				if asset != constants.Hbar {
					res, e := pw.mirrorNode.GetToken(asset)
					if e != nil {
						panic(e)
					}
					pw.registerHederaAssetsSupplyMetrics(res, isNativeAsset, nativeNetworkName)
					pw.registerBridgeAccAssetsMetrics(res, isNativeAsset, nativeNetworkName)
				} else { // HBAR
					pw.registerHbarSuppyMetric(asset, isNativeAsset, nativeNetworkName)
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

				pw.registerEvmAssetSupplyMetric(asset, name, isNativeAsset, nativeNetworkName)
				pw.registerEvmAssetBalanceMetric(asset, name, address, isNativeAsset, nativeNetworkName)
			}
		}
	}
}

func (pw Watcher) getNativeNetworkName(chainId int64, asset string, isNativeAsset bool) (string, error) {
	errMsg := "Network id %v is missing in id to name mapping."
	if isNativeAsset {
		nativeNetworkName, exist := constants.NetworksById[uint64(chainId)]
		if !exist {
			return "", errors.New(fmt.Sprintf(errMsg, chainId))
		} else {
			return nativeNetworkName, nil
		}
	} else {
		nativeNetworkName, exist := constants.NetworksById[uint64(pw.configuration.Bridge.Assets.GetWrappedToNativeNetwork(asset))]
		if !exist {
			return "", errors.New(fmt.Sprintf(errMsg, chainId))
		} else {
			return nativeNetworkName, nil
		}
	}
}

func (pw Watcher) registerHederaAssetsSupplyMetrics(
	res *model.TokenResponse,
	isNativeAsset bool,
	nativeNetworkName string,
) {
	tokenType := constants.Wrapped
	if isNativeAsset {
		tokenType = constants.Native
	}
	name := fmt.Sprintf("%s_%s%s%s",
		tokenType,
		nativeNetworkName,
		constants.SupplyAssetMetricNameSuffix,
		tokenIDtoMetricName(res.TokenID))
	help := fmt.Sprintf("%s%s %s",
		constants.SupplyAssetMetricsHelpPrefix,
		tokenType,
		res.Name)
	pw.initAndRegAssetMetric(res.TokenID, pw.supplyAssetsMetrics, name, help, res.Name)
}

func (pw Watcher) registerHbarSuppyMetric(asset string, isNativeAsset bool, nativeNetworkName string) {
	tokenType := constants.Wrapped
	if isNativeAsset {
		tokenType = constants.Native
	}
	name := fmt.Sprintf("%s_%s%s%s",
		tokenType,
		nativeNetworkName,
		constants.SupplyAssetMetricNameSuffix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s %s",
		constants.SupplyAssetMetricsHelpPrefix,
		tokenType,
		asset)
	pw.initAndRegAssetMetric(asset, pw.supplyAssetsMetrics, name, help, constants.Hbar)
}

func (pw Watcher) registerBridgeAccAssetsMetrics(res *model.TokenResponse, isNativeAsset bool, nativeNetworkName string) {
	tokenType := constants.Wrapped
	if isNativeAsset {
		tokenType = constants.Native
	}
	name := fmt.Sprintf("%s_%s%s%s",
		tokenType,
		nativeNetworkName,
		constants.BridgeAccAssetMetricsNameSuffix,
		tokenIDtoMetricName(res.TokenID))
	help := fmt.Sprintf("%s%s %s",
		constants.BridgeAccAssetMetricsNameHelp,
		tokenType,
		res.TokenID)
	pw.initAndRegAssetMetric(res.TokenID, pw.bridgeAccAssetsMetrics, name, help, res.Name)
}

func (pw Watcher) registerEvmAssetSupplyMetric(
	asset string,
	name string,
	isNativeAsset bool,
	nativeNetworkName string,
) {
	tokenType := constants.Wrapped
	if isNativeAsset {
		tokenType = constants.Native
	}
	assetMetricName := fmt.Sprintf("%s_%s%s%s",
		tokenType,
		nativeNetworkName,
		constants.SupplyAssetMetricNameSuffix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s %s",
		constants.SupplyAssetMetricsHelpPrefix,
		tokenType,
		name)
	pw.initAndRegAssetMetric(asset, pw.supplyAssetsMetrics, assetMetricName, help, name)
}

func (pw Watcher) registerEvmAssetBalanceMetric(
	asset string,
	name string,
	address string,
	isNativeAsset bool,
	nativeNetworkName string,
) {
	tokenType := constants.Wrapped
	if isNativeAsset {
		tokenType = constants.Native
	}
	metricName := fmt.Sprintf("%s_%s%s%s",
		tokenType,
		nativeNetworkName,
		constants.BalanceAssetMetricNameSuffix,
		tokenIDtoMetricName(asset))
	help := fmt.Sprintf("%s%s %s%s%s",
		constants.BalanceAssetMetricHelpPrefix,
		tokenType,
		name,
		constants.AssetMetricHelpSuffix,
		address)
	pw.initAndRegAssetMetric(asset, pw.balanceAssetsMetrics, metricName, help, name)
}

func tokenIDtoMetricName(id string) string {
	replace := metrics.PrepareIdForPrometheus(id)
	result := fmt.Sprintf("%s%s", constants.AssetMetricsNamePrefix, replace)
	return result
}

func (pw Watcher) initAndRegAssetMetric(
	asset string,
	metricsMap map[string]string,
	name string,
	help string,
	assetName string) {
	metricsMap[asset] = name
	pw.prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
		Name: name,
		Help: help,
		ConstLabels: prometheus.Labels{
			"name": assetName,
		},
	})
}

func (pw Watcher) setMetrics() {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
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
	balance := pw.convertToHbar(account.Balance.Balance)
	log.Infof("The Account with ID [%s] has balance = %f", account.Account, balance)
	return balance
}

func (pw Watcher) convertToHbar(amount int) float64 {
	hbar := hedera.HbarFromTinybar(int64(amount))
	return hbar.As(hedera.HbarUnits.Hbar)
}

func (pw Watcher) setAssetsMetrics() {
	fungibleNetworkAssets := pw.configuration.Bridge.Assets.GetFungibleNetworkAssets()
	for chainId, assetArr := range fungibleNetworkAssets {
		for _, asset := range assetArr {
			if chainId == constants.HederaNetworkId { // Hedera
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
				decimal, e := wrappedInstance.Decimals(&bind.CallOpts{})
				if e != nil {
					panic(e)
				}

				pw.setEvmAssetSupplyMetric(wrappedInstance, asset, decimal)
				pw.setEvmAssetBalanceMetric(wrappedInstance, asset, decimal, address)
			}
		}
	}
}

func (pw Watcher) setBridgeAccAssetsMetrics(bridgeAcc *model.AccountsResponse) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	for _, token := range bridgeAcc.Balance.Tokens {
		metric := pw.prometheusService.GetGauge(pw.bridgeAccAssetsMetrics[token.TokenID])
		if metric == nil {
			log.Infof("Skip metrics for Bridge Account asset with ID [%s]", token.TokenID)
		} else {
			balance := pw.convertToHbar(token.Balance)
			metric.Set(balance)
			log.Infof("The Bridge Account asset with ID [%s] has Balance = %f", token.TokenID, balance)
		}
	}
}

func (pw Watcher) setHederaAssetSupplyMetric(asset string, token *model.TokenResponse) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	totalSupplyMetric := pw.prometheusService.GetGauge(pw.supplyAssetsMetrics[asset])
	pw.setHederaAssetMetric(totalSupplyMetric, token.TotalSupply, token.Decimals, asset)
}

func (pw Watcher) setHederaNetworkSupply(asset string) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	supply, e := pw.mirrorNode.GetNetworkSupply()
	if e != nil {
		panic(e)
	}
	totalSupplyMetric := pw.prometheusService.GetGauge(pw.supplyAssetsMetrics[asset])
	pw.setHederaAssetMetric(totalSupplyMetric, supply.TotalSupply, strconv.Itoa(constants.HederaDecimal), asset)
}

func (pw Watcher) setHederaAssetMetric(metric prometheus.Gauge, value string, decimal string, asset string) {
	d, _ := strconv.Atoi(decimal)
	dec := uint8(d)

	intVal, ok := new(big.Int).SetString(value, 10)
	if !ok {
		log.Infof(`"SetString: error": [%s].`, value)
	}

	parseValue, e := pw.convertBasedOnDecimal(intVal, dec)
	if e != nil {
		panic(e)
	}
	metric.Set(*parseValue)
	log.Infof("The Asset with ID [%s] has Total Supply = %f", asset, parseValue)
}

func (pw Watcher) setEvmAssetSupplyMetric(wrappedInstance *wtoken.Wtoken, asset string, decimal uint8) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	ts, e := wrappedInstance.TotalSupply(&bind.CallOpts{})
	if e != nil {
		panic(e)
	}
	totalSupply, e := pw.convertBasedOnDecimal(ts, decimal)
	if e != nil {
		panic(e)
	}

	totalSupplyMetric := pw.prometheusService.GetGauge(pw.supplyAssetsMetrics[asset])
	pw.setEvmAssetMetric(totalSupplyMetric, asset, "Total Supply", *totalSupply)
}

func (pw Watcher) setEvmAssetBalanceMetric(
	wrappedInstance *wtoken.Wtoken,
	asset string,
	decimal uint8,
	address common.Address,
) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	b, e := wrappedInstance.BalanceOf(&bind.CallOpts{}, address)
	if e != nil {
		panic(e)
	}
	balance, e := pw.convertBasedOnDecimal(b, decimal)
	if e != nil {
		panic(e)
	}

	balanceMetric := pw.prometheusService.GetGauge(pw.balanceAssetsMetrics[asset])
	pw.setEvmAssetMetric(balanceMetric, asset, "Balance of", *balance)
}

func (pw Watcher) setEvmAssetMetric(metric prometheus.Gauge, asset string, metricName string, value float64) {
	metric.Set(value)
	log.Infof("The Asset with ID [%s] has %s = %f", asset, metricName, value)
}

func (pw Watcher) convertBasedOnDecimal(value *big.Int, decimal uint8) (*float64, error) {
	if decimal < 1 {
		return nil, errors.New(fmt.Sprintf(`Failed to calc with decimal: [%d].`, decimal))
	}
	parseValue, e := strconv.ParseFloat(constants.CreateDecimalPrefix+strings.Repeat(constants.CreateDecimalRepeat, int(decimal)), 64)
	if e != nil {
		return nil, e
	}
	res, _ := new(big.Float).Set(new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(parseValue))).Float64()
	return &res, nil
}
