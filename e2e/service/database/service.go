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

package database

import (
	"encoding/hex"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
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
		connection := persistence.Connect(db)
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
		if !transfersFieldsMatch(*expectedTransferRecord, *actualDbTx) {
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
		if !scheduleFieldsMatch(*expectedRecord, *actualDbTx) {
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
		if !feeFieldsMatch(*actual, *expectedRecord) {
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
