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
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
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
	// A mapping, storing all network ID - asset address - metric name
	assetsMetrics map[int64]map[string]string
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
		// A mapping, storing all network ID - asset address - metric name
		assetsMetrics = make(map[int64]map[string]string)
	)

	if prometheusService.GetIsMonitoringEnabled() {
		payerAccountBalanceGauge = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
			Name: constants.FeeAccountAmountGaugeName,
			Help: constants.FeeAccountAmountGaugeHelp,
			ConstLabels: prometheus.Labels{
				constants.AccountMetricLabelKey: configuration.Bridge.Hedera.PayerAccount,
			},
		})
		bridgeAccountBalanceGauge = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
			Name: constants.BridgeAccountAmountGaugeName,
			Help: constants.BridgeAccountAmountGaugeHelp,
			ConstLabels: prometheus.Labels{
				constants.AccountMetricLabelKey: configuration.Bridge.Hedera.BridgeAccount,
			},
		})
		operatorBalanceGauge = prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
			Name: constants.OperatorAccountAmountName,
			Help: constants.OperatorAccountAmountHelp,
			ConstLabels: prometheus.Labels{
				constants.AccountMetricLabelKey: configuration.Node.Clients.Hedera.Operator.AccountId,
			},
		})
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
		assetsMetrics:             assetsMetrics,
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
	nativeToWrapped := pw.configuration.Bridge.Assets.GetNativeToWrapped()
	for nativeNetworkId, nativeMap := range nativeToWrapped {
		for nativeAsset, wrappedMap := range nativeMap {
			// register native assets balance
			pw.regAssetMetric(
				nativeNetworkId,
				nativeNetworkId,
				nativeAsset,
				constants.BalanceAssetMetricNameSuffix,
				constants.BalanceAssetMetricHelpPrefix,
			)
			for wrappedNetworkId, wrappedAsset := range wrappedMap {
				//register wrapped assets total supply
				pw.regAssetMetric(
					nativeNetworkId,
					wrappedNetworkId,
					wrappedAsset,
					constants.SupplyAssetMetricNameSuffix,
					constants.SupplyAssetMetricsHelpPrefix,
				)
			}
		}
	}
}

func (pw Watcher) regAssetMetric(
	nativeNetworkId int64,
	wrappedNetworkId int64,
	assetAddress string,
	metricNameCnt string,
	metricHelpCnt string,
) {
	if assetAddress != constants.Hbar { // skip HBAR
		assetName, assetSymbol := pw.getAssetRegData(wrappedNetworkId, assetAddress)
		metricName, metricHelp := getMetricData(
			nativeNetworkId,
			wrappedNetworkId,
			assetAddress,
			assetName,
			metricNameCnt,
			metricHelpCnt,
		)
		if pw.assetsMetrics[wrappedNetworkId] == nil {
			pw.assetsMetrics[wrappedNetworkId] = make(map[string]string)
		}
		pw.assetsMetrics[wrappedNetworkId][assetAddress] = metricName
		pw.prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
			Name: metricName,
			Help: metricHelp,
			ConstLabels: prometheus.Labels{
				constants.AssetMetricLabelKey: assetSymbol,
			},
		})
	}
}

func (pw Watcher) getAssetRegData(networkId int64, assetAddress string) (name string, symbol string) {
	if networkId == constants.HederaNetworkId { // Hedera
		asset, e := pw.mirrorNode.GetToken(assetAddress)
		if e != nil {
			panic(e)
		}
		name = asset.Name
		symbol = asset.Symbol
	} else { // EVM
		evm := pw.EVMClients[networkId].GetClient()
		evmAssetInstance, e := wtoken.NewWtoken(common.HexToAddress(assetAddress), evm)
		if e != nil {
			panic(e)
		}
		resName, e := evmAssetInstance.Name(&bind.CallOpts{})
		if e != nil {
			panic(e)
		}
		name = resName
		resSymbol, e := evmAssetInstance.Symbol(&bind.CallOpts{})
		if e != nil {
			panic(e)
		}
		symbol = resSymbol
	}
	return name, symbol
}

func getMetricData(
	nativeNetworkId int64,
	wrappedNetworkId int64,
	assetAddress string,
	assetName string,
	metricNameSuffix string,
	metricsHelpPrefix string,
) (string, string) {

	nativeNetworkName := constants.NetworksById[uint64(nativeNetworkId)]
	wrappedNetworkName := constants.NetworksById[uint64(wrappedNetworkId)]
	assetType := constants.Wrapped
	if nativeNetworkId == wrappedNetworkId {
		assetType = constants.Native
	}
	name := fmt.Sprintf("%s_%s_%s%s%s",
		assetType,
		nativeNetworkName,
		wrappedNetworkName,
		metricNameSuffix,
		metrics.AssetAddressToMetricName(assetAddress))
	help := fmt.Sprintf("%s %s %s",
		metricsHelpPrefix,
		assetType,
		assetName)
	return name, help
}

func (pw Watcher) setMetrics() {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	for {
		payerAccount := pw.getAccount(pw.configuration.Bridge.Hedera.PayerAccount)
		bridgeAccount := pw.getAccount(pw.configuration.Bridge.Hedera.BridgeAccount)
		operatorAccount := pw.getAccount(pw.configuration.Node.Clients.Hedera.Operator.AccountId)

		pw.payerAccountBalanceGauge.Set(getAccountBalance(payerAccount))
		pw.bridgeAccountBalanceGauge.Set(getAccountBalance(bridgeAccount))
		pw.operatorBalanceGauge.Set(getAccountBalance(operatorAccount))

		pw.setAssetsMetrics(bridgeAccount)

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

func getAccountBalance(account *model.AccountsResponse) float64 {
	balance := metrics.ConvertToHbar(account.Balance.Balance)
	log.Infof("The Account with ID [%s] has balance = %f", account.Account, balance)
	return balance
}

func (pw Watcher) setAssetsMetrics(bridgeAccount *model.AccountsResponse) {
	nativeToWrapped := pw.configuration.Bridge.Assets.GetNativeToWrapped()
	for nativeNetworkId, nativeMap := range nativeToWrapped {
		for nativeAsset, wrappedMap := range nativeMap {
			// set native assets balance
			pw.setAssetMetric(nativeNetworkId, nativeAsset, bridgeAccount, true)
			for wrappedNetworkId, wrappedAsset := range wrappedMap {
				//set wrapped assets total supply
				pw.setAssetMetric(wrappedNetworkId, wrappedAsset, bridgeAccount, false)
			}
		}
	}
}

func (pw Watcher) setAssetMetric(networkId int64,
	assetAddress string,
	bridgeAccount *model.AccountsResponse,
	isNative bool,
) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	if assetAddress != constants.Hbar { // skip HBAR
		assetMetric := pw.prometheusService.GetGauge(pw.assetsMetrics[networkId][assetAddress])
		value := pw.getAssetMetricValue(networkId, assetAddress, bridgeAccount, isNative)
		logStr := constants.SupplyAssetMetricsHelpPrefix
		if isNative {
			logStr = constants.BalanceAssetMetricHelpPrefix
		}
		setMetric(assetMetric, assetAddress, value, logStr)
	}
}

func (pw Watcher) getAssetMetricValue(
	networkId int64,
	assetAddress string,
	bridgeAccount *model.AccountsResponse,
	isNative bool,
) (value float64) {
	var (
		evmAssetInstance *wtoken.Wtoken
		decimal          uint8
	)
	if networkId != constants.HederaNetworkId { // EVM
		evm := pw.EVMClients[networkId].GetClient()
		wrappedTokenInstance, e := wtoken.NewWtoken(common.HexToAddress(assetAddress), evm)
		if e != nil {
			panic(e)
		}
		evmAssetInstance = wrappedTokenInstance
		dec, e := evmAssetInstance.Decimals(&bind.CallOpts{})
		if e != nil {
			panic(e)
		}
		decimal = dec
	}

	if networkId == constants.HederaNetworkId && isNative { // Hedera native balance
		value = pw.getHederaTokenBalance(assetAddress, bridgeAccount)
	} else if networkId == constants.HederaNetworkId && !isNative { // Hedera wrapped total supply
		value = pw.getHederaTokenSupply(assetAddress)
	} else if networkId != constants.HederaNetworkId && isNative { // EVM native balance
		value = pw.getEVMBalance(networkId, evmAssetInstance, decimal)
	} else { // EVM wrapped total supply
		value = pw.getEVMSupply(evmAssetInstance, decimal)
	}
	return value
}

func (pw Watcher) getHederaTokenBalance(assetAddress string, bridgeAccount *model.AccountsResponse) (value float64) {
	for _, token := range bridgeAccount.Balance.Tokens {
		if assetAddress == token.TokenID {
			asset, e := pw.mirrorNode.GetToken(assetAddress)
			if e != nil {
				panic(e)
			}
			dec, _ := strconv.Atoi(asset.Decimals)
			decimal := uint8(dec)

			b := big.NewInt(int64(token.Balance))
			balance, e := metrics.ConvertBasedOnDecimal(b, decimal)
			if e != nil {
				panic(e)
			}
			value = *balance
		}
	}
	return value
}

func (pw Watcher) getHederaTokenSupply(assetAddress string) float64 {
	asset, e := pw.mirrorNode.GetToken(assetAddress)
	if e != nil {
		panic(e)
	}
	dec, _ := strconv.Atoi(asset.Decimals)
	decimal := uint8(dec)

	ts, ok := new(big.Int).SetString(asset.TotalSupply, 10)
	if !ok {
		log.Infof(`"SetString: error": [%s].`, asset.TotalSupply)
	}
	totalSupply, e := metrics.ConvertBasedOnDecimal(ts, decimal)
	if e != nil {
		panic(e)
	}
	return *totalSupply
}

func (pw Watcher) getEVMBalance(networkId int64, evmAssetInstance *wtoken.Wtoken, decimal uint8) float64 {
	address := common.HexToAddress(pw.configuration.Bridge.EVMs[networkId].RouterContractAddress)

	b, e := evmAssetInstance.BalanceOf(&bind.CallOpts{}, address)
	if e != nil {
		panic(e)
	}
	balance, e := metrics.ConvertBasedOnDecimal(b, decimal)
	if e != nil {
		panic(e)
	}
	return *balance
}

func (pw Watcher) getEVMSupply(evmAssetInstance *wtoken.Wtoken, decimal uint8) float64 {
	ts, e := evmAssetInstance.TotalSupply(&bind.CallOpts{})
	if e != nil {
		panic(e)
	}
	totalSupply, e := metrics.ConvertBasedOnDecimal(ts, decimal)
	if e != nil {
		panic(e)
	}
	return *totalSupply
}

func setMetric(metric prometheus.Gauge, assetAddress string, value float64, logString string) {
	metric.Set(value)
	log.Infof("The Asset with ID [%s] has %s = %f", assetAddress, logString, value)
}
