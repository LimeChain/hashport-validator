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

package calculator

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

const MaxPercentage = 100000
const MinPercentage = 0

type Service struct {
	feePercentages map[string]int64
	logger         *log.Entry
}

func New(feePercentages map[string]int64) *Service {
	for token, fee := range feePercentages {
		if fee < MinPercentage || fee > MaxPercentage {
			log.Fatalf("[%s] Invalid fee percentage: [%d]", token, fee)
		}
	}

	return &Service{
		feePercentages: feePercentages,
		logger:         config.GetLoggerFor("Fee Service")}
}

// CalculateFee calculates the fee and remainder of a given token and amount
func (s Service) CalculateFee(token string, amount int64) (fee, remainder int64) {
	fee = amount * s.feePercentages[token] / MaxPercentage
	remainder = amount - fee

	totalAmount := remainder + fee
	if totalAmount != amount {
		remainder += amount - totalAmount
	}

	return fee, remainder
}
