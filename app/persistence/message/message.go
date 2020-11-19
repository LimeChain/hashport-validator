package message

import (
	hcstopicmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/hcs-topic-message"
	"gorm.io/gorm"
)

type TransactionMessage struct {
	TransactionId string
	EthAddress    string
	Amount        uint64
	Fee           string
	Signature     string
}

type MessageRepository struct {
	dbClient *gorm.DB
}

func NewMessageRepository(dbClient *gorm.DB) *MessageRepository {
	return &MessageRepository{
		dbClient: dbClient,
	}
}

func (m MessageRepository) Get(TxId string) ([]TransactionMessage, error) {
	var signatures []TransactionMessage
	err := m.dbClient.Where("transaction_id = ?", TxId).Find(&signatures).Error
	if err != nil {
		return nil, err
	}
	return signatures, nil
}

func (m MessageRepository) Add(message *hcstopicmessage.ConsensusMessage) error {
	return m.dbClient.Create(&TransactionMessage{
		TransactionId: message.TransactionID,
		Signature:     message.Signature,
		EthAddress:    message.EthAddress,
		Fee:           message.Fee,
		Amount:        message.Amount,
	}).Error
}
