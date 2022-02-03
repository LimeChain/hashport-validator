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

package metrics

import (
	"errors"
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"strings"
)

func PrepareIdForPrometheus(id string) string {
	for symbolToReplace, repetitions := range constants.PrometheusNotAllowedSymbolsWithRepetitions {
		id = strings.Replace(id, symbolToReplace, constants.NotAllowedSymbolsReplacement, repetitions)
	}

	return id
}

func ConstructNameForMetric(sourceNetworkId, targetNetworkId uint64, tokenType, transactionId, metricTarget string) (string, error) {
	errMsg := "Network id %v is missing in id to name mapping."
	sourceNetworkName, exist := constants.NetworksById[sourceNetworkId]
	if !exist {
		return "", errors.New(fmt.Sprintf(errMsg, sourceNetworkId))
	}
	targetNetworkName, exist := constants.NetworksById[targetNetworkId]
	if !exist {
		return "", errors.New(fmt.Sprintf(errMsg, targetNetworkId))
	}

	transactionId = PrepareIdForPrometheus(transactionId)

	return fmt.Sprintf("%s_%s_to_%s_%s_%s", tokenType, sourceNetworkName, targetNetworkName, transactionId, metricTarget), nil
}
