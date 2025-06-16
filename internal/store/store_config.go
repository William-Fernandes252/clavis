package store

// Common configuration for all store implementations
type StoreConfig struct {
	LoggingLevel      int // 0=DEBUG, 1=INFO, 2=WARNING, 3=ERROR
	NumVersionsToKeep int // Number of versions to keep for each key
}

// Get the logging level for the store
func (sc StoreConfig) GetLoggingLevel() int {
	return sc.LoggingLevel
}

// Get the number of versions to keep for each key
func (sc StoreConfig) GetNumVersions() int {
	return sc.NumVersionsToKeep
}
