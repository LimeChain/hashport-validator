package persistence

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

// Establish connection to the Postgres Database
func connectToDb(dbConfig config.Db) *gorm.DB {
	connectionStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Name, dbConfig.Password)
	db, err := gorm.Open(postgres.Open(connectionStr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully connected to Database")

	return db
}

// Migrate tables
func migrateDb(db *gorm.DB) {
	err := db.AutoMigrate(status.Status{})
	if err != nil {
		log.Fatal(err)
	}
}

// Connect and Migrate
func RunDb(dbConfig config.Db) *gorm.DB {
	gorm := connectToDb(dbConfig)
	migrateDb(gorm)

	return gorm
}
