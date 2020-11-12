package repositories

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"

// AccountRepository Interface that all AccountRepository structs must implement
type StatusRepository interface {
	GetLastFetchedTimestamp() string
	UpdateLastFetchedTimestamp(timestamp string) (bool, error)
	GetStatus(code string) (*status.Status, error)
}
