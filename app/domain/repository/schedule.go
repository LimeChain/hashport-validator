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

package repository

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"

type Schedule interface {
	// Returns Schedule. Returns nil if not found
	Get(txId string) (*entity.Schedule, error)
	Create(entity *entity.Schedule) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusFailed(txId string) error
	GetReceiverTransferByTransactionID(id string) (*entity.Schedule, error)
	GetAllSubmittedIds() ([]*entity.Schedule, error)
}
