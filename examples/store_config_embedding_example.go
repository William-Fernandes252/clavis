//go:build !codeanalysis

package examples

import (
	"fmt"
	"log"

	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
)

// Function that works with any store config that embeds StoreConfig
func validateCommonConfig(config *store.StoreConfig) error {
	if config.GetLoggingLevel() < 0 || config.GetLoggingLevel() > 3 {
		return fmt.Errorf("invalid logging level")
	}
	if config.GetNumVersions() < 1 {
		return fmt.Errorf("invalid number of versions")
	}
	return nil
}

func store_config_embedding_example() {
	// Creating BadgerConfig with embedded StoreConfig fields
	config := &badger.BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      2, // WARNING
			NumVersionsToKeep: 5,
		},
		Path:       "/tmp/clavis",
		SyncWrites: true,
	}

	// Can access embedded fields directly
	fmt.Printf("Logging Level: %d\n", config.LoggingLevel)
	fmt.Printf("Versions: %d\n", config.NumVersionsToKeep)
	fmt.Printf("Path: %s\n", config.Path)
	fmt.Printf("Sync Writes: %t\n", config.SyncWrites)

	// Could validate using a common interface
	if err := validateCommonConfig(&config.StoreConfig); err != nil {
		log.Printf("Config validation failed: %v", err)
	}
}
