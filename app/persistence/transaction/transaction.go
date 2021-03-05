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
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
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

type Repository struct {
	dbClient *gorm.DB
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		dbClient: dbClient,
	}
}

func (tr Repository) GetByTransactionId(transactionId string) (*Transaction, error) {
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

func (tr Repository) GetInitialAndSignatureSubmittedTx() ([]*Transaction, error) {
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

func (tr *Repository) GetSkippedOrInitialTransactionsAndMessages() (map[transaction.CTMKey][]string, error) {
	var messages []*transaction.JoinedTxnMessage

	err := tr.dbClient.Preload("transaction_messages").Raw("SELECT " +
		"transactions.transaction_id, " +
		"transactions.eth_address, " +
		"transactions.amount, " +
		"transactions.fee, " +
		"transaction_messages.signature, " +
		"transactions.gas_price_gwei " +
		"FROM transactions " +
		"LEFT JOIN transaction_messages ON transactions.transaction_id = transaction_messages.transaction_id " +
		"WHERE transactions.status = 'SKIPPED' OR transactions.status = 'INITIAL' ").
		Scan(&messages).Error
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	result := make(map[transaction.CTMKey][]string)

	for _, txnMessage := range messages {
		t := transaction.CTMKey{
			TransactionId: txnMessage.TransactionId,
			EthAddress:    txnMessage.EthAddress,
			Amount:        txnMessage.Amount,
			Fee:           txnMessage.Fee,
			GasPriceGwei:  txnMessage.GasPriceGwei,
		}
		result[t] = append(result[t], txnMessage.Signature)
	}

	return result, nil
}

func (tr Repository) Create(ct *proto.CryptoTransferMessage) error {
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

func (tr *Repository) Skip(ct *proto.CryptoTransferMessage) error {
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

func (tr Repository) UpdateStatusCompleted(txId string) error {
	return tr.updateStatus(txId, StatusCompleted)
}

func (tr Repository) UpdateStatusInsufficientFee(txId string) error {
	return tr.updateStatus(txId, StatusInsufficientFee)
}

func (tr Repository) UpdateStatusSignatureProvided(txId string) error {
	return tr.updateStatus(txId, StatusSignatureProvided)
}

func (tr Repository) UpdateStatusSignatureFailed(txId string) error {
	return tr.updateStatus(txId, StatusSignatureFailed)
}

func (tr Repository) UpdateStatusEthTxSubmitted(txId string, hash string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{Status: StatusEthTxSubmitted, EthHash: hash}).
		Error
}

func (tr Repository) UpdateStatusEthTxReverted(txId string) error {
	return tr.updateStatus(txId, StatusEthTxReverted)
}

func (tr *Repository) UpdateStatusSkipped(txId string) error {
	return tr.updateStatus(txId, StatusSkipped)
}

func (tr Repository) UpdateStatusSignatureSubmitted(txId string, submissionTxId string, signature string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{Status: StatusSignatureSubmitted, SubmissionTxId: submissionTxId, Signature: signature}).
		Error
}

func (tr Repository) updateStatus(txId string, status string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", status).
		Error
}
