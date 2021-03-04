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

package transaction

import (
	"errors"
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"gorm.io/gorm"
)

// Enum Transaction Status
const (
	StatusCompleted          = "COMPLETED"
	StatusSignatureSubmitted = "SIGNATURE_SUBMITTED"
	StatusInitial            = "INITIAL"
	StatusInsufficientFee    = "INSUFFICIENT_FEE"
	StatusSignatureProvided  = "SIGNATURE_PROVIDED"
	StatusSignatureFailed    = "SIGNATURE_FAILED"
	StatusEthTxSubmitted     = "ETH_TX_SUBMITTED"
	StatusEthTxReverted      = "ETH_TX_REVERTED"
	StatusSkipped            = "SKIPPED"
)

type Transaction struct {
	gorm.Model
	TransactionId  string `gorm:"unique"`
	EthAddress     string
	Amount         string
	Fee            string
	Signature      string
	SubmissionTxId string
	Status         string
	EthHash        string
	GasPriceGwei   string
}

type TransactionRepository struct {
	dbClient *gorm.DB
}

func NewTransactionRepository(dbClient *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		dbClient: dbClient,
	}
}

func (tr *TransactionRepository) GetByTransactionId(transactionId string) (*Transaction, error) {
	tx := &Transaction{}
	result := tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", transactionId).
		First(tx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return tx, nil
}

func (tr *TransactionRepository) GetInitialAndSignatureSubmittedTx() ([]*Transaction, error) {
	var transactions []*Transaction

	err := tr.dbClient.
		Model(Transaction{}).
		Where("status = ? OR status = ?", StatusInitial, StatusSignatureSubmitted).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

type JoinedTxnMessage struct {
	TransactionId string
	EthAddress    string
	Amount        string
	Fee           string
	Signature     string
	Hash          string
	SignerAddress string
	GasPriceGwei  string
}

func (tr *TransactionRepository) GetSkippedOrInitialTransactionsAndMessages() (map[string][]*message.TransactionMessage, error) {
	var messages []*message.TransactionMessage

	err := tr.dbClient.Preload("transaction_messages").Raw("SELECT " +
		"transactions.transaction_id, " +
		"transaction_messages.eth_address, " +
		"transaction_messages.amount, " +
		"transaction_messages.fee, " +
		"transaction_messages.signature, " +
		"transaction_messages.hash, " +
		"transaction_messages.signer_address " +
		"FROM transactions " +
		"LEFT JOIN transaction_messages ON transactions.transaction_id = transaction_messages.transaction_id " +
		"WHERE transactions.status = 'SKIPPED' OR transactions.status = 'INITIAL' ").
		Scan(&messages).Error
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	result := make(map[string][]*message.TransactionMessage)

	fmt.Println("Skipped Transactions Result")
	fmt.Println(len(messages))

	for _, txnMessage := range messages {
		result[txnMessage.TransactionId] = append(result[txnMessage.TransactionId], txnMessage)
		fmt.Println(txnMessage)
	}

	return result, nil
}

func (tr *TransactionRepository) Create(ct *proto.CryptoTransferMessage) error {
	return tr.dbClient.Create(&Transaction{
		Model:         gorm.Model{},
		TransactionId: ct.TransactionId,
		EthAddress:    ct.EthAddress,
		Amount:        ct.Amount,
		Fee:           ct.Fee,
		Status:        StatusInitial,
		GasPriceGwei:  ct.GasPriceGwei,
	}).Error
}

func (tr *TransactionRepository) Skip(ct *proto.CryptoTransferMessage) error {
	return tr.dbClient.Create(&Transaction{
		Model:         gorm.Model{},
		TransactionId: ct.TransactionId,
		EthAddress:    ct.EthAddress,
		Amount:        ct.Amount,
		Fee:           ct.Fee,
		Status:        StatusSkipped,
		GasPriceGwei:  ct.GasPriceGwei,
	}).Error
}

func (tr *TransactionRepository) UpdateStatusCompleted(txId string) error {
	return tr.updateStatus(txId, StatusCompleted)
}

func (tr *TransactionRepository) UpdateStatusInsufficientFee(txId string) error {
	return tr.updateStatus(txId, StatusInsufficientFee)
}

func (tr *TransactionRepository) UpdateStatusSignatureProvided(txId string) error {
	return tr.updateStatus(txId, StatusSignatureProvided)
}

func (tr *TransactionRepository) UpdateStatusSignatureFailed(txId string) error {
	return tr.updateStatus(txId, StatusSignatureFailed)
}

func (tr *TransactionRepository) UpdateStatusEthTxSubmitted(txId string, hash string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{Status: StatusEthTxSubmitted, EthHash: hash}).
		Error
}

func (tr *TransactionRepository) UpdateStatusEthTxReverted(txId string) error {
	return tr.updateStatus(txId, StatusEthTxReverted)
}

func (tr *TransactionRepository) UpdateStatusSkipped(txId string) error {
	return tr.updateStatus(txId, StatusSkipped)
}

func (tr *TransactionRepository) UpdateStatusSignatureSubmitted(txId string, submissionTxId string, signature string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{Status: StatusSignatureSubmitted, SubmissionTxId: submissionTxId, Signature: signature}).
		Error
}

func (tr *TransactionRepository) updateStatus(txId string, status string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", status).
		Error
}
