package transaction

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"gorm.io/gorm"
)

// Enum Transaction Status
const (
	StatusCompleted         = "COMPLETED"
	StatusSubmitted         = "SUBMITTED"
	StatusInitial           = "INITIAL"
	StatusInsufficientFee   = "INSUFFICIENT_FEE"
	StatusSignatureProvided = "SIGNATURE_PROVIDED"
	StatusSignatureFailed   = "SIGNATURE_FAILED"
	StatusEthTxSubmitted    = "ETH_TX_SUBMITTED"
	StatusEthTxReverted     = "ETH_TX_REVERTED"
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
	EthHash        string
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

func (tr *TransactionRepository) GetIncompleteTransactions() ([]*Transaction, error) {
	var transactions []*Transaction

	err := tr.dbClient.
		Model(Transaction{}).
		Where("status = ? OR status = ? OR status = ?", StatusInitial, StatusSubmitted, StatusEthTxSubmitted).
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
		Status:        StatusInitial,
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
		Updates(Transaction{Status: StatusSubmitted, EthHash: hash}).
		Error
}

func (tr *TransactionRepository) UpdateStatusEthTxReverted(txId string) error {
	return tr.updateStatus(txId, StatusEthTxReverted)
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
