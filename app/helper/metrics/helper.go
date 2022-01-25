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
	sourceNetworkName, exist := constants.NetworkIdToName[sourceNetworkId]
	if !exist {
		return "", errors.New(fmt.Sprintf(errMsg, sourceNetworkId))
	}
	targetNetworkName, exist := constants.NetworkIdToName[targetNetworkId]
	if !exist {
		return "", errors.New(fmt.Sprintf(errMsg, targetNetworkId))
	}

	transactionId = PrepareIdForPrometheus(transactionId)

	return fmt.Sprintf("%s_%s_to_%s_%s_%s", tokenType, sourceNetworkName, targetNetworkName, transactionId, metricTarget), nil
}
