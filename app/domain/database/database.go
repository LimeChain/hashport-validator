package database

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"gorm.io/gorm"
)

type Database interface {
	ConnectWithMigration(config config.Database) *gorm.DB
}
