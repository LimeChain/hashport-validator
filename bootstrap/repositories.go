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

package bootstrap

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/database"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
)

// Repositories struct holding the referenced repositories
type Repositories struct {
	TransferStatus repository.Status
	MessageStatus  repository.Status
	Transfer       repository.Transfer
	Message        repository.Message
	Fee            repository.Fee
	Schedule       repository.Schedule
}

// PrepareRepositories initialises connection to the Database and instantiates the repositories
func PrepareRepositories(db database.Database) *Repositories {
	connection := db.GetConnection()
	return &Repositories{
		TransferStatus: status.NewRepositoryForStatus(connection, status.Transfer),
		MessageStatus:  status.NewRepositoryForStatus(connection, status.Message),
		Transfer:       transfer.NewRepository(connection),
		Message:        message.NewRepository(connection),
		Fee:            fee.NewRepository(connection),
		Schedule:       schedule.NewRepository(connection),
	}
}
