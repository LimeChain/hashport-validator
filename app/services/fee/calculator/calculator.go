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
	feePercentage int64
	logger        *log.Entry
}

func New(feePercentage int64) *Service {
	if feePercentage < MinPercentage || feePercentage > MaxPercentage {
		log.Fatalf("Invalid fee percentage: [%d]", feePercentage)
	}

	return &Service{
		feePercentage: feePercentage,
		logger:        config.GetLoggerFor("Fee Service")}
}

// CalculateFee calculates the fee and remainder of a given amount
func (s Service) CalculateFee(amount int64) (fee, remainder int64) {
	fee = amount * s.feePercentage / MaxPercentage
	remainder = amount - fee

	totalAmount := remainder + fee
	if totalAmount != amount {
		remainder += amount - totalAmount
	}

	return fee, remainder
}
