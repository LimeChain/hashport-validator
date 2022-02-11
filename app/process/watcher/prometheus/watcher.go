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
	logger                    *log.Entry
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
		logger:                    config.GetLoggerFor(fmt.Sprintf("Prometheus Metrics Watcher on interval [%s]", dashboardPolling)),
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
			pw.registerAssetMetric(
				nativeNetworkId,
				nativeNetworkId,
				nativeAsset,
				constants.BalanceAssetMetricNameSuffix,
				constants.BalanceAssetMetricHelpPrefix,
			)
			for wrappedNetworkId, wrappedAsset := range wrappedMap {
				//register wrapped assets total supply
				pw.registerAssetMetric(
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

func (pw Watcher) registerAssetMetric(
	nativeNetworkId int64,
	wrappedNetworkId int64,
	assetAddress string,
	metricNameCnt string,
	metricHelpCnt string,
) {
	if assetAddress != constants.Hbar { // skip HBAR
		assetName, assetSymbol, e := pw.getAssetData(wrappedNetworkId, assetAddress)
		if e != nil {
			return
		}
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

func (pw Watcher) getAssetData(networkId int64, assetAddress string) (name string, symbol string, err error) {
	if networkId == constants.HederaNetworkId { // Hedera
		asset, e := pw.mirrorNode.GetToken(assetAddress)
		if e != nil {
			pw.logger.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetAddress, e)
			return "", "", e
		}
		name = asset.Name
		symbol = asset.Symbol
	} else { // EVM
		evm := pw.EVMClients[networkId].GetClient()
		evmAssetInstance, e := wtoken.NewWtoken(common.HexToAddress(assetAddress), evm)
		if e != nil {
			pw.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method NewWtoken - Error: [%s]", networkId, assetAddress, e)
			return "", "", e
		}
		resName, e := evmAssetInstance.Name(&bind.CallOpts{})
		if e != nil {
			pw.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Name - Error: [%s]", networkId, assetAddress, e)
			return "", "", e
		}
		name = resName
		resSymbol, e := evmAssetInstance.Symbol(&bind.CallOpts{})
		if e != nil {
			pw.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Symbol - Error: [%s]", networkId, assetAddress, e)
			return "", "", e
		}
		symbol = resSymbol
	}
	return name, symbol, nil
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
		payerAccount, errPayerAcc := pw.getAccount(pw.configuration.Bridge.Hedera.PayerAccount)
		bridgeAccount, errBridgeAcc := pw.getAccount(pw.configuration.Bridge.Hedera.BridgeAccount)
		operatorAccount, errOperatorAcc := pw.getAccount(pw.configuration.Node.Clients.Hedera.Operator.AccountId)

		if errPayerAcc == nil {
			pw.payerAccountBalanceGauge.Set(pw.getAccountBalance(payerAccount))
		}
		if errBridgeAcc == nil {
			pw.bridgeAccountBalanceGauge.Set(pw.getAccountBalance(bridgeAccount))
		}
		if errOperatorAcc == nil {
			pw.operatorBalanceGauge.Set(pw.getAccountBalance(operatorAccount))
		}

		pw.setAssetsMetrics(bridgeAccount)

		pw.logger.Infoln("Dashboard Polling interval: ", pw.dashboardPolling)
		time.Sleep(pw.dashboardPolling)
	}
}

func (pw Watcher) getAccount(accountId string) (*model.AccountsResponse, error) {
	account, e := pw.mirrorNode.GetAccount(accountId)
	if e != nil {
		pw.logger.Errorf("Hedera Mirror Node for Account ID [%s] method GetAccount - Error: [%s]", accountId, e)
		return nil, e
	}
	return account, nil
}

func (pw Watcher) getAccountBalance(account *model.AccountsResponse) float64 {
	balance := metrics.ConvertToHbar(account.Balance.Balance)
	pw.logger.Infof("The Account with ID [%s] has balance = %f", account.Account, balance)
	return balance
}

func (pw Watcher) setAssetsMetrics(bridgeAccount *model.AccountsResponse) {
	nativeToWrapped := pw.configuration.Bridge.Assets.GetNativeToWrapped()
	for nativeNetworkId, nativeMap := range nativeToWrapped {
		for nativeAsset, wrappedMap := range nativeMap {
			// set native assets balance
			pw.prepareAndSetAssetMetric(nativeNetworkId, nativeAsset, bridgeAccount, true)
			for wrappedNetworkId, wrappedAsset := range wrappedMap {
				//set wrapped assets total supply
				pw.prepareAndSetAssetMetric(wrappedNetworkId, wrappedAsset, bridgeAccount, false)
			}
		}
	}
}

func (pw Watcher) prepareAndSetAssetMetric(networkId int64,
	assetAddress string,
	bridgeAccount *model.AccountsResponse,
	isNative bool,
) {
	if !pw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	if assetAddress != constants.Hbar { // skip HBAR
		assetMetric := pw.prometheusService.GetGauge(pw.assetsMetrics[networkId][assetAddress])
		value, e := pw.getAssetMetricValue(networkId, assetAddress, bridgeAccount, isNative)
		if e != nil {
			pw.logger.Errorf("Network ID [%d] and asset [%s] for getAssetMetricValue Error: [%s]", networkId, assetAddress, e)
			return
		}
		logString := constants.SupplyAssetMetricsHelpPrefix
		if isNative {
			logString = constants.BalanceAssetMetricHelpPrefix
		}

		assetMetric.Set(value)
		pw.logger.Infof("The Asset with ID [%s] has %s = %f", assetAddress, logString, value)
	}
}

func (pw Watcher) getAssetMetricValue(
	networkId int64,
	assetAddress string,
	bridgeAccount *model.AccountsResponse,
	isNative bool,
) (value float64, err error) {
	var (
		evmAssetInstance *wtoken.Wtoken
		decimal          uint8
	)
	if networkId != constants.HederaNetworkId { // EVM
		evm := pw.EVMClients[networkId].GetClient()
		wrappedTokenInstance, e := wtoken.NewWtoken(common.HexToAddress(assetAddress), evm)
		if e != nil {
			pw.logger.Errorf("EVM with networkId [%d] for asset [%s], and method NewWtoken - Error: [%s]", networkId, assetAddress, e)
			return 0, e
		}
		evmAssetInstance = wrappedTokenInstance
		dec, e := evmAssetInstance.Decimals(&bind.CallOpts{})
		if e != nil {
			pw.logger.Errorf("EVM with networkId [%d] for asset [%s], and method Decimals - Error: [%s]", networkId, assetAddress, e)
			return 0, e
		}
		decimal = dec
	}

	if networkId == constants.HederaNetworkId { //Hedera
		if isNative { // Hedera native balance
			value, err = pw.getHederaTokenBalance(assetAddress, bridgeAccount)
		} else { // Hedera wrapped total supply
			value, err = pw.getHederaTokenSupply(assetAddress)
		}
	} else { // EVM
		if isNative { // EVM native balance
			value, err = pw.getEVMBalance(networkId, evmAssetInstance, decimal, assetAddress)
		} else { // EVM wrapped total supply
			value, err = pw.getEVMSupply(evmAssetInstance, decimal, networkId, assetAddress)
		}
	}
	return value, err
}

func (pw Watcher) getHederaTokenBalance(assetAddress string, bridgeAccount *model.AccountsResponse) (value float64, err error) {
	if bridgeAccount == nil {
		return 0, errors.New(fmt.Sprintf("Bridge account cannot be nil"))
	}
	for _, token := range bridgeAccount.Balance.Tokens {
		if assetAddress == token.TokenID {
			asset, e := pw.mirrorNode.GetToken(assetAddress)
			if e != nil {
				pw.logger.Errorf("Hedera Mirror Node for asset [%s] method GetToken - Error: [%s]", assetAddress, e)
				return 0, e
			}
			dec, e := strconv.Atoi(asset.Decimals)
			if e != nil {
				pw.logger.Errorf("Hedera asset [%s] convert decimals to string method Atio - Error: [%s]", assetAddress, e)
				return 0, e
			}
			decimal := uint8(dec)

			b := big.NewInt(int64(token.Balance))
			balance, e := metrics.ConvertBasedOnDecimal(b, decimal)
			if e != nil {
				pw.logger.Errorf("Hedera asset [%s] balance ConvertBasedOnDecimal - Error: [%s]", assetAddress, e)
				return 0, e
			}
			value = *balance
		}
	}
	return value, nil
}

func (pw Watcher) getHederaTokenSupply(assetAddress string) (float64, error) {
	asset, e := pw.mirrorNode.GetToken(assetAddress)
	if e != nil {
		pw.logger.Errorf("Hedera Mirror Node for asset [%s] method GetToken - Error: [%s]", assetAddress, e)
		return 0, e
	}
	dec, e := strconv.Atoi(asset.Decimals)
	if e != nil {
		pw.logger.Errorf("Hedera asset [%s] convert decimals to string, method Atio - Error: [%s]", assetAddress, e)
		return 0, e
	}
	decimal := uint8(dec)

	ts, ok := new(big.Int).SetString(asset.TotalSupply, 10)
	if !ok {
		pw.logger.Errorf(`"Hedera assed [%s] total supply SetString - Error": [%s].`, assetAddress, asset.TotalSupply)
		return 0, e
	}
	totalSupply, e := metrics.ConvertBasedOnDecimal(ts, decimal)
	if e != nil {
		pw.logger.Errorf("Hedera asset [%s] total supply ConvertBasedOnDecimal - Error: [%s]", assetAddress, e)
		return 0, e
	}
	return *totalSupply, nil
}

func (pw Watcher) getEVMBalance(
	networkId int64,
	evmAssetInstance *wtoken.Wtoken,
	decimal uint8,
	assetAddress string,
) (float64, error) {
	address := common.HexToAddress(pw.configuration.Bridge.EVMs[networkId].RouterContractAddress)

	b, e := evmAssetInstance.BalanceOf(&bind.CallOpts{}, address)
	if e != nil {
		pw.logger.Errorf("EVM with networkId [%d] for asset [%s], and method BalanceOf - Error: [%s]", networkId, assetAddress, e)
		return 0, e
	}
	balance, e := metrics.ConvertBasedOnDecimal(b, decimal)
	if e != nil {
		pw.logger.Errorf("EVM with networkId [%d] asset [%s] for balance, and method ConvertBasedOnDecimal - Error: [%s]", networkId, assetAddress, e)
		return 0, e
	}
	return *balance, nil
}

func (pw Watcher) getEVMSupply(evmAssetInstance *wtoken.Wtoken,
	decimal uint8,
	networkId int64,
	assetAddress string,
) (float64, error) {
	ts, e := evmAssetInstance.TotalSupply(&bind.CallOpts{})
	if e != nil {
		pw.logger.Errorf("EVM networkId [%d] asset [%s] for method TotalSupply - Error: [%s]", networkId, assetAddress, e)
		return 0, e
	}
	totalSupply, e := metrics.ConvertBasedOnDecimal(ts, decimal)
	if e != nil {
		pw.logger.Errorf("EVM networkId [%d] asset [%s] for total supply method ConvertBasedOnDecimal - Error: [%s]", networkId, assetAddress, e)
		return 0, e
	}
	return *totalSupply, nil
}
