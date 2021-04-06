package database

import (
	"encoding/hex"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type dbVerifier struct {
	transactions repository.Transfer
	messages     repository.Message
}

type Service struct {
	verifiers []dbVerifier
	logger    *log.Entry
}

func NewService(dbConfigs []config.Db) *Service {
	var verifiers []dbVerifier
	for _, db := range dbConfigs {
		connection := persistence.Connect(db)
		newVerifier := dbVerifier{
			transactions: transfer.NewRepository(connection),
			messages:     message.NewRepository(connection),
		}
		verifiers = append(verifiers, newVerifier)
	}
	return &Service{
		verifiers: verifiers,
		logger:    config.GetLoggerFor("DB Validation Service"),
	}
}

func (s *Service) VerifyDatabaseRecords(expectedTransferRecord *entity.Transfer, signatures []string) (bool, error) {
	exists, record, err := s.validTransactionRecord(expectedTransferRecord)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	exists, err = s.validSignatureMessages(record, signatures)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	return true, nil
}

func (s *Service) validTransactionRecord(expectedTransferRecord *entity.Transfer) (bool, *entity.Transfer, error) {
	for _, verifier := range s.verifiers {
		actualDbTx, err := verifier.transactions.GetByTransactionId(expectedTransferRecord.TransactionID)
		if err != nil {
			return false, nil, err
		}
		if !transfersFieldsMatch(*expectedTransferRecord, *actualDbTx) {
			return false, nil, nil
		}
	}
	return true, expectedTransferRecord, nil
}

func (s *Service) validSignatureMessages(record *entity.Transfer, signatures []string) (bool, error) {
	var expectedMessageRecords []entity.Message

	authMsgBytes, err := auth_message.EncodeBytesFrom(record.TransactionID, record.WrappedToken, record.Receiver, record.Amount, record.TxReimbursement, record.GasPrice)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", record.TransactionID, err)
		return false, err
	}
	authMessageStr := hex.EncodeToString(authMsgBytes)

	for _, signature := range signatures {
		signer, signature, err := ethereum.RecoverSignerFromStr(signature, authMsgBytes)
		if err != nil {
			s.logger.Errorf("[%s] - Signature Retrieval failed. Error: [%s]", record.TransactionID, err)
			return false, err
		}

		tm := entity.Message{
			TransferID: record.TransactionID,
			Transfer:   *record,
			Hash:       authMessageStr,
			Signature:  signature,
			Signer:     signer,
		}
		expectedMessageRecords = append(expectedMessageRecords, tm)
	}

	for _, verifier := range s.verifiers {
		messages, err := verifier.messages.Get(record.TransactionID)
		if err != nil {
			return false, err
		}

		for _, m := range expectedMessageRecords {
			if !contains(m, messages) {
				return false, nil
			}
		}
		if len(messages) != len(expectedMessageRecords) {
			return false, nil
		}
	}
	return true, nil
}

func contains(m entity.Message, array []entity.Message) bool {
	for _, a := range array {
		if messagesFieldsMatch(a, m) {
			return true
		}
	}
	return false
}