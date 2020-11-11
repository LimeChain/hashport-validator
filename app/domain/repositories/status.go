package repositories

import (
	"github.com/hashgraph/hedera-sdk-go"
)

// AccountRepository Interface that all AccountRepository structs must implement
type StatusRepository interface {
	GetLastFetchedTimestamp(accountID hedera.AccountID) (string, error)
	UpdateLastFetchedTimestamp(accountID hedera.AccountID, timestamp string) error
	CreateTimestamp(accountID hedera.AccountID, timestamp string) error
}
