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

func (m MessageRepository) Get(txId, signature string) ([]TransactionMessage, error) {
	var signatures []TransactionMessage
	err := m.dbClient.Where("transaction_id = ? and signature = ?", txId, signature).Find(&signatures).Error
	if err != nil {
		return nil, err
	}
	return signatures, nil
}

func (m MessageRepository) Add(message *TransactionMessage) error {
	return m.dbClient.Create(message).Error
}

func (m MessageRepository) GetByTransactionId(txId string) ([]TransactionMessage, error) {
	var signatures []TransactionMessage
	err := m.dbClient.Where("transaction_id = ?", txId).Find(&signatures).Error
	if err != nil {
		return nil, err
	}
	return signatures, nil
}
