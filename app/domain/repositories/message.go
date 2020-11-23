package repositories

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"

type MessageRepository interface {
	Get(txId, signature string, hash string) ([]message.TransactionMessage, error)
	GetByTransactionId(txId string, txHash string) ([]message.TransactionMessage, error)
	Elect(signature string, txHash string) error
	Add(message *message.TransactionMessage) error
}
