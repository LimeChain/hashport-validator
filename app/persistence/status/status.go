package status

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/timestamp"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// This table will contain information for latest status of the application
type Status struct {
	Name         string
	EntityID     string
	Code         string
	WholeValue   int64
	DecimalValue int64
}

type StatusRepository struct {
	dbClient                 *gorm.DB
	lastFetchedTimestampCode string //"LAST_FETCHED_TIMESTAMP"
}

func NewStatusRepository(dbClient *gorm.DB, statusType string) *StatusRepository {
	typeCheck(statusType)
	return &StatusRepository{
		dbClient:                 dbClient,
		lastFetchedTimestampCode: fmt.Sprintf("LAST_%s_TIMESTAMP", statusType),
	}
}

func typeCheck(statusType string) {
	switch statusType {
	case "HCS_TOPIC_MSG":
	case "HCS_CRYPTO_TRANSFER":
		return
	default:
		log.Fatal("Invalid status type.")
	}
}

func (s StatusRepository) GetLastFetchedTimestamp(entityID string) (*timestamp.Timestamp, error) {
	lastFetchedStatus := &Status{}
	err := s.dbClient.
		Model(&Status{}).
		Where("code = ? and entity_id = ?", s.lastFetchedTimestampCode, entityID).
		First(&lastFetchedStatus).Error
	if err != nil {
		return nil, err
	}
	return timestamp.NewTimestamp(lastFetchedStatus.WholeValue, 0), nil
}

func (s StatusRepository) CreateTimestamp(entityID string, timestamp *timestamp.Timestamp) error {
	return s.dbClient.Create(Status{
		Name:         "Last fetched timestamp",
		EntityID:     entityID,
		Code:         s.lastFetchedTimestampCode,
		WholeValue:   timestamp.Whole,
		DecimalValue: timestamp.Dec,
	}).Error
}

func (s StatusRepository) UpdateLastFetchedTimestamp(entityID string, timestamp *timestamp.Timestamp) error {
	return s.dbClient.
		Where("code = ? and entity_id = ?", s.lastFetchedTimestampCode, entityID).
		Save(Status{
			Name:         "Last fetched timestamp",
			EntityID:     entityID,
			Code:         s.lastFetchedTimestampCode,
			WholeValue:   timestamp.Whole,
			DecimalValue: timestamp.Dec,
		}).
		Error
}
