/*
 * Copyright 2022 LimeChain Ltd.
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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
	"strings"
	"time"
)

var (
	whiteSpacesPattern = regexp.MustCompile(constants.WhiteSpacesPattern)
)

type Watcher struct {
	dashboardPolling           time.Duration
	mirrorNode                 client.MirrorNode
	EvmFungibleTokenClients    map[uint64]map[string]client.EvmFungibleToken
	EvmNonFungibleTokenClients map[uint64]map[string]client.EvmNft
	configuration              config.Config
	prometheusService          service.Prometheus
	logger                     *log.Entry
	bridgeAccountBalanceGauge  prometheus.Gauge
	payerAccountBalanceGauge   prometheus.Gauge
	// A mapping, storing all Prometheus Gauges by AccountId
	monitoredAccountsGauges map[string]monitoredAccountsInfo
	// A mapping, storing all network ID - asset address - metric name
	assetsMetrics map[uint64]map[string]string
	assetsService service.Assets
}

type monitoredAccountsInfo struct {
	AccountId string
	Gauge     prometheus.Gauge
}

func NewWatcher(
	dashboardPolling time.Duration,
	mirrorNode client.MirrorNode,
	configuration config.Config,
	prometheusService service.Prometheus,
	EvmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken,
	EvmNonFungibleTokenClients map[uint64]map[string]client.EvmNft,
	assetsService service.Assets,
) *Watcher {

	return &Watcher{
		dashboardPolling:           dashboardPolling,
		mirrorNode:                 mirrorNode,
		EvmFungibleTokenClients:    EvmFungibleTokenClients,
		EvmNonFungibleTokenClients: EvmNonFungibleTokenClients,
		configuration:              configuration,
		prometheusService:          prometheusService,
		logger:                     config.GetLoggerFor(fmt.Sprintf("Prometheus Metrics Watcher on interval [%s]", dashboardPolling)),
		payerAccountBalanceGauge:   initPayerAccountBalanceGauge(prometheusService, configuration),
		bridgeAccountBalanceGauge:  initBridgeAccountBalanceGauge(prometheusService, configuration),
		monitoredAccountsGauges:    initMonitoredAccountsGauges(configuration, prometheusService),
		assetsMetrics:              make(map[uint64]map[string]string),
		assetsService:              assetsService,
	}
}

func initBridgeAccountBalanceGauge(prometheusService service.Prometheus, configuration config.Config) prometheus.Gauge {
	bridgeAccountBalanceGauge := prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
		Name: constants.BridgeAccountAmountGaugeName,
		Help: constants.BridgeAccountAmountGaugeHelp,
		ConstLabels: prometheus.Labels{
			constants.AccountMetricLabelKey: configuration.Bridge.Hedera.BridgeAccount,
		},
	})
	return bridgeAccountBalanceGauge
}

func initPayerAccountBalanceGauge(prometheusService service.Prometheus, configuration config.Config) prometheus.Gauge {
	payerAccountBalanceGauge := prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
		Name: constants.FeeAccountAmountGaugeName,
		Help: constants.FeeAccountAmountGaugeHelp,
		ConstLabels: prometheus.Labels{
			constants.AccountMetricLabelKey: configuration.Bridge.Hedera.PayerAccount,
		},
	})
	return payerAccountBalanceGauge
}

func initMonitoredAccountsGauges(configuration config.Config, prometheusService service.Prometheus) map[string]monitoredAccountsInfo {
	monitoredAccountsGauges := make(map[string]monitoredAccountsInfo)

	for name, accountId := range configuration.Bridge.MonitoredAccounts {
		preparedGaugeName := whiteSpacesPattern.ReplaceAllString(strings.ToLower(name), constants.NotAllowedSymbolsReplacement)
		preparedGaugeName = constants.AccountBalanceGaugeNamePrefix + metrics.PrepareValueForPrometheusMetricName(preparedGaugeName)
		gaugeHelp := constants.AccountBalanceGaugeHelpPrefix + accountId
		monitoredAccountInfo := monitoredAccountsInfo{
			AccountId: accountId,
			Gauge: prometheusService.CreateGaugeIfNotExists(prometheus.GaugeOpts{
				Name: preparedGaugeName,
				Help: gaugeHelp,
				ConstLabels: prometheus.Labels{
					constants.NameMetricLabelKey:    name,
					constants.AccountMetricLabelKey: accountId,
				},
			}),
		}
		monitoredAccountsGauges[name] = monitoredAccountInfo
	}

	return monitoredAccountsGauges
}

func (pw Watcher) Watch(q qi.Queue) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		pw.logger.Warnf("Tried to executed Prometheus watcher, when monitoring is not enabled.")
		return
	}

	// there will be no handler, so the q is to implement the interface
	go pw.beginWatching()
}

func (pw Watcher) beginWatching() {
	//The queue will be not used
	pw.registerAllAssetsMetrics()
	pw.setMetrics()
}

func (pw Watcher) registerAllAssetsMetrics() {
	fungibleAssets := pw.assetsService.FungibleNetworkAssets()
	nonFungibleAssets := pw.assetsService.NonFungibleNetworkAssets()
	pw.registerAssetMetrics(fungibleAssets, true)
	pw.registerAssetMetrics(nonFungibleAssets, false)
}

func (pw Watcher) registerAssetMetrics(assets map[uint64][]string, isFungible bool) {
	for networkId, networkAssets := range assets {
		for _, assetAddress := range networkAssets {
			if pw.assetsService.IsNative(networkId, assetAddress) { // native
				// register native assets balance
				pw.registerAssetMetric(
					networkId,
					networkId,
					assetAddress,
					constants.BalanceAssetMetricNameSuffix,
					constants.BalanceAssetMetricHelpPrefix,
					isFungible,
				)
				wrappedFromNative := pw.assetsService.WrappedFromNative(networkId, assetAddress)
				for wrappedNetworkId, wrappedAssetAddress := range wrappedFromNative {
					//register wrapped assets total supply
					pw.registerAssetMetric(
						networkId,
						wrappedNetworkId,
						wrappedAssetAddress,
						constants.SupplyAssetMetricNameSuffix,
						constants.SupplyAssetMetricsHelpPrefix,
						isFungible,
					)
				}
			}
		}
	}
}

func (pw Watcher) registerAssetMetric(
	nativeNetworkId,
	wrappedNetworkId uint64,
	assetAddress string,
	metricNameCnt string,
	metricHelpCnt string,
	isFungible bool,
) {
	if assetAddress != constants.Hbar { // skip HBAR
		var (
			name, symbol string
		)
		if isFungible {
			assetInfo, exist := pw.assetsService.FungibleAssetInfo(wrappedNetworkId, assetAddress)
			if !exist {
				return
			}
			name = assetInfo.Name
			symbol = assetInfo.Symbol
		} else {
			assetInfo, exist := pw.assetsService.NonFungibleAssetInfo(wrappedNetworkId, assetAddress)
			if !exist {
				return
			}
			name = assetInfo.Name
			symbol = assetInfo.Symbol
		}

		metricName, metricHelp := getMetricData(
			nativeNetworkId,
			wrappedNetworkId,
			assetAddress,
			name,
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
				constants.AssetMetricLabelKey: symbol,
			},
		})
	}
}

func getMetricData(
	nativeNetworkId,
	wrappedNetworkId uint64,
	assetAddress string,
	assetName string,
	metricNameSuffix string,
	metricsHelpPrefix string,
) (string, string) {

	nativeNetworkName := constants.NetworksById[nativeNetworkId]
	wrappedNetworkName := constants.NetworksById[wrappedNetworkId]
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

	for {
		payerAccount, errPayerAcc := pw.getAccount(pw.configuration.Bridge.Hedera.PayerAccount)
		if errPayerAcc == nil {
			pw.payerAccountBalanceGauge.Set(pw.getAccountBalance(payerAccount))
		}
		hbarAssetInfo, ok := pw.assetsService.FungibleAssetInfo(constants.HederaNetworkId, constants.Hbar)
		if ok {
			pw.bridgeAccountBalanceGauge.Set(metrics.ConvertToHbar(int(hbarAssetInfo.ReserveAmount.Int64())))
		}

		for _, monitoredAccountInfo := range pw.monitoredAccountsGauges {
			monitoredAccount, monitoredAccErr := pw.getAccount(monitoredAccountInfo.AccountId)
			if monitoredAccErr == nil {
				monitoredAccountInfo.Gauge.Set(pw.getAccountBalance(monitoredAccount))
			}
		}

		pw.setAllAssetsMetrics()

		pw.logger.Infoln("Dashboard Polling interval: ", pw.dashboardPolling)
		time.Sleep(pw.dashboardPolling)
	}
}

func (pw Watcher) getAccount(accountId string) (*account.AccountsResponse, error) {
	account, e := pw.mirrorNode.GetAccount(accountId)
	if e != nil {
		pw.logger.Errorf("Hedera Mirror Node for Account ID [%s] method GetAccount - Error: [%s]", accountId, e)
		return nil, e
	}
	return account, nil
}

func (pw Watcher) getAccountBalance(account *account.AccountsResponse) float64 {
	balance := metrics.ConvertToHbar(account.Balance.Balance)
	pw.logger.Infof("The Account with ID [%s] has balance = %f", account.Account, balance)
	return balance
}

func (pw Watcher) setAllAssetsMetrics() {
	fungibleAssets := pw.assetsService.FungibleNetworkAssets()
	nonFungibleAssets := pw.assetsService.NonFungibleNetworkAssets()
	pw.setAssetsMetrics(fungibleAssets, true)
	pw.setAssetsMetrics(nonFungibleAssets, false)
}

func (pw Watcher) setAssetsMetrics(assets map[uint64][]string, isFungible bool) {
	for networkId, networkAssets := range assets {
		for _, assetAddress := range networkAssets {
			if assetAddress == constants.Hbar { // skip HBAR
				continue
			}
			isNative := pw.assetsService.IsNative(networkId, assetAddress)
			pw.prepareAndSetAssetMetric(networkId, assetAddress, isFungible, isNative)
		}
	}
}

func (pw Watcher) prepareAndSetAssetMetric(networkId uint64,
	assetAddress string,
	isFungible,
	isNative bool,
) {
	var value float64
	assetMetric := pw.prometheusService.GetGauge(pw.assetsMetrics[networkId][assetAddress])
	var (
		ReserveAmountInLowestDenomination *big.Int
		decimals                          uint8
	)
	if isFungible {
		assetInfo, ok := pw.assetsService.FungibleAssetInfo(networkId, assetAddress)
		if ok {
			ReserveAmountInLowestDenomination = assetInfo.ReserveAmount
			decimals = assetInfo.Decimals
		}
	} else {
		decimals = 0
		assetInfo, ok := pw.assetsService.NonFungibleAssetInfo(networkId, assetAddress)
		if ok {
			ReserveAmountInLowestDenomination = assetInfo.ReserveAmount
		}
	}

	if decimals != 0 {
		converted, err := metrics.ConvertBasedOnDecimal(ReserveAmountInLowestDenomination, decimals)
		if err != nil {
			pw.logger.Errorf("Failed to convert asset ReserveAmount to decimal. Error: %s", err)
			return
		}
		value = *converted
	} else {
		value, _ = new(big.Float).SetInt(ReserveAmountInLowestDenomination).Float64()
	}

	logString := constants.SupplyAssetMetricsHelpPrefix
	if isNative {
		logString = constants.BalanceAssetMetricHelpPrefix
	}
	assetMetric.Set(value)
	pw.logger.Infof("The Assets with ID [%s] has %s = %f", assetAddress, logString, value)
}
