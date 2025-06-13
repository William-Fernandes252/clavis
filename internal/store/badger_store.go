package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// BadgerConfig holds the configuration options for BadgerDB
type BadgerConfig struct {
	StoreConfig        // Embedded struct with common config
	Path        string // BadgerDB-specific: database path
	SyncWrites  bool   // BadgerDB-specific: sync writes to disk
}

// DefaultBadgerConfig returns a BadgerConfig with sensible defaults
func DefaultBadgerConfig(path string) *BadgerConfig {
	return &BadgerConfig{
		StoreConfig: StoreConfig{
			LoggingLevel:      3, // ERROR level
			NumVersionsToKeep: 1,
		},
		Path:       path,
		SyncWrites: true,
	}
}

// ToBadgerOptions converts BadgerConfig to badger.Options
func (c *BadgerConfig) ToBadgerOptions() badger.Options {
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

type BadgerStore struct {
	db *badger.DB
}

func NewBadgerStore(config *BadgerConfig) (*BadgerStore, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	opts := config.ToBadgerOptions()

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

func NewBadgerStoreWithPath(path string) (*BadgerStore, error) {
	return NewBadgerStore(DefaultBadgerConfig(path))
}

// Close the BadgerDB instance
func (bs *BadgerStore) Close() error {
	return bs.db.Close()
}

// Get retrieves the value associated with the key
func (bs *BadgerStore) Get(key string) ([]byte, bool, error) {
	var value []byte
	var found bool

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Not an error, just not found
			}
			return err
		}

		found = true
		value, err = item.ValueCopy(nil)
		return err
	})

	return value, found, err
}

// Put stores the value associated with the key
func (bs *BadgerStore) Put(key string, value []byte) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// Delete removes the key and its associated value from the store
func (bs *BadgerStore) Delete(key string) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Scan retrieves all key-value pairs that start with the given prefix
func (bs *BadgerStore) Scan(prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	prefixBytes := []byte(prefix)

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefixBytes); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			// Check if key starts with prefix
			if !hasPrefix(key, prefixBytes) {
				break
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			result[string(key)] = value
		}
		return nil
	})

	return result, err
}

// hasPrefix checks if key starts with prefix
func hasPrefix(key, prefix []byte) bool {
	if len(prefix) > len(key) {
		return false
	}
	for i, b := range prefix {
		if key[i] != b {
			return false
		}
	}
	return true
}
