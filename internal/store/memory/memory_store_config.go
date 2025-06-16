package memory

import "github.com/William-Fernandes252/clavis/internal/store"

type MemoryStoreConfig struct {
	store.StoreConfig // Embedded struct with common config
	// No additional fields needed for MemoryStoreConfig
}

func DefaultConfig() *MemoryStoreConfig {
	return &MemoryStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level
			NumVersionsToKeep: 1,
		},
	}
}
