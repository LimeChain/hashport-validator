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
	"testing"

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
	verifiers []dbVerifier
	logger    *log.Entry
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
		return false, nil
	}

	valid, err = s.validSignatureMessages(record, authMsgBytes, signatures)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	return true, nil
}

func (s *Service) VerifyTransferRecord(expectedTransferRecord *entity.Transfer) (bool, error) {
	valid, _, err := s.validTransactionRecord(expectedTransferRecord)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}

	return true, nil
}

func (s *Service) VerifyScheduleRecord(expectedRecord *entity.Schedule) (bool, error) {
	valid, err := s.validScheduleRecord(expectedRecord)
	if err != nil {
		return false, err
	}
	if !valid {
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
		if !transferIsAsExpected(expectedTransferRecord, actualDbTx) {
			return false, nil, nil
		}
	}
	return true, expectedTransferRecord, nil
}

func (s *Service) validScheduleRecord(expectedRecord *entity.Schedule) (bool, error) {
	for _, verifier := range s.verifiers {
		actualDbTx, err := verifier.schedule.Get(expectedRecord.TransactionID)
		if err != nil {
			return false, err
		}
		if !scheduleIsAsExpected(expectedRecord, actualDbTx) {
			return false, nil
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

func (s *Service) VerifyFeeRecord(expectedRecord *entity.Fee) (bool, error) {
	for _, verifier := range s.verifiers {
		actual, err := verifier.fee.Get(expectedRecord.TransactionID)
		if err != nil {
			return false, err
		}
		if !feeIsAsExpected(actual, expectedRecord) {
			return false, nil
		}
	}
	return true, nil
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

func contains(m entity.Message, array []entity.Message) bool {
	for _, a := range array {
		if messagesFieldsMatch(a, m) {
			return true
		}
	}
	return false
}

func messagesFieldsMatch(comparing, comparable entity.Message) bool {
	return comparable.TransferID == comparing.TransferID &&
		comparable.Signature == comparing.Signature &&
		comparable.Hash == comparing.Hash &&
		transfersFieldsMatch(comparable.Transfer, comparing.Transfer) &&
		comparable.Signer == comparing.Signer
}

func transfersFieldsMatch(comparing, comparable entity.Transfer) bool {
	return comparable.TransactionID == comparing.TransactionID &&
		comparable.SourceChainID == comparing.SourceChainID &&
		comparable.TargetChainID == comparing.TargetChainID &&
		comparable.NativeChainID == comparing.NativeChainID &&
		comparable.SourceAsset == comparing.SourceAsset &&
		comparable.TargetAsset == comparing.TargetAsset &&
		comparable.NativeAsset == comparing.NativeAsset &&
		comparable.Receiver == comparing.Receiver &&
		comparable.Amount == comparing.Amount &&
		comparable.Status == comparing.Status &&
		comparable.Metadata == comparing.Metadata &&
		comparable.SerialNumber == comparing.SerialNumber &&
		comparable.IsNft == comparing.IsNft &&
		comparable.Originator == comparing.Originator &&
		comparable.Timestamp == comparing.Timestamp
}

func transferIsAsExpected(expected, actual *entity.Transfer) bool {
	if expected.TransactionID != actual.TransactionID ||
		expected.SourceChainID != actual.SourceChainID ||
		expected.TargetChainID != actual.TargetChainID ||
		expected.NativeChainID != actual.NativeChainID ||
		expected.SourceAsset != actual.SourceAsset ||
		expected.TargetAsset != actual.TargetAsset ||
		expected.NativeAsset != actual.NativeAsset ||
		expected.Receiver != actual.Receiver ||
		expected.Amount != actual.Amount ||
		expected.Status != actual.Status ||
		expected.SerialNumber != actual.SerialNumber ||
		expected.Metadata != actual.Metadata ||
		expected.IsNft != actual.IsNft ||
		expected.Timestamp != actual.Timestamp {
		return false
	}
	return true
}

func scheduleIsAsExpected(expected, actual *entity.Schedule) bool {
	if expected.TransactionID != actual.TransactionID ||
		expected.ScheduleID != actual.ScheduleID ||
		expected.HasReceiver != actual.HasReceiver ||
		expected.Operation != actual.Operation ||
		expected.Status != actual.Status ||
		expected.TransferID.String != actual.TransferID.String {
		return false
	}
	return true
}

func feeIsAsExpected(expected, actual *entity.Fee) bool {
	if expected.TransactionID != actual.TransactionID ||
		expected.ScheduleID != actual.ScheduleID ||
		expected.Amount != actual.Amount ||
		expected.Status != actual.Status ||
		expected.TransferID.String != actual.TransferID.String {
		return false
	}
	return true
}
