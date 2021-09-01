/*
 * Copyright 2021 LimeChain Ltd.
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
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Database struct {
	connection *gorm.DB
}

func (db *Database) GetConnection() *gorm.DB {
	return db.connection
}

func NewDatabase(config config.Database) *Database {
	return &Database{
		connection: ConnectWithMigration(config),
	}
}

// Establish connection to the Postgres Database
func Connect(dbConfig config.Database) *gorm.DB {
	connectionStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Name, dbConfig.Password)

	db := tryConnection(connectionStr)
	log.Infoln("Successfully connected to Database")

	return db
}

// TryConnection, tries to connect to the database associated to the validator node. If it fails, it retries after 10 seconds.
// This function will try to reconnect until it succeeds or the validator node gets stopped manually
func tryConnection(connectionStr string) *gorm.DB {
	db, err := gorm.Open(
		postgres.Open(connectionStr),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)

	if err != nil {
		log.Error(err)
		time.Sleep(10 * time.Second)
		log.Infof("Retrying to connect to DB with connection string [%s]", connectionStr)
		return tryConnection(connectionStr)
	}
	return db
}

// Migrate tables
func migrateDb(db *gorm.DB) {
	err := db.AutoMigrate(
		entity.Transfer{},
		entity.Fee{},
		entity.Message{},
		entity.Schedule{},
		entity.Status{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Migrations passed successfully")
}

// Connect and Migrate
func ConnectWithMigration(config config.Database) *gorm.DB {
	gorm := Connect(config)
	migrateDb(gorm)
	return gorm
}
