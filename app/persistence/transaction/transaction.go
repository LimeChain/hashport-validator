package transaction

import (
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
	Fee            uint64
	SubmissionTxId string
	Status         string
}

type TransactionRepository struct {
	dbClient *gorm.DB
}

func (tr *TransactionRepository) Exists(transactionId string) (bool, error) {
	err := tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", transactionId).
		First(&Transaction{}).Error
	if err != nil {
		return false, err
	}
	return true, nil
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

func (tr *TransactionRepository) UpdateStatusSubmitted(txId string, submissionTxId string) error {
	return tr.dbClient.
		Model(Transaction{}).
		Where("transaction_id = ?", txId).
		Updates(Transaction{Status: StatusSubmitted, SubmissionTxId: submissionTxId}).
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
