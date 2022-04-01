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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/database"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Database struct {
	connection *gorm.DB
	connector  database.Connector
}

func NewDatabase(connector database.Connector) *Database {
	return &Database{
		connector: connector,
	}
}

// Connection establishes a connection or returns an existing one
func (db *Database) Connection() *gorm.DB {
	if db.connection != nil {
		return db.connection
	}

	pgdb := db.connector.Connect()
	db.connection = pgdb
	return db.connection
}

// Migrate auto-migrates the database
func (db *Database) Migrate() error {
	err := db.Connection().
		AutoMigrate(
			entity.Transfer{},
			entity.Fee{},
			entity.Message{},
			entity.Schedule{},
			entity.Status{})
	if err != nil {
		return err
	}
	log.Println("Migrations passed successfully")
	return nil
}
