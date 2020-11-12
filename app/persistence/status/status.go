package status

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// This table will contain information for latest status of the application
type Status struct {
	Name  string
	Code  string
	Value string
}

type StatusRepository struct {
	dbClient *gorm.DB
}

func NewStatusRepository(dbClient *gorm.DB) *StatusRepository {
	return &StatusRepository{
		dbClient: dbClient,
	}
}

func (s StatusRepository) GetLastFetchedTimestamp() string {
	lastFetchedStatus, err := s.GetStatus("LAST_FETCHED_TIMESTAMP")
	lastFetchedTimestamp := lastFetchedStatus.Value
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println(err)
		lastFetchedTimestamp = strconv.FormatInt(time.Now().Unix(), 10)
		s.dbClient.Create(Status{Name: "Last fetched timestamp", Code: "LAST_FETCHED_TIMESTAMP", Value: lastFetchedTimestamp})
	}
	return lastFetchedTimestamp
}

func (s StatusRepository) UpdateLastFetchedTimestamp(timestamp string) (bool, error) {
	failure := s.dbClient.Where("code = ?", "LAST_FETCHED_TIMESTAMP").
		Save(Status{Name: "Last fetched timestamp", Code: "LAST_FETCHED_TIMESTAMP", Value: timestamp})
	if failure.Error != nil {
		return false, failure.Error
	}
	return true, nil
}

func (s StatusRepository) GetStatus(code string) (*Status, error) {
	status := &Status{}
	failure := s.dbClient.Table("statuses").Where("code = ?", code).First(&status)
	if failure.Error != nil {
		return nil, failure.Error
	}
	return status, nil
}
