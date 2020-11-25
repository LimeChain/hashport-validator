package transaction

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"gorm.io/gorm"
)

// Enum Transaction Status
const (
	StatusCancelled = "CANCELLED"
	StatusCompleted = "COMPLETED"
	StatusPending   = "PENDING"
	StatusSubmitted = "SUBMITTED"
)

type Transaction struct {
	gorm.Model
	TransactionId  string `gorm:"unique"`
	EthAddress     string
	Amount         uint64
	Fee            string
	Signature      string
	SubmissionTxId string
	Status         string
}

type TransactionRepository struct {
	dbClient *gorm.DB
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

func (tr *TransactionRepository) GetPendingOrSubmittedTransactions() ([]*Transaction, error) {
	var transactions []*Transaction

	err := tr.dbClient.
		Model(Transaction{}).
		Where("status = ? OR status = ?", StatusPending, StatusSubmitted).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (tr *TransactionRepository) Create(ct *proto.CryptoTransferMessage) error {
	return tr.dbClient.Create(&Transaction{
		Model:         gorm.Model{},
		TransactionId: ct.TransactionId,
		EthAddress:    ct.EthAddress,
		Amount:        ct.Amount,
		Fee:           ct.Fee,
		Status:        StatusPending,
	}).Error
}

func (tr *TransactionRepository) UpdateStatusCancelled(txId string) error {
	return tr.updateStatus(txId, StatusCancelled)
}

func (tr *TransactionRepository) UpdateStatusCompleted(txId string) error {
	return tr.updateStatus(txId, StatusCompleted)
}

func (tr *TransactionRepository) UpdateStatusSubmitted(txId string, submissionTxId string, signature string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{Status: StatusSubmitted, SubmissionTxId: submissionTxId, Signature: signature}).
		Error
}

func (tr *TransactionRepository) updateStatus(txId string, status string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", status).
		Error
}

func NewTransactionRepository(dbClient *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		dbClient: dbClient,
	}
}
