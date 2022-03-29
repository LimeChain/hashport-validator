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

package status

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

var (
	repository          *Repository
	dbConnection        *gorm.DB
	sqlMock             sqlmock.Sqlmock
	db                  *sql.DB
	entityId            = "1"
	entityLastTimestamp = int64(10)
)

func Test_NewRepositoryForStatus(t *testing.T) {
	setup()

	cases := []struct {
		status      string
		expectFatal bool
	}{
		{
			status:      Transfer,
			expectFatal: false,
		},
		{
			status:      Message,
			expectFatal: false,
		},
		{
			status:      "",
			expectFatal: true,
		},
	}

	// Testing all cases including the one which needs to call Fatal and exit
	defer func() { log.StandardLogger().ExitFunc = nil }()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	for _, c := range cases {
		fatal = false
		actualRepository := NewRepositoryForStatus(dbConnection, c.status)
		msg := "testing NewStatusRepository where:\n  status - '%s'\n  is fatal expected - %v\n  fatal called - %v."
		assert.Equalf(t, c.expectFatal, fatal, msg, c.status, c.expectFatal, fatal)
		if !c.expectFatal {
			assert.Equal(t, repository, actualRepository)
		}
	}
}

func Test_Create(t *testing.T) {
	setup()
	defer checkSqlMockExpectationsMet(t)
	query := regexp.QuoteMeta(`INSERT INTO "statuses" ("entity_id","last") VALUES ($1,$2)`)
	result := sqlmock.NewResult(1, 1)
	sqlMock.ExpectExec(query).WithArgs(entityId, entityLastTimestamp).WillReturnResult(result)

	err := repository.Create(entityId, entityLastTimestamp)

	assert.Nil(t, err)
}

func Test_Create_Err(t *testing.T) {
	setup()
	defer checkSqlMockExpectationsMet(t)
	query := regexp.QuoteMeta(`INSERT INTO "statuses" ("entity_id","last") VALUES ($1,$2)`)
	result := sqlmock.NewResult(0, 0)
	returnErr := errors.New("failed to update record")
	sqlMock.ExpectExec(query).WithArgs(entityId, entityLastTimestamp).WillReturnResult(result).WillReturnError(returnErr)

	err := repository.Create(entityId, entityLastTimestamp)

	assert.Error(t, err, returnErr)
}

func Test_Update(t *testing.T) {
	setup()
	defer checkSqlMockExpectationsMet(t)
	updatedLast := int64(5)
	query := regexp.QuoteMeta(`UPDATE "statuses" SET "entity_id"=$1,"last"=$2 WHERE entity_id = $3`)
	result := sqlmock.NewResult(1, 1)
	sqlMock.ExpectExec(query).WithArgs(entityId, updatedLast, entityId).WillReturnResult(result)

	err := repository.Update(entityId, updatedLast)

	assert.Nil(t, err)
}

func Test_Update_Err(t *testing.T) {
	setup()
	defer checkSqlMockExpectationsMet(t)
	updatedLast := int64(5)
	query := regexp.QuoteMeta(`UPDATE "statuses" SET "entity_id"=$1,"last"=$2 WHERE entity_id = $3`)
	result := sqlmock.NewResult(0, 0)
	returnErr := errors.New("failed to update record")
	sqlMock.ExpectExec(query).WithArgs(entityId, updatedLast, entityId).WillReturnResult(result).WillReturnError(returnErr)

	err := repository.Update(entityId, updatedLast)

	assert.Error(t, err, returnErr)
}

func Test_Get(t *testing.T) {
	setup()
	defer checkSqlMockExpectationsMet(t)
	rows := sqlmock.NewRows([]string{"entity_id", "last"}).AddRow(entityId, entityLastTimestamp)
	query := regexp.QuoteMeta(`SELECT * FROM "statuses" WHERE entity_id = $1 ORDER BY "statuses"."entity_id" LIMIT 1`)
	sqlMock.ExpectQuery(query).WithArgs(entityId).WillReturnRows(rows)

	lastTimestamp, err := repository.Get(entityId)

	assert.Nil(t, err)
	assert.Equal(t, entityLastTimestamp, lastTimestamp)
}

func Test_Get_Err(t *testing.T) {
	setup()
	defer checkSqlMockExpectationsMet(t)
	query := regexp.QuoteMeta(`SELECT * FROM "statuses" WHERE entity_id = $1 ORDER BY "statuses"."entity_id" LIMIT 1`)
	returnErr := errors.New("no record found")
	sqlMock.ExpectQuery(query).WithArgs(entityId).WillReturnError(returnErr)

	lastTimestamp, err := repository.Get(entityId)

	assert.Error(t, err, returnErr)
	assert.Equal(t, int64(0), lastTimestamp)
}

func checkSqlMockExpectationsMet(t *testing.T) {
	// we make sure that all expectations were met
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func setup() {
	mocks.Setup()
	var err error

	db, sqlMock, err = sqlmock.New()
	if err != nil {
		panic("failed to initialize 'sqlmock'")
	}

	dbConnection, err = gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		panic("failed to open 'gorm.Db' connection")
	}

	mocks.MDatabase.On("GetConnection").Return(dbConnection)
	dbConnection = mocks.MDatabase.GetConnection()

	repository = &Repository{
		dbClient: dbConnection,
	}
}
