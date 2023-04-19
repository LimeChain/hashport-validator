/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package verify

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
)

type dbVerifier struct {
	transactions repository.Transfer
	messages     repository.Message
	fee          repository.Fee
	schedule     repository.Schedule
}

type Service struct {
	verifiers               []dbVerifier
	logger                  *log.Entry
	DatabaseRetryCount      int
	DatabaseRetryTimeout    time.Duration
	ExpectedValidatorsCount int
}

func NewService(dbConfigs []config.Database) *Service {
	var verifiers []dbVerifier
	for _, db := range dbConfigs {
		connection := persistence.NewPgConnector(db).Connect()
		newVerifier := dbVerifier{
			transactions: transfer.NewRepository(connection),
			messages:     message.NewRepository(connection),
			fee:          fee.NewRepository(connection),
			schedule:     schedule.NewRepository(connection),
		}
		verifiers = append(verifiers, newVerifier)
	}
	return &Service{
		verifiers: verifiers,
		logger:    config.GetLoggerFor("DB Validation Service"),
	}
}

func (s *Service) VerifyTransferAndSignatureRecords(expectedTransferRecord *entity.Transfer, authMsgBytes []byte, signatures []string) (bool, error) {
	valid, record, err := s.validTransactionRecord(expectedTransferRecord)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, fmt.Errorf("transaction record is not valid")
	}

	valid, err = s.validSignatureMessages(record, authMsgBytes, signatures)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, fmt.Errorf("signature message is not valid")
	}
	return true, nil
}

func (s *Service) VerifyTransferRecord(expectedTransferRecord *entity.Transfer) (bool, error) {
	valid, _, err := s.validTransactionRecord(expectedTransferRecord)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, fmt.Errorf("transaction record is not valid")
	}

	return true, nil
}

func (s *Service) VerifyScheduleRecord(expectedRecord *entity.Schedule) (bool, error) {
	valid, err := s.validScheduleRecord(expectedRecord)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, fmt.Errorf("schedule record is not valid")
	}

	return true, nil
}

func (s *Service) validTransactionRecord(expectedRecord *entity.Transfer) (bool, *entity.Transfer, error) {
	for _, verifier := range s.verifiers {
		actualDbTx, err := s.getTransactionById(verifier, expectedRecord)
		if err != nil {
			return false, nil, err
		}

		if actualDbTx == nil {
			return false, nil, fmt.Errorf("database transaction record [%s] not found", expectedRecord.TransactionID)
		}

		if !s.transfersFieldsMatch(*expectedRecord, *actualDbTx) {
			s.logger.Infof("expected transaction: [%+v]; actual: [%+v]", expectedRecord, actualDbTx)
			return false, nil, fmt.Errorf("database transaction record [%s] not expected", expectedRecord.TransactionID)
		}
	}
	return true, expectedRecord, nil
}

func (s *Service) validScheduleRecord(expectedRecord *entity.Schedule) (bool, error) {
	for _, verifier := range s.verifiers {
		actualDbTx, err := s.getScheduleByTransactionId(verifier, expectedRecord)
		if err != nil {
			return false, err
		}

		if actualDbTx == nil {
			return false, fmt.Errorf("database schedule record [%s] not found", expectedRecord.TransactionID)
		}

		if !s.scheduleIsAsExpected(expectedRecord, actualDbTx) {
			s.logger.Infof("expected schedule: [%+v]; actual: [%+v]", expectedRecord, actualDbTx)
			return false, fmt.Errorf("database schedule record [%s] not expected", expectedRecord.TransactionID)
		}
	}
	return true, nil
}

func (s *Service) validSignatureMessages(record *entity.Transfer, authMsgBytes []byte, signatures []string) (bool, error) {
	var expectedMessageRecords []entity.Message

	authMessageStr := hex.EncodeToString(authMsgBytes)

	for _, signature := range signatures {
		signer, signature, err := evm.RecoverSignerFromStr(signature, authMsgBytes)
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
		actualMessages, err := s.getMessageListByTransactionId(record, verifier) // TODO : Check if retry is needed
		if err != nil {
			return false, err
		}

		if actualMessages == nil {
			return false, fmt.Errorf("database message records [%s] not found", record.TransactionID)

		}

		messagesLength := len(actualMessages)
		expectedMessagesLength := len(expectedMessageRecords)
		if messagesLength != expectedMessagesLength {
			return false, fmt.Errorf("expected database message records [%s] length go be [%d], but was [%d]", record.TransactionID, expectedMessagesLength, messagesLength)
		}

		for _, ele := range expectedMessageRecords {
			if !s.messageListContains(ele, actualMessages) {
				s.logger.Infof("expected message: [%+v]; actual database message records: [%+v]", ele, actualMessages)
				return false, fmt.Errorf("unexpected message")
			}
		}

	}
	return true, nil
}

func (s *Service) VerifyFeeRecord(expectedRecord *entity.Fee) (bool, error) {
	for _, verifier := range s.verifiers {
		actualDbTx, err := s.getFeeByTransactionId(verifier, expectedRecord)
		if err != nil {
			return false, err
		}

		if actualDbTx == nil {
			return false, fmt.Errorf("database fee record [%s] not found", expectedRecord.TransactionID)
		}

		if !s.feeIsAsExpected(expectedRecord, actualDbTx) {
			s.logger.Infof("expected fee record: [%+v]; actual: [%+v]", expectedRecord, actualDbTx)
			return false, fmt.Errorf("database fee record [%s] not expected", expectedRecord.TransactionID)
		}
	}

	return true, nil
}

func (s *Service) getMessageListByTransactionId(expectedTransferRecord *entity.Transfer, verifier dbVerifier) ([]entity.Message, error) {
	var result []entity.Message
	var err error

	currentCount := 0

	for currentCount < s.DatabaseRetryCount {
		currentCount++

		result, err = verifier.messages.Get(expectedTransferRecord.TransactionID)

		// if result == nil && err == nil - the record is not found in the database - retry
		if (result != nil && len(result) == s.ExpectedValidatorsCount) || err != nil {
			return result, err
		}

		time.Sleep(s.DatabaseRetryTimeout * time.Second)
		s.logger.Infof("Database Message records [%s] retry %d", expectedTransferRecord.TransactionID, currentCount)
	}

	s.logger.Errorf("Database Message records [%s] not found after %d retries", expectedTransferRecord.TransactionID, currentCount)
	return result, err
}

// getTransactionById returns a record from the validator database. The method will retry few times to get completed record.
func (s *Service) getTransactionById(verifier dbVerifier, expectedTransferRecord *entity.Transfer) (*entity.Transfer, error) {
	var result *entity.Transfer
	var err error

	currentCount := 0

	for currentCount < s.DatabaseRetryCount {
		currentCount++

		result, err = verifier.transactions.GetByTransactionId(expectedTransferRecord.TransactionID)

		// if result == nil && err == nil - the record is not found in the database - retry
		// if status != COMPLETED - the record processing is not finished - retry
		if (result != nil && result.Status == "COMPLETED") || err != nil {
			return result, err
		}

		time.Sleep(s.DatabaseRetryTimeout * time.Second)
		s.logger.Infof("Database Transaction record [%s] retry %d", expectedTransferRecord.TransactionID, currentCount)
	}

	s.logger.Errorf("Database Transaction record [%s] not found after %d retries", expectedTransferRecord.TransactionID, currentCount)
	return result, err
}

// getScheduleByTransactionId returns a record from the validator database. The method will retry few times to get completed record.
func (s *Service) getScheduleByTransactionId(verifier dbVerifier, expectedRecord *entity.Schedule) (*entity.Schedule, error) {
	var result *entity.Schedule
	var err error

	currentCount := 0

	for currentCount < s.DatabaseRetryCount {
		currentCount++

		result, err = verifier.schedule.Get(expectedRecord.TransactionID)

		// if result == nil && err == nil - the record is not found in the database - retry
		// if status != COMPLETED - the record processing is not finished - retry
		if (result != nil && result.Status == "COMPLETED") || err != nil {
			return result, err
		}

		time.Sleep(s.DatabaseRetryTimeout * time.Second)
		s.logger.Infof("Database Schedule record [%s] retry %d", expectedRecord.TransactionID, currentCount)
	}

	s.logger.Errorf("Database Schedule record [%s] not found after %d retries", expectedRecord.TransactionID, currentCount)
	return result, err
}

// getFeeByTransactionId returns a record from the validator database. The method will retry few times to get completed record.
func (s *Service) getFeeByTransactionId(verifier dbVerifier, expectedRecord *entity.Fee) (*entity.Fee, error) {
	var result *entity.Fee
	var err error

	currentCount := 0

	for currentCount < s.DatabaseRetryCount {
		currentCount++

		result, err = verifier.fee.Get(expectedRecord.TransactionID)

		// if result == nil && err == nil - the record is not found in the database - retry
		// if status != COMPLETED - the record processing is not finished - retry
		if (result != nil && result.Status == "COMPLETED") || err != nil {
			return result, err
		}

		time.Sleep(s.DatabaseRetryTimeout * time.Second)
		s.logger.Infof("Database Fee record [%s] retry %d", expectedRecord.TransactionID, currentCount)
	}

	s.logger.Errorf("Database Fee record [%s] not found after %d retries", expectedRecord.TransactionID, currentCount)
	return result, err
}

func FeeRecord(t *testing.T, dbValidation *Service, expectedRecord *entity.Fee) {
	t.Helper()
	ok, err := dbValidation.VerifyFeeRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !ok {
		t.Fatalf("[%s] - Database does not contain expected fee records", expectedRecord.TransactionID)
	}
}

func TransferRecord(t *testing.T, dbValidation *Service, expectedRecord *entity.Transfer) {
	t.Helper()
	exist, err := dbValidation.VerifyTransferRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}

	if !exist {
		t.Fatalf("[%s] - Database does not contain expected transfer records", expectedRecord.TransactionID)
	}
}

func ScheduleRecord(t *testing.T, dbValidation *Service, expectedRecord *entity.Schedule) {
	t.Helper()
	exist, err := dbValidation.VerifyScheduleRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}

	if !exist {
		t.Fatalf("[%s] - Database does not contain expected schedule records", expectedRecord.TransactionID)
	}
}

func TransferRecordAndSignatures(t *testing.T, dbValidation *Service, expectedRecord *entity.Transfer, authMsgBytes []byte, signatures []string) {
	t.Helper()
	exist, err := dbValidation.VerifyTransferAndSignatureRecords(expectedRecord, authMsgBytes, signatures)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}

	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func (s *Service) messageListContains(message entity.Message, actualMessages []entity.Message) bool {
	for _, a := range actualMessages {
		if s.messagesFieldsMatch(a, message) {
			return true
		}
	}

	return false
}

func (s *Service) messagesFieldsMatch(comparing, comparable entity.Message) bool {
	return comparable.TransferID == comparing.TransferID &&
		comparable.Signature == comparing.Signature &&
		comparable.Hash == comparing.Hash &&
		s.transfersFieldsMatch(comparable.Transfer, comparing.Transfer) &&
		comparable.Signer == comparing.Signer
}

func (s *Service) transfersFieldsMatch(comparing, comparable entity.Transfer) bool {
	return reflect.DeepEqual(comparable, comparing)
}

func (s *Service) scheduleIsAsExpected(expected, actual *entity.Schedule) bool {
	return reflect.DeepEqual(expected, actual)
}

func (s *Service) feeIsAsExpected(expected, actual *entity.Fee) bool {
	return reflect.DeepEqual(expected, actual)
}
