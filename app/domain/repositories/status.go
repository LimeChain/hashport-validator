package repositories

type StatusRepository interface {
	GetLastFetchedTimestamp(entityID string) (int64, error)
	UpdateLastFetchedTimestamp(entityID string, timestamp int64) error
	CreateTimestamp(entityID string, timestamp int64) error
}
