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

package fee

import (
	"database/sql"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	entityStatus "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

var (
	repository    *Repository
	dbConn        *gorm.DB
	sqlMock       sqlmock.Sqlmock
	db            *sql.DB
	transactionId = "transactionId"
	scheduleId    = "scheduleId"
	amount        = "1"
	someStatus    = entityStatus.Completed
	transferId    = "transferId"
	expectedFee   = &entity.Fee{
		TransactionID: transactionId,
		ScheduleID:    scheduleId,
		Amount:        amount,
		Status:        someStatus,
		TransferID:    sql.NullString{String: transferId, Valid: true},
	}
	rowArgs = []driver.Value{transactionId, scheduleId, amount, someStatus, transferId}
	columns = []string{"transaction_id", "schedule_id", "amount", "status", "transfer_id"}

	feeQuery                = regexp.QuoteMeta(`SELECT * FROM "fees" WHERE transaction_id = $1 ORDER BY "fees"."transaction_id" LIMIT 1`)
	createQuery             = regexp.QuoteMeta(`INSERT INTO "fees" ("transaction_id","schedule_id","amount","status","transfer_id") VALUES ($1,$2,$3,$4,$5)`)
	updateStatusQuery       = regexp.QuoteMeta(`UPDATE "fees" SET "status"=$1 WHERE transaction_id = $2`)
	getAllSubmittedIdsQuery = regexp.QuoteMeta(`SELECT "transaction_id" FROM "fees" WHERE status = $1`)
)

func setup() {
	mocks.Setup()
	dbConn, sqlMock, db = helper.SetupSqlMock()

	repository = &Repository{
		dbClient: dbConn,
		logger:   config.GetLoggerFor("Fee Repository"),
	}
}

func Test_Get(t *testing.T) {
	setup()
	helper.SqlMockPrepareQuery(sqlMock, columns, rowArgs, feeQuery, transactionId)

	actualFee, err := repository.Get(transactionId)
	assert.Nil(t, err)
	assert.Equal(t, expectedFee, actualFee)
}

func Test_Get_NotFound(t *testing.T) {
	setup()
	_ = helper.SqlMockPrepareQueryWithErr(sqlMock, feeQuery, transactionId)

	actual, err := repository.Get(transactionId)
	assert.Nil(t, err)
	assert.Nil(t, actual)
}

func Test_Create(t *testing.T) {
	setup()
	helper.SqlMockPrepareExec(sqlMock, createQuery,
		expectedFee.TransactionID,
		expectedFee.ScheduleID,
		expectedFee.Amount,
		expectedFee.Status,
		expectedFee.TransferID)

	err := repository.Create(expectedFee)
	assert.Nil(t, err)
}

func Test_Create_Err(t *testing.T) {
	setup()
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, createQuery,
		expectedFee.TransactionID,
		expectedFee.ScheduleID,
		expectedFee.Amount,
		expectedFee.Status,
		expectedFee.TransferID)

	err := repository.Create(expectedFee)
	assert.NotNil(t, err)
}

func Test_UpdateStatusCompleted(t *testing.T) {
	setup()
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery,
		entityStatus.Completed,
		expectedFee.TransactionID)

	err := repository.UpdateStatusCompleted(transactionId)
	assert.Nil(t, err)
}

func Test_UpdateStatusCompleted_Err(t *testing.T) {
	setup()
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery,
		entityStatus.Completed,
		expectedFee.TransactionID)

	err := repository.UpdateStatusCompleted(transactionId)
	assert.NotNil(t, err)
}

func Test_UpdateStatusFailed(t *testing.T) {
	setup()
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery,
		entityStatus.Failed,
		expectedFee.TransactionID)

	err := repository.UpdateStatusFailed(transactionId)
	assert.Nil(t, err)
}

func Test_UpdateStatusFailed_Err(t *testing.T) {
	setup()
	_ = helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery,
		entityStatus.Failed,
		expectedFee.TransactionID)

	err := repository.UpdateStatusCompleted(transactionId)
	assert.NotNil(t, err)
}

func Test_GetAllSubmittedIds(t *testing.T) {
	setup()
	helper.SqlMockPrepareQuery(sqlMock, columns, rowArgs, getAllSubmittedIdsQuery,
		entityStatus.Submitted)

	actual, err := repository.GetAllSubmittedIds()
	assert.Nil(t, err)
	assert.NotEmpty(t, actual)
}

func Test_GetAllSubmittedIds_Err(t *testing.T) {
	setup()
	_ = helper.SqlMockPrepareQueryWithErr(sqlMock, getAllSubmittedIdsQuery,
		entityStatus.Submitted)

	actual, err := repository.GetAllSubmittedIds()
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}
