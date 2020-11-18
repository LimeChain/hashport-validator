package repositories

// AccountRepository Interface that all AccountRepository structs must implement
type StatusRepository interface {
	GetLastFetchedTimestamp(entityID string) (string, error)
	UpdateLastFetchedTimestamp(entityID string, timestamp string) error
	CreateTimestamp(entityID string, timestamp string) error
}
