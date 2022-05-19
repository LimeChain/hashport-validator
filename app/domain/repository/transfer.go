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

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
)

type Transfer interface {
	// Returns Transfer. Returns nil if not found
	GetByTransactionId(txId string) (*entity.Transfer, error)
	// Returns Transfer with preloaded Fee table. Returns nil if not found
	GetWithFee(txId string) (*entity.Transfer, error)
	GetWithPreloads(txId string) (*entity.Transfer, error)
	UpdateFee(txId string, fee string) error

	Create(ct *payload.Transfer) (*entity.Transfer, error)
	UpdateStatusCompleted(txId string) error
	UpdateStatusFailed(txId string) error
	Paged(req *transfer.PagedRequest) ([]*entity.Transfer, error)
}
