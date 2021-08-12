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

package lock_event

const (
	// StatusCompleted is a status set once the LockEvent operation is successfully completed.
	// This is a terminal status
	StatusCompleted = "COMPLETED"
	// StatusFailed is a status set once the LockEvent operation has failed
	// This is a terminal status
	StatusFailed = "FAILED"
	// StatusInitial is the initial status on LockEvent Record creation
	StatusInitial = "INITIAL"
	// StatusTransferSubmitted is set once the Hedera Scheduled Token Transfer Transaction (Create/Sign)
	StatusTransferSubmitted = "TOKEN_TRANSFER_SUBMITTED"
	// StatusSubmitted is set once the Hedera Scheduled Token Mint Transaction (Create/Sign)
	StatusMintSubmitted = "TOKEN_MINT_SUBMITTED"
	// StatusMintCompleted is set once the Hedera Scheduled Token Mint Transaction (Create/Sign)
	StatusMintCompleted = "TOKEN_MINT_COMPLETED"
)
