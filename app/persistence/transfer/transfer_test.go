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

package transfer

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	repository    *Repository
	dbConn        *gorm.DB
	sqlMock       sqlmock.Sqlmock
	transactionId = "transactionId"
	sourceChainId = uint64(0)
	targetChainId = uint64(1)
	nativeChainId = uint64(2)
	sourceAsset   = "sourceAsset"
	targetAsset   = "targetAsset"
	nativeAsset   = "nativeAsset"
	receiver      = "receiver"
	amount        = "amount"
	fee           = ""
	someStatus    = status.Initial
	serialNumber  = int64(0)
	metadata      = "metadata"
	isNft         = false
	now           = time.Now().UTC()
	nanoTime      = entity.NanoTime{Time: now}
	originator    = "originator"

	transferColumns = []string{"transaction_id", "source_chain_id", "target_chain_id", "native_chain_id", "source_asset", "target_asset", "native_asset", "receiver", "amount", "fee", "status", "serial_number", "metadata", "is_nft", "timestamp", "originator"}
	feeColumns      = []string{"transaction_id", "schedule_id", "amount", "status", "transfer_id"}
	messageColumns  = []string{"transfer_id", "hash", "signature", "signer", "transaction_timestamp"}

	transferRowArgs = []driver.Value{transactionId, sourceChainId, targetChainId, nativeChainId, sourceAsset, targetAsset, nativeAsset, receiver, amount, fee, someStatus, serialNumber, metadata, isNft, nanoTime, originator}
	feesRowArgs     = []driver.Value{
		transactionId,
		expectedEntityFee.ScheduleID,
		expectedEntityFee.Amount,
		expectedEntityFee.Status,
		expectedEntityFee.TransferID,
	}
	messageRowArgs = []driver.Value{transactionId, "hash", "signature", "signer", uint8(1)}

	expectedEntityTransfer = &entity.Transfer{
		TransactionID: transactionId,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		NativeChainID: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Amount:        amount,
		Fee:           fee,
		Status:        someStatus,
		SerialNumber:  serialNumber,
		Metadata:      metadata,
		IsNft:         isNft,
		Timestamp:     nanoTime,
		Originator:    originator,
	}
	expectedModelTransfer = &model.Transfer{
		TransactionId:    transactionId,
		SourceChainId:    sourceChainId,
		TargetChainId:    targetChainId,
		NativeChainId:    nativeChainId,
		SourceAsset:      sourceAsset,
		TargetAsset:      targetAsset,
		NativeAsset:      nativeAsset,
		Receiver:         receiver,
		Amount:           amount,
		SerialNum:        serialNumber,
		Metadata:         metadata,
		IsNft:            isNft,
		NetworkTimestamp: time.Now().String(),
		Timestamp:        now,
		Originator:       originator,
	}

	expectedEntityFee = entity.Fee{
		TransactionID: transactionId,
		ScheduleID:    "scheduleId",
		Amount:        "amount",
		Status:        "status",
		TransferID: sql.NullString{
			String: transactionId,
			Valid:  true,
		},
	}
	expectedEntityMessage = entity.Message{
		TransferID:           transactionId,
		Hash:                 "hash",
		Signature:            "signature",
		Signer:               "signer",
		TransactionTimestamp: 1,
	}

	expectedEntityTransferWithFee = &entity.Transfer{
		TransactionID: transactionId,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		NativeChainID: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Amount:        amount,
		Fee:           fee,
		Status:        someStatus,
		SerialNumber:  serialNumber,
		Metadata:      metadata,
		IsNft:         isNft,
		Timestamp:     nanoTime,
		Originator:    originator,
		Fees: []entity.Fee{
			expectedEntityFee,
		},
	}
	expectedEntityTransferWithPreloads = &entity.Transfer{
		TransactionID: transactionId,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		NativeChainID: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Amount:        amount,
		Fee:           fee,
		Status:        someStatus,
		SerialNumber:  serialNumber,
		Metadata:      metadata,
		IsNft:         isNft,
		Timestamp:     nanoTime,
		Originator:    originator,
		Fees: []entity.Fee{
			expectedEntityFee,
		},
		Messages: []entity.Message{
			expectedEntityMessage,
		},
	}

	getByTransactionIdQuery       = regexp.QuoteMeta(`SELECT * FROM "transfers" WHERE transaction_id = $1`)
	getWithPreloadsTransfersQuery = regexp.QuoteMeta(`SELECT * FROM "transfers" WHERE transaction_id = $1`)
	getWithPreloadsFeesQuery      = regexp.QuoteMeta(`SELECT * FROM "fees" WHERE "fees"."transfer_id" = $1`)
	getWithPreloadsMessagesQuery  = regexp.QuoteMeta(`SELECT * FROM "messages" WHERE "messages"."transfer_id" = $1`)

	createQuery       = regexp.QuoteMeta(`INSERT INTO "transfers" ("transaction_id","source_chain_id","target_chain_id","native_chain_id","source_asset","target_asset","native_asset","receiver","amount","fee","status","serial_number","metadata","is_nft","timestamp","originator") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`)
	saveQuery         = regexp.QuoteMeta(`UPDATE "transfers" SET "source_chain_id"=$1,"target_chain_id"=$2,"native_chain_id"=$3,"source_asset"=$4,"target_asset"=$5,"native_asset"=$6,"receiver"=$7,"amount"=$8,"fee"=$9,"status"=$10,"serial_number"=$11,"metadata"=$12,"is_nft"=$13,"timestamp"=$14,"originator"=$15 WHERE "transaction_id" = $16`)
	updateFeeQuery    = regexp.QuoteMeta(`UPDATE "transfers" SET "fee"=$1 WHERE transaction_id = $2`)
	updateStatusQuery = regexp.QuoteMeta(`UPDATE "transfers" SET "status"=$1 WHERE transaction_id = $2`)
)

func setup() {
	mocks.Setup()
	dbConn, sqlMock, _ = helper.SetupSqlMock()
	repository = &Repository{
		dbClient: dbConn,
		logger:   config.GetLoggerFor("Transfer Repository"),
	}
}

func Test_NewRepository(t *testing.T) {
	setup()

	actual := NewRepository(dbConn)
	assert.Equal(t, repository, actual)
}

func Test_GetByTransactionId(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, transferColumns, transferRowArgs, getByTransactionIdQuery, transactionId)

	actual, err := repository.GetByTransactionId(transactionId)
	assert.Nil(t, err)
	assert.Equal(t, expectedEntityTransfer, actual)
}

func Test_GetByTransactionId_NotFound(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, getByTransactionIdQuery, transactionId)

	actual, err := repository.GetByTransactionId(transactionId)
	assert.Nil(t, err)
	assert.Nil(t, actual)
}

func Test_GetByTransactionId_InvalidData(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrInvalidData(sqlMock, getByTransactionIdQuery, transactionId)

	actual, err := repository.GetByTransactionId(transactionId)
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func Test_GetWithPreloads(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, transferColumns, transferRowArgs, getWithPreloadsTransfersQuery, transactionId)
	helper.SqlMockPrepareQuery(sqlMock, feeColumns, feesRowArgs, getWithPreloadsFeesQuery, transactionId)
	helper.SqlMockPrepareQuery(sqlMock, messageColumns, messageRowArgs, getWithPreloadsMessagesQuery, transactionId)

	actual, err := repository.GetWithPreloads(transactionId)
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedEntityTransferWithPreloads, actual)
}

func Test_GetWithPreloads_NotFound(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, getWithPreloadsTransfersQuery, transactionId)

	actual, err := repository.GetWithPreloads(transactionId)
	assert.Nil(t, err)
	assert.Nil(t, actual)
}

func Test_GetWithPreloads_InvalidData(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrInvalidData(sqlMock, getWithPreloadsTransfersQuery, transactionId)

	actual, err := repository.GetWithPreloads(transactionId)
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func Test_GetWithFee(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, transferColumns, transferRowArgs, getWithPreloadsTransfersQuery, transactionId)
	helper.SqlMockPrepareQuery(sqlMock, feeColumns, feesRowArgs, getWithPreloadsFeesQuery, transactionId)

	actual, err := repository.GetWithFee(transactionId)
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedEntityTransferWithFee, actual)
}

func Test_GetWithFee_NotFound(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, getWithPreloadsTransfersQuery, transactionId)

	actual, err := repository.GetWithFee(transactionId)
	assert.Nil(t, err)
	assert.Nil(t, actual)
}

func Test_GetWithFee_InvalidData(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrInvalidData(sqlMock, getWithPreloadsTransfersQuery, transactionId)

	actual, err := repository.GetWithFee(transactionId)
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func Test_Create(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, createQuery,
		transactionId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		sourceAsset,
		targetAsset,
		nativeAsset,
		receiver,
		amount,
		"", //fee
		someStatus,
		serialNumber,
		metadata,
		isNft,
		nanoTime,
		originator)

	actual, err := repository.Create(expectedModelTransfer)
	assert.Nil(t, err)
	assert.Equal(t, expectedEntityTransfer, actual)
}

func Test_Create_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, createQuery,
		transactionId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		sourceAsset,
		targetAsset,
		nativeAsset,
		receiver,
		amount,
		"", //fee
		someStatus,
		serialNumber,
		metadata,
		isNft,
		nanoTime,
		originator)

	actual, err := repository.Create(expectedModelTransfer)
	assert.NotNil(t, err)
	assert.NotNil(t, actual)
}

func Test_Save(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, saveQuery,
		sourceChainId,
		targetChainId,
		nativeChainId,
		sourceAsset,
		targetAsset,
		nativeAsset,
		receiver,
		amount,
		fee,
		someStatus,
		serialNumber,
		metadata,
		isNft,
		nanoTime,
		originator,
		transactionId)

	err := repository.Save(expectedEntityTransfer)
	assert.Nil(t, err)
}

func Test_Save_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, saveQuery,
		sourceChainId,
		targetChainId,
		nativeChainId,
		sourceAsset,
		targetAsset,
		nativeAsset,
		receiver,
		amount,
		fee,
		someStatus,
		serialNumber,
		metadata,
		isNft,
		nanoTime,
		originator,
		transactionId)

	err := repository.Save(expectedEntityTransfer)
	assert.NotNil(t, err)
}

func Test_UpdateFee(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, updateFeeQuery,
		fee, transactionId)

	err := repository.UpdateFee(transactionId, fee)
	assert.Nil(t, err)
}

func Test_UpdateFee_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, updateFeeQuery,
		fee, transactionId)

	err := repository.UpdateFee(transactionId, fee)
	assert.NotNil(t, err)
}

func Test_UpdateStatusCompleted(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery,
		status.Completed,
		transactionId)

	err := repository.UpdateStatusCompleted(transactionId)
	assert.Nil(t, err)
}

func Test_UpdateStatusCompleted_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery,
		status.Completed,
		transactionId)

	err := repository.UpdateStatusCompleted(transactionId)
	assert.NotNil(t, err)
}

func Test_UpdateStatusFailed(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery,
		status.Failed,
		transactionId)

	err := repository.UpdateStatusFailed(transactionId)
	assert.Nil(t, err)
}

func Test_UpdateStatusFailed_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery,
		status.Failed,
		transactionId)

	err := repository.UpdateStatusFailed(transactionId)
	assert.NotNil(t, err)
}

func Test_create(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, createQuery,
		transactionId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		sourceAsset,
		targetAsset,
		nativeAsset,
		receiver,
		amount,
		"", //fee
		someStatus,
		serialNumber,
		metadata,
		isNft,
		nanoTime,
		originator)

	actual, err := repository.create(expectedModelTransfer, someStatus)
	assert.Nil(t, err)
	assert.Equal(t, expectedEntityTransfer, actual)
}

func Test_create_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, createQuery,
		transactionId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		sourceAsset,
		targetAsset,
		nativeAsset,
		receiver,
		amount,
		"", //fee
		someStatus,
		serialNumber,
		metadata,
		isNft,
		nanoTime,
		originator)

	actual, err := repository.create(expectedModelTransfer, someStatus)
	assert.NotNil(t, err)
	assert.NotNil(t, actual)
}

func Test_updateStatus(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery,
		status.Initial,
		transactionId)

	err := repository.updateStatus(transactionId, status.Initial)
	assert.Nil(t, err)
}

func Test_updateStatus_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery,
		status.Initial,
		transactionId)

	err := repository.updateStatus(transactionId, status.Initial)
	assert.NotNil(t, err)
}
