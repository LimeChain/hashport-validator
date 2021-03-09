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
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
)

// Scheduler provides the required scheduling logic for submitting Ethereum transactions using a slot-based algorithm
type Scheduler interface {
	// Schedule - Schedules new Transaction for execution at the right leader elected slot
	Schedule(id string, submission ethsubmission.Submission) error
	// Cancel - Removes and cancels an already scheduled Transaction
	Cancel(id string) error
}
