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
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type pgConnector struct {
	connString string
}

func NewPgConnector(cfg config.Database) *pgConnector {
	connString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Name, cfg.Password)

	return &pgConnector{
		connString,
	}
}

// Connect establishes a connection to Postgresql
func (c *pgConnector) Connect() *gorm.DB {
	db := c.tryConnection()
	log.Infoln("Successfully connected to Database")

	return db
}

// TryConnection, tries to connect to the database associated to the validator node. If it fails, it retries after 10 seconds.
// This function will try to reconnect until it succeeds or the validator node gets stopped manually
func (c *pgConnector) tryConnection() *gorm.DB {
	db, err := gorm.Open(
		postgres.Open(c.connString),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)

	if err != nil {
		log.Error(err)
		time.Sleep(10 * time.Second)
		log.Infof("Retrying to connect to DB with connection string [%s]", c.connString)
		return c.tryConnection()
	}
	return db
}
