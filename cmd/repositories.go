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

package main

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/database"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/fee"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
)

// Repositories struct holding the referenced repositories
type Repositories struct {
	transferStatus repository.Status
	messageStatus  repository.Status
	transfer       repository.Transfer
	message        repository.Message
	burnEvent      repository.BurnEvent
	lockEvent      repository.LockEvent
	fee            repository.Fee
}

// PrepareRepositories initialises connection to the Database and instantiates the repositories
func PrepareRepositories(db database.Database) *Repositories {
	connection := db.GetConnection()
	return &Repositories{
		transferStatus: status.NewRepositoryForStatus(connection, status.Transfer),
		messageStatus:  status.NewRepositoryForStatus(connection, status.Message),
		transfer:       transfer.NewRepository(connection),
		message:        message.NewRepository(connection),
		burnEvent:      burn_event.NewRepository(connection),
		// TODO: Define Lock Event interface
		lockEvent: lock_event.NewRepository(connection),
		fee:       fee.NewRepository(connection),
	}
}
