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
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Enum Transaction Statuses
const (
	// StatusInitial is the first status on Transaction Record creation
	StatusInitial = "INITIAL"
	// StatusInsufficientFee is status set once transfer is made but the provided TX
	// reimbursement is not enough for validators to process it. This is a terminal status
	StatusInsufficientFee = "INSUFFICIENT_FEE"
	// StatusInProgress is status set once the transfer is accepted and the process
	// of bridging the asset has started
	StatusInProgress = "IN_PROGRESS"
	// StatusCompleted is status set once the Transfer operation is successfully finished.
	// This is a terminal status
	StatusCompleted = "COMPLETED"
	// StatusRecovered TODO after recovery is completed
	StatusRecovered = "RECOVERED"

	// StatusSignatureSubmitted is a SignatureStatus set once the signature is submitted to HCS
	StatusSignatureSubmitted = "SIGNATURE_SUBMITTED"
	// StatusSignatureMined is a SignatureStatus set once the signature submission TX is successfully mined.
	// This is a terminal status
	StatusSignatureMined = "SIGNATURE_MINED"
	// StatusSignatureFailed is a SignatureStatus set if the signature submission TX fails.
	// This is a terminal status
	StatusSignatureFailed = "SIGNATURE_FAILED"

	// StatusEthTxSubmitted is a EthTxStatus set once the Ethereum transaction is submitted to the Ethereum network
	StatusEthTxSubmitted = "ETH_TX_SUBMITTED"
	// StatusEthTxMined is a EthTxStatus set once the Ethereum transaction is successfully mined.
	// This is a terminal status
	StatusEthTxMined = "ETH_TX_MINED"
	// StatusEthTxReverted is a EthTxStatus set if the Ethereum transaction reverts.
	// This is a terminal status
	StatusEthTxReverted = "ETH_TX_REVERTED"

	// StatusEthTxMsgSubmitted is a EthTxMsgStatus set once the `Ethereum TX Hash` is submitted to HCS
	StatusEthTxMsgSubmitted = "ETH_TX_MSG_SUBMITTED"
	// StatusEthTxMsgMined is a EthTxMsgStatus set once the `Ethereum TX Hash` HCS message is mined.
	// This is a terminal status
	StatusEthTxMsgMined = "ETH_TX_MSG_MINED"
	// StatusEthTxMsgFailed is a EthTxMsgStatus set once the `Ethereum TX Hash` HCS message fails
	// This is a terminal status
	StatusEthTxMsgFailed = "ETH_TX_MSG_FAILED"
)

type Transaction struct {
	gorm.Model
	TransactionId         string `gorm:"unique"`
	EthAddress            string
	Amount                string
	Fee                   string
	Signature             string
	SignatureMsgTxId      string
	Status                string
	SignatureMsgStatus    string
	EthTxMsgStatus        string
	EthTxStatus           string
	EthHash               string
	GasPriceGwei          string
	Asset                 string
	ExecuteEthTransaction bool
}

type Repository struct {
	dbClient *gorm.DB
	logger   *log.Entry
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		dbClient: dbClient,
		logger:   config.GetLoggerFor("Transaction Repository"),
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

// TODO Move to message repo
func (tr *Repository) GetSkippedOrInitialTransactionsAndMessages() (map[transaction.CTMKey][]string, error) {
	var messages []*transaction.JoinedTxnMessage

	// TODO
	// Get all Message records which have TXID = ONE OF (Select TXID where Status = RECOVERED || INITIAL)
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
			Asset:         txnMessage.Asset,
		}
		result[t] = append(result[t], txnMessage.Signature)
	}

	return result, nil
}

// Create creates new record of Transaction
func (tr Repository) Create(ct *proto.TransferMessage) (*Transaction, error) {
	tx := &Transaction{
		Model:                 gorm.Model{},
		TransactionId:         ct.TransactionId,
		EthAddress:            ct.EthAddress,
		Amount:                ct.Amount,
		Fee:                   ct.Fee,
		Status:                StatusInitial,
		GasPriceGwei:          ct.GasPriceGwei,
		Asset:                 ct.Asset,
		ExecuteEthTransaction: ct.ExecuteEthTransaction,
	}
	err := tr.dbClient.Create(tx).Error
	return tx, err
}

// Save updates the provided Transaction instance
func (tr Repository) Save(tx *Transaction) error {
	return tr.dbClient.Save(tx).Error
}

func (tr *Repository) SaveRecoveredTxn(ct *proto.TransferMessage) error {
	return tr.dbClient.Create(&Transaction{
		Model:                 gorm.Model{},
		TransactionId:         ct.TransactionId,
		EthAddress:            ct.EthAddress,
		Amount:                ct.Amount,
		Fee:                   ct.Fee,
		Status:                StatusRecovered,
		GasPriceGwei:          ct.GasPriceGwei,
		Asset:                 ct.Asset,
		ExecuteEthTransaction: ct.ExecuteEthTransaction,
	}).Error
}

func (tr Repository) UpdateStatusInsufficientFee(txId string) error {
	return tr.updateStatus(txId, StatusInsufficientFee)
}

func (tr Repository) UpdateStatusSignatureMined(txId string) error {
	return tr.updateSignatureStatus(txId, StatusSignatureMined)
}

func (tr Repository) UpdateStatusSignatureFailed(txId string) error {
	return tr.updateSignatureStatus(txId, StatusSignatureFailed)
}

func (tr Repository) UpdateEthTxSubmitted(txId string, hash string) error {
	err := tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{EthTxStatus: StatusEthTxSubmitted, EthHash: hash}).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Ethereum TX Status of TX [%s] to [%s]", txId, StatusEthTxSubmitted)
	}
	return err
}

func (tr Repository) UpdateEthTxMined(txId string) error {
	err := tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{EthTxStatus: StatusEthTxMined, Status: StatusCompleted}).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Ethereum TX Status of TX [%s] to [%s] and Transaction status to [%s]", txId, StatusEthTxMined, StatusCompleted)
	}
	return err
}

func (tr Repository) UpdateEthTxReverted(txId string) error {
	return tr.updateEthereumTxStatus(txId, StatusEthTxReverted)
}

func (tr Repository) UpdateStatusEthTxMsgSubmitted(txId string) error {
	return tr.updateEthereumTxMsgStatus(txId, StatusEthTxMsgSubmitted)
}

func (tr Repository) UpdateStatusEthTxMsgMined(txId string) error {
	return tr.updateEthereumTxMsgStatus(txId, StatusEthTxMsgMined)
}

func (tr Repository) UpdateStatusEthTxMsgFailed(txId string) error {
	return tr.updateEthereumTxMsgStatus(txId, StatusEthTxMsgFailed)
}

func (tr Repository) updateStatus(txId string, status string) error {
	// Sanity check
	if status != StatusInitial && status != StatusInsufficientFee && status != StatusInProgress && status != StatusCompleted {
		return errors.New("invalid signature status")
	}

	err := tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", status).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Status of TX [%s] to [%s]", txId, status)
	}
	return err
}

func (tr Repository) updateSignatureStatus(txId string, status string) error {
	err := tr.baseUpdateStatus("signature_msg_status", txId, status, []string{StatusSignatureSubmitted, StatusSignatureMined, StatusSignatureFailed})
	if err == nil {
		tr.logger.Debugf("Updated Signature Message Status of TX [%s] to [%s]", txId, status)
	}
	return err
}

func (tr Repository) updateEthereumTxStatus(txId string, status string) error {
	err := tr.baseUpdateStatus("eth_tx_status", txId, status, []string{StatusEthTxSubmitted, StatusEthTxMined, StatusEthTxReverted})
	if err == nil {
		tr.logger.Debugf("Updated Ethereum TX Status of TX [%s] to [%s]", txId, status)
	}
	return err
}

func (tr Repository) updateEthereumTxMsgStatus(txId string, status string) error {
	err := tr.baseUpdateStatus("eth_tx_msg_status", txId, status, []string{StatusEthTxMsgSubmitted, StatusEthTxMsgMined, StatusEthTxMsgFailed})
	if err == nil {
		tr.logger.Debugf("Updated Ethereum TX Message Status of TX [%s] to [%s]", txId, status)
	}
	return err
}

func (tr Repository) baseUpdateStatus(statusColumn, txId, status string, possibleStatuses []string) error {
	if isValidStatus(status, possibleStatuses) {
		return errors.New("invalid status")
	}

	err := tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		UpdateColumn(statusColumn, status).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Signature Message Status of TX [%s] to [%s]", txId, status)
	}
	return err
}

func isValidStatus(status string, possibleStatuses []string) bool {
	for _, option := range possibleStatuses {
		if status == option {
			return true
		}
	}
	return false
}
