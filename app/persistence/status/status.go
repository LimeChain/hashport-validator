package status

import (
	"github.com/hashgraph/hedera-sdk-go"
	"gorm.io/gorm"
)

var (
	lastFetchedTimestampCode = "LAST_FETCHED_TIMESTAMP"
)

// This table will contain information for latest status of the application
type Status struct {
	Name      string
	AccountID string
	Code      string
	Value     string
}

type StatusRepository struct {
	dbClient *gorm.DB
}

func NewStatusRepository(dbClient *gorm.DB) *StatusRepository {
	return &StatusRepository{
		dbClient: dbClient,
	}
}

func (s StatusRepository) GetLastFetchedTimestamp(accountID hedera.AccountID) (string, error) {
	lastFetchedStatus := &Status{}
	err := s.dbClient.
		Model(&Status{}).
		Where("code = ? and account_id = ?", lastFetchedTimestampCode, accountID.String()).
		First(&lastFetchedStatus).Error
	if err != nil {
		return "", err
	}
	return lastFetchedStatus.Value, nil
}

func (s StatusRepository) CreateTimestamp(accountID hedera.AccountID, timestamp string) error {
	err := s.dbClient.Create(Status{
		Name:      "Last fetched timestamp",
		AccountID: accountID.String(),
		Code:      lastFetchedTimestampCode,
		Value:     timestamp,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (s StatusRepository) UpdateLastFetchedTimestamp(accountID hedera.AccountID, timestamp string) error {
	return s.dbClient.
		Where("code = ? and account_id = ?", lastFetchedTimestampCode, accountID.String()).
		Save(Status{
			Name:      "Last fetched timestamp",
			AccountID: accountID.String(),
			Code:      lastFetchedTimestampCode,
			Value:     timestamp,
		}).
		Error
}
