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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
)

// Distributor interface is implemented by the Distributor Service
// Handles distribution of proportional amounts to members
type Distributor interface {
	// PrepareTransfers Returns an equally divided array of transfers to each member
	PrepareTransfers(fee int64, token string) ([]model.Transfer, error)
	// CalculateMemberDistribution Returns an equally divided to each member
	CalculateMemberDistribution(validFee int64) ([]transfer.Hedera, error)
	// ValidAmount Returns the closest amount, which can be equally divided to members
	ValidAmount(amount int64) int64
}
