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

package schedule

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	repository   *Repository
	dbConnection *gorm.DB
	sqlMock      sqlmock.Sqlmock
	db           *sql.DB

	insertQuery                 = regexp.QuoteMeta(`INSERT INTO "schedules" ("transaction_id","schedule_id","has_receiver","operation","status","transfer_id") VALUES ($1,$2,$3,$4,$5,$6)`)
	updateStatusQuery           = regexp.QuoteMeta(`UPDATE "schedules" SET "status"=$1 WHERE transaction_id = $2`)
	selectQuery                 = regexp.QuoteMeta(`SELECT * FROM "schedules" WHERE transaction_id = $1 ORDER BY "schedules"."transaction_id" LIMIT 1`)
	selectIdsByStatusQuery      = regexp.QuoteMeta(`SELECT "transaction_id" FROM "schedules" WHERE status = $1`)
	selectReceiverTransferQuery = regexp.QuoteMeta(`SELECT * FROM "schedules" WHERE transfer_id = $1 AND operation = $2 AND has_receiver = true ORDER BY "schedules"."transaction_id" LIMIT 1`)

	transactionId  = "someTransactionId"
	scheduleId     = "someScheduleId"
	hasReceiver    = true
	operation      = "someOperation"
	expectedStatus = status.Submitted
	transferId     = sql.NullString{String: "someTransferId", Valid: true}

	entityColumns = []string{"transaction_id", "schedule_id", "has_receiver", "operation", "status", "transfer_id"}
	entityArgs    = []driver.Value{transactionId, scheduleId, hasReceiver, operation, expectedStatus, transferId}

	expectedSchedule = &entity.Schedule{
		TransactionID: transactionId,
		ScheduleID:    scheduleId,
		HasReceiver:   hasReceiver,
		Operation:     operation,
		Status:        expectedStatus,
		TransferID:    transferId,
	}
)

func Test_NewRepository(t *testing.T) {
	setup()

	actualRepository := NewRepository(dbConnection)

	assert.Equal(t, repository, actualRepository)
}

func Test_Create(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, insertQuery, transactionId, scheduleId, hasReceiver, operation, expectedStatus, transferId)

	err := repository.Create(expectedSchedule)

	assert.Nil(t, err)
}

func Test_Create_Error(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareExecWithErr(sqlMock, insertQuery, transactionId, scheduleId, hasReceiver, operation, expectedStatus, transferId)

	err := repository.Create(expectedSchedule)

	assert.Error(t, err, expectedErr)
}

func Test_UpdateStatusCompleted(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery, status.Completed, transactionId)

	err := repository.UpdateStatusCompleted(transactionId)

	assert.Nil(t, err)
}

func Test_UpdateStatusCompleted_Error(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery, status.Completed, transactionId)

	err := repository.UpdateStatusCompleted(transactionId)

	assert.Error(t, err, expectedErr)
}

func Test_UpdateStatusFailed(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, updateStatusQuery, status.Failed, transactionId)

	err := repository.UpdateStatusFailed(transactionId)

	assert.Nil(t, err)
}

func Test_UpdateStatusFailed_Error(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareExecWithErr(sqlMock, updateStatusQuery, status.Failed, transactionId)

	err := repository.UpdateStatusFailed(transactionId)

	assert.Error(t, err, expectedErr)
}

func Test_Get(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, entityColumns, entityArgs, selectQuery, transactionId)

	fetchedSchedule, err := repository.Get(transactionId)

	assert.Nil(t, err)
	assert.Equal(t, expectedSchedule, fetchedSchedule)
}

func Test_Get_Error(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr1 := helper.SqlMockPrepareQueryWithErrInvalidData(sqlMock, selectQuery, transactionId)

	fetchedSchedule1, err1 := repository.Get(transactionId)

	_ = helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, selectQuery, transactionId)
	fetchedSchedule2, err2 := repository.Get(transactionId)

	assert.Error(t, err1, expectedErr1)
	assert.Nil(t, fetchedSchedule1)
	assert.Nil(t, fetchedSchedule2)
	assert.Nil(t, err2)
}

func Test_GetReceiverTransferByTransactionID(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, entityColumns, entityArgs, selectReceiverTransferQuery, transferId.String, schedule.TRANSFER)

	fetchedSchedule, err := repository.GetReceiverTransferByTransactionID(transferId.String)

	assert.Nil(t, err)
	assert.Equal(t, expectedSchedule, fetchedSchedule)
}

func Test_GetReceiverTransferByTransactionID_Error(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr1 := helper.SqlMockPrepareQueryWithErrInvalidData(sqlMock, selectReceiverTransferQuery, transferId.String, schedule.TRANSFER)

	fetchedSchedule1, err1 := repository.GetReceiverTransferByTransactionID(transferId.String)

	_ = helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, selectReceiverTransferQuery, transferId.String, schedule.TRANSFER)
	fetchedSchedule2, err2 := repository.GetReceiverTransferByTransactionID(transferId.String)

	assert.Error(t, err1, expectedErr1)
	assert.Nil(t, fetchedSchedule1)
	assert.Nil(t, fetchedSchedule2)
	assert.Nil(t, err2)
}

func Test_GetAllSubmittedIds(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, entityColumns, entityArgs, selectIdsByStatusQuery, expectedStatus)

	fetchedSchedules, err := repository.GetAllSubmittedIds()

	assert.Nil(t, err)
	assert.Len(t, fetchedSchedules, 1)
	assert.Equal(t, expectedSchedule, fetchedSchedules[0])
}

func Test_GetAllSubmittedIds_Error(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, selectIdsByStatusQuery, expectedStatus)

	fetchedSchedule, err := repository.GetAllSubmittedIds()

	assert.Error(t, err, expectedErr)
	assert.Nil(t, fetchedSchedule)
}

func setup() {
	mocks.Setup()
	dbConnection, sqlMock, db = helper.SetupSqlMock()

	repository = &Repository{
		db:     dbConnection,
		logger: config.GetLoggerFor("Transfer Repository"),
	}
}
