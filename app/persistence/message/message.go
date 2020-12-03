package message

import (
	"gorm.io/gorm"
)

type TransactionMessage struct {
	gorm.Model
	TransactionId        string
	EthAddress           string
	Amount               uint64
	Fee                  string
	Signature            string
	Hash                 string
	SignerAddress        string
	TransactionTimestamp int64
}

type MessageRepository struct {
	dbClient *gorm.DB
}

func NewMessageRepository(dbClient *gorm.DB) *MessageRepository {
	return &MessageRepository{
		dbClient: dbClient,
	}
}

func (m MessageRepository) GetTransaction(txId, signature, hash string) (*TransactionMessage, error) {
	var message TransactionMessage
	err := m.dbClient.Model(&TransactionMessage{}).
		Where("transaction_id = ? and signature = ? and hash = ?", txId, signature, hash).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m MessageRepository) Create(message *TransactionMessage) error {
	return m.dbClient.Create(message).Error
}

func (m MessageRepository) GetTransactions(txId string, hash string) ([]TransactionMessage, error) {
	var messages []TransactionMessage
	err := m.dbClient.Where("transaction_id = ? and hash = ?", txId, hash).Order("transaction_timestamp").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
