package repositories

import "github.com/limechain/hedera-eth-bridge-validator/app/process/model/timestamp"

type StatusRepository interface {
	GetLastFetchedTimestamp(entityID string) (*timestamp.Timestamp, error)
	UpdateLastFetchedTimestamp(entityID string, timestamp *timestamp.Timestamp) error
	CreateTimestamp(entityID string, timestamp *timestamp.Timestamp) error
}
