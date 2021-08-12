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

package service

import (
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
)

// LockEvent is the major service used for processing BurnEvent operations
type LockEvent interface {
	// ProcessEvent processes the lock event by submitting the appropriate
	// Scheduled Token Mint and Transfer transactions
	ProcessEvent(event lock_event.LockEvent)
}
