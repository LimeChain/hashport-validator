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

package constants

// Prometheus metrics
const (
	ValidatorsParticipationRateInitialValue = 100
	ValidatorsParticipationRateGaugeName    = "validators_participation_rate"
	ValidatorsParticipationRateGaugeHelp    = "Participation rate: Track validators' activity in %."
	FeeAccountAmountGaugeName               = "fee_account_amount"
	FeeAccountAmountGaugeHelp               = "Fee account amount."
	BridgeAccountAmountGaugeName            = "bridge_account_amount"
	BridgeAccountAmountGaugeHelp            = "Bridge account amount."
	OperatorAccountAmountName               = "operator_account_amount"
	OperatorAccountAmountHelp               = "Operator account amount."
	DotSymbol                               = "." // not fit prometheus validation https://github.com/prometheus/common/blob/main/model/metric.go#L97
	DashSymbol                              = "-" // not fit prometheus validation https://github.com/prometheus/common/blob/main/model/metric.go#L97
	OpenSquareBracket                       = "[" // not fit prometheus validation https://github.com/prometheus/common/blob/main/model/metric.go#L97
	CloseSquareBracket                      = "]" // not fit prometheus validation https://github.com/prometheus/common/blob/main/model/metric.go#L97
	Space                                   = " " // not fit prometheus validation https://github.com/prometheus/common/blob/main/model/metric.go#L97
	NotAllowedSymbolsReplacement            = "_"
	DotSymbolRep                            = 2
	DashSymbolRep                           = 2
	NoLimitRep                              = -1
	AssetMetricsNamePrefix                  = "asset_id_"
	SupplyAssetMetricNameSuffix             = "_total_supply_"
	SupplyAssetMetricsHelpPrefix            = "total supply"
	BalanceAssetMetricNameSuffix            = "_balance_"
	BalanceAssetMetricHelpPrefix            = "balance"
	AssetMetricLabelKey                     = "symbol"
	AccountMetricLabelKey                   = "account_id"

	CreateDecimalPrefix = "1"
	CreateDecimalRepeat = "0"

	// API

	PrometheusMetricsEndpoint = "/metrics"

	// Success Rate Metrics //

	MajorityReachedNameSuffix  = "majority_reached"
	MajorityReachedHelp        = "Majority reached for hedera transaction."
	FeeTransferredNameSuffix   = "fee_transferred"
	FeeTransferredHelp         = "Fee transferred to the bridge account."
	UserGetHisTokensNameSuffix = "user_get_his_tokens"
	UserGetHisTokensHelp       = "The user get his tokens after bridging."
)

var (
	PrometheusNotAllowedSymbolsWithRepetitions = map[string]int{
		DotSymbol:          DotSymbolRep,
		DashSymbol:         DashSymbolRep,
		OpenSquareBracket:  NoLimitRep,
		CloseSquareBracket: NoLimitRep,
		Space:              NoLimitRep,
	}
)
