package message

import (
	"gorm.io/gorm"
	"strconv"
)

type TransactionMessage struct {
	gorm.Model
	TransactionId        string
	EthAddress           string
	Amount               uint64
	Fee                  string
	Signature            string
	Hash                 string
	Leader               bool
	SignerAddress        string
	TransactionTimestamp string
}

type ByTimestamp []TransactionMessage

func (tm ByTimestamp) Len() int {
	return len(tm)
}
func (tm ByTimestamp) Swap(i, j int) {
	tm[i], tm[j] = tm[j], tm[i]
}
func (tm ByTimestamp) Less(i, j int) bool {
	firstTimestamp, err := strconv.ParseFloat(tm[i].TransactionTimestamp, 32)
	if err != nil {

	}
	secondTimestamp, err := strconv.ParseFloat(tm[j].TransactionTimestamp, 32)
	if err != nil {

	}
	return firstTimestamp < secondTimestamp
}

type MessageRepository struct {
	dbClient *gorm.DB
}

func NewMessageRepository(dbClient *gorm.DB) *MessageRepository {
	return &MessageRepository{
		dbClient: dbClient,
	}
}

func (m MessageRepository) GetTransaction(txId, signature, hash string) ([]TransactionMessage, error) {
	var messages []TransactionMessage
	err := m.dbClient.Where("transaction_id = ? and signature = ? and hash = ?", txId, signature, hash).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (m MessageRepository) Create(message *TransactionMessage) error {
	return m.dbClient.Create(message).Error
}

func (m MessageRepository) GetByTransactionId(txId string, hash string) ([]TransactionMessage, error) {
	var signatures []TransactionMessage
	err := m.dbClient.Where("transaction_id = ? and hash = ?", txId, hash).Find(&signatures).Error
	if err != nil {
		return nil, err
	}
	return signatures, nil
}

func (m MessageRepository) Elect(signature string, hash string) error {
	return m.dbClient.Model(&TransactionMessage{}).
		Where("signature = ? and hash = ?", signature, hash).
		Update("leader", "true").Error
}
