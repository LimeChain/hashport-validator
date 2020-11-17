package repositories

import "github.com/limechain/hedera-eth-bridge-validator/proto"

type TransactionRepository interface {
	Exists(transactionId string) error
	Create(ct *proto.CryptoTransferMessage) error
	UpdateStatusCancelled(txId string) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusSubmitted(txId string, submissionTxId string) error
}
