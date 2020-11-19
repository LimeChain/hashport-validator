package repositories

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"

type MessageRepository interface {
	Get(txId, signature string) ([]message.TransactionMessage, error)
	Add(message *message.TransactionMessage) error
}
