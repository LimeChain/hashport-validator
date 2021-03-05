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

package repositories

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	joined "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
)

type TransactionRepository interface {
	GetByTransactionId(transactionId string) (*transaction.Transaction, error)
	GetInitialAndSignatureSubmittedTx() ([]*transaction.Transaction, error)
	GetSkippedOrInitialTransactionsAndMessages() (map[joined.CTMKey][]string, error)
	Create(ct *proto.CryptoTransferMessage) error
	Skip(ct *proto.CryptoTransferMessage) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusInsufficientFee(txId string) error
	UpdateStatusSignatureProvided(txId string) error
	UpdateStatusSignatureFailed(txId string) error
	UpdateStatusEthTxSubmitted(txId string, hash string) error
	UpdateStatusEthTxReverted(txId string) error
	UpdateStatusSignatureSubmitted(txId string, submissionTxId string, signature string) error
}
