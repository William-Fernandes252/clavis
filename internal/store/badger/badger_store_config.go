package badger

import (
	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/dgraph-io/badger/v4"
)

// BadgerStoreConfig holds the configuration options for BadgerDB
type BadgerStoreConfig struct {
	store.StoreConfig        // Embedded struct with common config
	Path              string // BadgerDB-specific: database path
	SyncWrites        bool   // BadgerDB-specific: sync writes to disk
}

// DefaultConfig returns a BadgerConfig with sensible defaults
func DefaultConfig(path string) *BadgerStoreConfig {
	return &BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level
			NumVersionsToKeep: 1,
		},
		Path:       path,
		SyncWrites: true,
	}
}

// ToBadgerOptions converts BadgerConfig to badger.Options
func (c *BadgerStoreConfig) ToBadgerOptions() badger.Options {
	opts := badger.DefaultOptions(c.Path).
		WithSyncWrites(c.SyncWrites).
		WithNumVersionsToKeep(c.NumVersionsToKeep)

	switch c.LoggingLevel {
	case 0:
		opts = opts.WithLoggingLevel(badger.DEBUG)
	case 1:
		opts = opts.WithLoggingLevel(badger.INFO)
	case 2:
		opts = opts.WithLoggingLevel(badger.WARNING)
	case 3:
		opts = opts.WithLoggingLevel(badger.ERROR)
	default:
		opts = opts.WithLoggingLevel(badger.ERROR)
	}

	return opts
}
