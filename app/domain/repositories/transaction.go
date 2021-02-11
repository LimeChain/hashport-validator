package repositories

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
)

type TransactionRepository interface {
	GetByTransactionId(transactionId string) (*transaction.Transaction, error)
	GetInitialAndSignatureSubmittedTx() ([]*transaction.Transaction, error)
	Create(ct *proto.CryptoTransferMessage) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusInsufficientFee(txId string) error
	UpdateStatusSignatureProvided(txId string) error
	UpdateStatusSignatureFailed(txId string) error
	UpdateStatusEthTxSubmitted(txId string, hash string) error
	UpdateStatusEthTxReverted(txId string) error
	UpdateStatusSignatureSubmitted(txId string, submissionTxId string, signature string) error
}
