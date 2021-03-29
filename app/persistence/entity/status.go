package entity

// This table will contain information for latest status of the application
type Status struct {
	Name      string
	EntityID  string
	Code      string
	Timestamp int64
}
