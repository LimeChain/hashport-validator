package message

import (
	"gorm.io/gorm"
)

type TransactionMessage struct {
	TransactionId string
	EthAddress    string
	Amount        uint64
	Fee           string
	Signature     string
	Hash          string
}

type MessageRepository struct {
	dbClient *gorm.DB
}

func NewMessageRepository(dbClient *gorm.DB) *MessageRepository {
	return &MessageRepository{
		dbClient: dbClient,
	}
}

func (m MessageRepository) GetTransaction(txId, signature string) ([]TransactionMessage, error) {
	var messages []TransactionMessage
	err := m.dbClient.Where("transaction_id = ? and signature = ?", txId, signature).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (m MessageRepository) Create(message *TransactionMessage) error {
	return m.dbClient.Create(message).Error
}
