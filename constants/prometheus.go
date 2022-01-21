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

package constants

// Prometheus metrics
const (
	ValidatorsParticipationRateGaugeName = "validators_participation_rate"
	ValidatorsParticipationRateGaugeHelp = "Participation rate: Track validators' activity in %."
	FeeAccountAmountGaugeName            = "fee_account_amount"
	FeeAccountAmountGaugeHelp            = "Fee account amount."
	BridgeAccountAmountGaugeName         = "bridge_account_amount"
	BridgeAccountAmountGaugeHelp         = "Bridge account amount."
	OperatorAccountAmountName            = "operator_account_amount"
	OperatorAccountAmountHelp            = "Operator account amount."
	DotSymbol                            = "." //not fit prometheus validation https://github.com/prometheus/common/blob/main/model/metric.go#L97
	ReplaceDotSymbol                     = "_"
	DotSymbolRep                         = 2
	AssetMetricsNamePrefix               = "asset_id_"
	SupplyAssetMetricNamePrefix          = "total_supply_"
	SupplyAssetMetricsHelpPrefix         = "The total supply of "
	BridgeAccAssetMetricsNamePrefix      = "bridge_acc_"
	BridgeAccAssetMetricsNameHelp        = "Bridge account balance for "
	BalanceAssetMetricNamePrefix         = "balance_"
	BalanceAssetMetricHelpPrefix         = "The balance of "
	AssetMetricHelpSuffix                = " at router address "
)
