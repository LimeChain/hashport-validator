package repositories

type StatusRepository interface {
	GetLastFetchedTimestamp(entityID string) (string, error)
	UpdateLastFetchedTimestamp(entityID string, timestamp string) error
	CreateTimestamp(entityID string, timestamp string) error
}
