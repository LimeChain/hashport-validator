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

package services

import (
	hederaAPIModel "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
)

// Bridge is the major service used for processing Bridge operations
type Bridge interface {
	// SanityCheck performs any validation required prior to handling the transaction
	// (memo, state proof verification)
	SanityCheck(tx hederaAPIModel.Transaction) (*encoding.Memo, error)
	// SaveRecoveredTxn creates new Transaction record persisting the recovered Transfer TXn
	SaveRecoveredTxn(txId, amount string, m encoding.Memo) error
	// InitiateNewTransfer Stores the incoming transfer message into the Database
	// aware of already processed transactions
	InitiateNewTransfer(tm encoding.TransferMessage) (*transaction.Transaction, error)
}
