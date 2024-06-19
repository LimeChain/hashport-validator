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

package calculator

import (
	"errors"
	"fmt"

	"github.com/gookit/event"
	bridge_config_event "github.com/limechain/hedera-eth-bridge-validator/app/model/bridge-config-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	feePercentages map[string]int64
	logger         *log.Entry
}

func New(feePercentages map[string]int64) *Service {
	for token, fee := range feePercentages {
		if fee < constants.FeeMinPercentage || fee > constants.FeeMaxPercentage {
			log.Fatalf("[%s] Invalid fee percentage: [%d]", token, fee)
		}
	}
	instance := &Service{
		feePercentages: feePercentages,
		logger:         config.GetLoggerFor("Fee Service"),
	}
	event.On(constants.EventBridgeConfigUpdate, event.ListenerFunc(func(e event.Event) error {
		return bridgeCfgUpdateEventHandler(e, instance)
	}), constants.AssetServicePriority)

	return instance
}

// CalculateFee calculates the fee and remainder of a given token and amount
func (s Service) CalculateFee(token string, amount int64) (fee, remainder int64) {
	fee = amount * s.feePercentages[token] / constants.FeeMaxPercentage
	remainder = amount - fee

	totalAmount := remainder + fee
	if totalAmount != amount {
		remainder += amount - totalAmount
	}

	return fee, remainder
}

func bridgeCfgUpdateEventHandler(e event.Event, instance *Service) error {
	params, ok := e.Get(constants.BridgeConfigUpdateEventParamsKey).(*bridge_config_event.Params)
	if !ok {
		errMsg := fmt.Sprintf("failed to cast params from event [%s]", constants.EventBridgeConfigUpdate)
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}

	instance.feePercentages = params.Bridge.Hedera.FeePercentages

	return nil
}
