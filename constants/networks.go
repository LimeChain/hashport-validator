package constants

var (
	// Both Maps will be initialized on starting of the application and parsing the Bridge config.
	// They are defined here as global constants for convenience.

	NetworksById = map[uint64]string{}

	NetworksByName = map[string]uint64{}
)
