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

package helper

import (
	"database/sql"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func CheckSqlMockExpectationsMet(sqlMock sqlmock.Sqlmock, t *testing.T) {
	// we make sure that all expectations were met
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func SetupSqlMock() (dbConnection *gorm.DB, sqlMock sqlmock.Sqlmock, db *sql.DB) {
	var err error

	db, sqlMock, err = sqlmock.New()
	if err != nil {
		panic("failed to initialize 'sqlmock'")
	}

	dbConnection, err = gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		panic("failed to open 'gorm.Db' connection")
	}

	mocks.MDatabase.On("Connection").Return(dbConnection)
	dbConnection = mocks.MDatabase.Connection()

	return dbConnection, sqlMock, db
}

// Exec //

func SqlMockPrepareExec(sqlMock sqlmock.Sqlmock, query string, queryArgs ...driver.Value) {
	result := sqlmock.NewResult(1, 1)
	sqlMock.ExpectExec(query).WithArgs(queryArgs...).WillReturnResult(result)
}

func SqlMockPrepareExecWithErr(sqlMock sqlmock.Sqlmock, query string, queryArgs ...driver.Value) error {
	err := gorm.ErrInvalidData
	result := sqlmock.NewResult(0, 0)
	sqlMock.ExpectExec(query).WithArgs(queryArgs...).WillReturnResult(result).WillReturnError(err)

	return err
}

// Query //

func SqlMockPrepareQuery(sqlMock sqlmock.Sqlmock, columns []string, rowArgs []driver.Value, query string, queryArgs ...driver.Value) {
	msgRow := sqlmock.NewRows(columns).AddRow(rowArgs...)
	sqlMock.ExpectQuery(query).WithArgs(queryArgs...).WillReturnRows(msgRow)
}

func SqlMockPrepareQueryWithErrNotFound(sqlMock sqlmock.Sqlmock, query string, queryArgs ...driver.Value) error {
	err := gorm.ErrRecordNotFound
	sqlMock.ExpectQuery(query).WithArgs(queryArgs...).WillReturnError(err)

	return err
}

func SqlMockPrepareQueryWithErrInvalidData(sqlMock sqlmock.Sqlmock, query string, queryArgs ...driver.Value) error {
	err := gorm.ErrInvalidData
	sqlMock.ExpectQuery(query).WithArgs(queryArgs...).WillReturnError(err)

	return err
}
