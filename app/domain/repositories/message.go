package repositories

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"

type MessageRepository interface {
	GetTransaction(txId, signature string) ([]message.TransactionMessage, error)
	Create(message *message.TransactionMessage) error
}
