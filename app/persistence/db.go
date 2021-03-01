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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/scheduled"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Establish connection to the Postgres Database
func connectToDb(dbConfig config.Db) *gorm.DB {
	connectionStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Name, dbConfig.Password)
	db, err := gorm.Open(
		postgres.Open(connectionStr),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Infoln("Successfully connected to Database")

	return db
}

// Migrate tables
func migrateDb(db *gorm.DB) {
	err := db.AutoMigrate(
		status.Status{},
		transaction.Transaction{},
		message.TransactionMessage{},
		scheduled.Scheduled{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Migrations passed successfully")
}

// Connect and Migrate
func RunDb(dbConfig config.Db) *gorm.DB {
	gorm := connectToDb(dbConfig)
	migrateDb(gorm)
	return gorm
}
