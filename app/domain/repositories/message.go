package repositories

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"

type MessageRepository interface {
	GetTransaction(txId, signature, hash string) ([]message.TransactionMessage, error)
	Create(message *message.TransactionMessage) error
	GetByTransactionId(txId string, txHash string) ([]message.TransactionMessage, error)
}
