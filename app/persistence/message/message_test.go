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

package message

import (
	"database/sql"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

var (
	repository   *Repository
	dbConnection *gorm.DB
	sqlMock      sqlmock.Sqlmock
	db           *sql.DB

	insertQuery                   = regexp.QuoteMeta(`INSERT INTO "messages" ("transfer_id","hash","signature","signer","transaction_timestamp") VALUES ($1,$2,$3,$4,$5)`)
	selectQuery                   = regexp.QuoteMeta(`SELECT * FROM "messages" WHERE transfer_id = $1 and signature = $2 and hash = $3 ORDER BY "messages"."transfer_id" LIMIT 1`)
	selectByTransferIdQuery       = regexp.QuoteMeta(`SELECT * FROM "messages" WHERE transfer_id = $1 ORDER BY transaction_timestamp`)
	selectTransferForeignKeyQuery = regexp.QuoteMeta(`SELECT * FROM "transfers" WHERE "transfers"."transaction_id" = $1`)

	transferId           = "someTransferId"
	transfer             = entity.Transfer{}
	signature            = "someSignature"
	hash                 = "someHash"
	signer               = "someSigner"
	transactionTimestamp = time.Now().UnixNano()
	columns              = []string{"transfer_id", "hash", "signature", "signer", "transaction_timestamp"}
	rowArgs              = []driver.Value{transferId, hash, signature, signer, transactionTimestamp}
	expectedMsg          = &entity.Message{
		TransferID:           transferId,
		Transfer:             transfer,
		Hash:                 hash,
		Signature:            signature,
		Signer:               signer,
		TransactionTimestamp: transactionTimestamp,
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
	helper.SqlMockPrepareExec(sqlMock, insertQuery, transferId, hash, signature, signer, transactionTimestamp)

	err := repository.Create(expectedMsg)

	assert.Nil(t, err)
}

func Test_Create_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareExecWithErr(sqlMock, insertQuery, transferId, hash, signature, signer, transactionTimestamp)

	err := repository.Create(expectedMsg)

	assert.Error(t, err, expectedErr)
}

func Test_GetMessageWith(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, columns, rowArgs, selectQuery, transferId, signature, hash)

	fetchedMsg, err := repository.GetMessageWith(transferId, signature, hash)

	assert.Nil(t, err)
	assert.Equal(t, expectedMsg, fetchedMsg)
}

func Test_GetMessageWith_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, selectQuery, transferId, signature, hash)

	fetchedMsg, err := repository.GetMessageWith(transferId, signature, hash)

	assert.Error(t, err, expectedErr)
	assert.Nil(t, fetchedMsg)
}

func Test_Exist(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, columns, rowArgs, selectQuery, transferId, signature, hash)

	exist, err := repository.Exist(transferId, signature, hash)

	assert.Nil(t, err)
	assert.True(t, exist)
}

func Test_Exist_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	_ = helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, selectQuery, transferId, signature, hash)
	expectedErr2 := helper.SqlMockPrepareQueryWithErrInvalidData(sqlMock, selectQuery, transferId, signature, hash)

	exist1, err1 := repository.Exist(transferId, signature, hash)
	exist2, err2 := repository.Exist(transferId, signature, hash)

	assert.Nil(t, err1)
	assert.False(t, exist1)

	assert.Error(t, err2, expectedErr2)
	assert.False(t, exist2)
}

func Test_Get(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	helper.SqlMockPrepareQuery(sqlMock, columns, rowArgs, selectByTransferIdQuery, transferId)
	sqlMock.ExpectQuery(selectTransferForeignKeyQuery).WithArgs(transferId).WillReturnRows(&sqlmock.Rows{})

	fetchedMessages, err := repository.Get(transferId)

	assert.Nil(t, err)
	assert.Len(t, fetchedMessages, 1)
	assert.Equal(t, *expectedMsg, fetchedMessages[0])
}

func Test_Get_Err(t *testing.T) {
	setup()
	defer helper.CheckSqlMockExpectationsMet(sqlMock, t)
	expectedErr := helper.SqlMockPrepareQueryWithErrNotFound(sqlMock, selectByTransferIdQuery, transferId)

	fetchedMessages, err := repository.Get(transferId)

	assert.Error(t, err, expectedErr)
	assert.Len(t, fetchedMessages, 0)
}

func setup() {
	mocks.Setup()
	dbConnection, sqlMock, db = helper.SetupSqlMock()

	repository = &Repository{
		dbClient: dbConnection,
	}
}
