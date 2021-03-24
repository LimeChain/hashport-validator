package db_validation

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	transactions repository.Transaction
	messages     repository.Message
	logger       *log.Entry
}

func NewService(dbConfig config.Db) *Service {
	return &Service{
		transactions: transaction.NewRepository(persistence.RunDb(dbConfig)),
		messages:     message.NewRepository(persistence.RunDb(dbConfig)),
		logger:       config.GetLoggerFor("DB Validation Service"),
	}
}

func (s *Service) TransactionRecordExists(expectedTx *transaction.Transaction) (bool, error) {
	actualDbTx, err := s.transactions.GetByTransactionId(expectedTx.TransactionId)
	if err != nil {
		return false, err
	}
	return expectedTx == actualDbTx, nil
}

func (s *Service) SignatureMessagesExist(txId string, expectedMessages []message.TransactionMessage) (bool, error) {
	messages, err := s.messages.GetMessagesFor(txId)
	if err != nil {
		return false, err
	}

	for _, m := range expectedMessages {
		if !contains(m, messages) {
			return false, nil
		}
	}
	return len(messages) == len(expectedMessages), nil
}

func contains(m message.TransactionMessage, array []message.TransactionMessage) bool {
	for _, a := range array {
		if a == m {
			return true
		}
	}
	return false
}
