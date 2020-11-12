package status

import (
	"errors"
	"github.com/hashgraph/hedera-sdk-go"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"time"
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

func (s StatusRepository) GetLastFetchedTimestamp(accountID hedera.AccountID) string {
	lastFetchedStatus := &Status{}
	failure := s.dbClient.
		Table("statuses").
		Where("code = ? and account_id = ?", "LAST_FETCHED_TIMESTAMP", accountID.String()).
		First(&lastFetchedStatus).Error

	if failure != nil && errors.Is(failure, gorm.ErrRecordNotFound) {
		log.Errorf("Could not get last fetched timestamp: [%s]\n", failure)
		now := time.Now()
		newLastFetchedTimestamp := strconv.FormatInt(now.Unix(), 10)
		log.Infof("Proceeding monitoring from current moment [%s] - [%s].\n", now.String(), now.Unix())
		s.dbClient.Create(Status{
			Name:      "Last fetched timestamp",
			AccountID: accountID.String(),
			Code:      "LAST_FETCHED_TIMESTAMP",
			Value:     newLastFetchedTimestamp,
		})
		return newLastFetchedTimestamp
	}
	return lastFetchedStatus.Value
}

func (s StatusRepository) UpdateLastFetchedTimestamp(accountID hedera.AccountID, timestamp string) error {
	return s.dbClient.
		Where("code = ? and account_id = ?", "LAST_FETCHED_TIMESTAMP", accountID.String()).
		Save(Status{
			Name:      "Last fetched timestamp",
			AccountID: accountID.String(),
			Code:      "LAST_FETCHED_TIMESTAMP",
			Value:     timestamp,
		}).
		Error
}
