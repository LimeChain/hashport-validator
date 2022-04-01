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

package persistence

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

var (
	db      *Database
	dbConn  *gorm.DB
	sqlMock sqlmock.Sqlmock

	createTransfers = regexp.QuoteMeta(`CREATE TABLE "transfers" ("transaction_id" text,"source_chain_id" bigint,"target_chain_id" bigint,"native_chain_id" bigint,"source_asset" text,"target_asset" text,"native_asset" text,"receiver" text,"amount" text,"fee" text,"status" text,"serial_number" bigint,"metadata" text,"is_nft" boolean DEFAULT false,PRIMARY KEY ("transaction_id"))`)
	createFees      = regexp.QuoteMeta(`CREATE TABLE "fees" ("transaction_id" text,"schedule_id" text,"amount" text,"status" text,"transfer_id" text,PRIMARY KEY ("transaction_id"),CONSTRAINT "fk_transfers_fees" FOREIGN KEY ("transfer_id") REFERENCES "transfers"("transaction_id"))`)
	createMessages  = regexp.QuoteMeta(`CREATE TABLE "messages" ("transfer_id" text,"hash" text,"signature" text UNIQUE,"signer" text,"transaction_timestamp" bigint,CONSTRAINT "fk_messages_transfer" FOREIGN KEY ("transfer_id") REFERENCES "transfers"("transaction_id"),CONSTRAINT "fk_transfers_messages" FOREIGN KEY ("transfer_id") REFERENCES "transfers"("transaction_id"))`)
	createSchedules = regexp.QuoteMeta(`CREATE TABLE "schedules" ("transaction_id" text,"schedule_id" text,"has_receiver" boolean,"operation" text,"status" text,"transfer_id" text,PRIMARY KEY ("transaction_id"),CONSTRAINT "fk_transfers_schedules" FOREIGN KEY ("transfer_id") REFERENCES "transfers"("transaction_id"))`)
	createStatuses  = regexp.QuoteMeta(`CREATE TABLE "statuses" ("entity_id" text,"last" bigint)`)
)

func setupDatabase() {
	mocks.Setup()
	dbConn, sqlMock, _ = helper.SetupSqlMock()

	db = &Database{
		connector: mocks.MConnector,
	}
}

func Test_NewDatabase(t *testing.T) {
	setupDatabase()

	actual := NewDatabase(mocks.MConnector)
	assert.NotNil(t, actual)
	assert.Equal(t, db, actual)
}

func Test_GetConnectionInitial(t *testing.T) {
	setupDatabase()

	mocks.MConnector.On("Connect").Return(dbConn)

	actual := db.Connection()
	assert.Equal(t, dbConn, actual)
}

func Test_GetConnectionAfterInit(t *testing.T) {
	setupDatabase()

	db.connection = dbConn
	actual := db.Connection()

	assert.Equal(t, dbConn, actual)
	mocks.MConnector.AssertNotCalled(t, "Connect")
}

func Test_Migrate(t *testing.T) {
	setupDatabase()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareExec(sqlMock, createTransfers)
	helper.SqlMockPrepareExec(sqlMock, createFees)
	helper.SqlMockPrepareExec(sqlMock, createMessages)
	helper.SqlMockPrepareExec(sqlMock, createSchedules)
	helper.SqlMockPrepareExec(sqlMock, createStatuses)

	mocks.MConnector.On("Connect").Return(dbConn)

	err := db.Migrate()
	assert.Nil(t, err)
}
