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
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

var (
	repository          *Repository
	dbConnection        *gorm.DB
	sqlMock             sqlmock.Sqlmock
	db                  *sql.DB
	entityColumns       = []string{"entity_id", "last"}
	entityArgs          = []driver.Value{entityId, entityLastTimestamp}
	entityId            = "1"
	entityLastTimestamp = time.Now().UnixNano()
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
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	query := regexp.QuoteMeta(`INSERT INTO "statuses" ("entity_id","last") VALUES ($1,$2)`)
	helper.SqlMockPrepareExec(sqlMock, query, entityId, entityLastTimestamp)

	err := repository.Create(entityId, entityLastTimestamp)

	assert.Nil(t, err)
}

func Test_Create_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	query := regexp.QuoteMeta(`INSERT INTO "statuses" ("entity_id","last") VALUES ($1,$2)`)
	expectedErr := helper.SqlMockPrepareExecWithErr(sqlMock, query, entityId, entityLastTimestamp)

	err := repository.Create(entityId, entityLastTimestamp)

	assert.Error(t, err, expectedErr)
}

func Test_Update(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	updatedLast := int64(5)
	query := regexp.QuoteMeta(`UPDATE "statuses" SET "entity_id"=$1,"last"=$2 WHERE entity_id = $3`)
	helper.SqlMockPrepareExec(sqlMock, query, entityId, updatedLast, entityId)

	err := repository.Update(entityId, updatedLast)

	assert.Nil(t, err)
}

func Test_Update_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	updatedLast := int64(5)
	query := regexp.QuoteMeta(`UPDATE "statuses" SET "entity_id"=$1,"last"=$2 WHERE entity_id = $3`)
	expectedErr := helper.SqlMockPrepareExecWithErr(sqlMock, query, entityId, updatedLast, entityId)

	err := repository.Update(entityId, updatedLast)

	assert.Error(t, err, expectedErr)
}

func Test_Get(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	query := regexp.QuoteMeta(`SELECT * FROM "statuses" WHERE entity_id = $1 ORDER BY "statuses"."entity_id" LIMIT 1`)
	helper.SqlMockPrepareQuery(sqlMock, entityColumns, entityArgs, query, entityId)

	lastTimestamp, err := repository.Get(entityId)

	assert.Nil(t, err)
	assert.Equal(t, entityLastTimestamp, lastTimestamp)
}

func Test_Get_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	query := regexp.QuoteMeta(`SELECT * FROM "statuses" WHERE entity_id = $1 ORDER BY "statuses"."entity_id" LIMIT 1`)
	expectedErr := helper.SqlMockPrepareQueryWithErr(sqlMock, query, entityId)

	lastTimestamp, err := repository.Get(entityId)

	assert.Error(t, err, expectedErr)
	assert.Equal(t, int64(0), lastTimestamp)
}

func setup() {
	mocks.Setup()
	dbConnection, sqlMock, db = helper.SetupSqlMock()

	repository = &Repository{
		dbClient: dbConnection,
	}
}
