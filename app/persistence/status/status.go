package status

import (
	"fmt"
	"gorm.io/gorm"
	"log"
)

// This table will contain information for latest status of the application
type Status struct {
	Name     string
	EntityID string
	Code     string
	Value    string
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
	case "HCS_TOPIC":
	case "CRYPTO_TRANSFER":
		return
	default:
		log.Fatal("Invalid status type.")
	}
}

func (s StatusRepository) GetLastFetchedTimestamp(entityID string) (string, error) {
	lastFetchedStatus := &Status{}
	err := s.dbClient.
		Model(&Status{}).
		Where("code = ? and entity_id = ?", s.lastFetchedTimestampCode, entityID).
		First(&lastFetchedStatus).Error
	if err != nil {
		return "", err
	}
	return lastFetchedStatus.Value, nil
}

func (s StatusRepository) CreateTimestamp(entityID string, timestamp string) error {
	return s.dbClient.Create(Status{
		Name:     "Last fetched timestamp",
		EntityID: entityID,
		Code:     s.lastFetchedTimestampCode,
		Value:    timestamp,
	}).Error
}

func (s StatusRepository) UpdateLastFetchedTimestamp(entityID string, timestamp string) error {
	return s.dbClient.
		Where("code = ? and entity_id = ?", s.lastFetchedTimestampCode, entityID).
		Save(Status{
			Name:     "Last fetched timestamp",
			EntityID: entityID,
			Code:     s.lastFetchedTimestampCode,
			Value:    timestamp,
		}).
		Error
}
