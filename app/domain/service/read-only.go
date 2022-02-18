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
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
)

type ReadOnly interface {
	FindTransfer(transferID string, fetch func() (*mirror_node.Response, error), save func(transactionID, scheduleID, status string) error)
	FindAssetTransfer(transferID string, asset string, transfers []model.Hedera, fetch func() (*mirror_node.Response, error), save func(transactionID, scheduleID, status string) error)
	FindNftTransfer(transferID string, tokenID string, serialNum int64, sender string, receiver string,
		save func(transactionID, scheduleID, status string) error)
}
