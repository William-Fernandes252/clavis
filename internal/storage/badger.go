package storage

import (
	"fmt"
	"log/slog"

	"github.com/William-Fernandes252/clavis/internal/errors"
	"github.com/William-Fernandes252/clavis/internal/keys"
	"github.com/dgraph-io/badger/v4"
)

// BadgerLogger is a wrapper around slog.Logger that implements the badger.Logger interface
// This allows BadgerDB to use slog for logging, providing a consistent logging interface
type BadgerLogger struct {
	*slog.Logger
}

// NewBadgerLogger creates a new BadgerLogger with the provided slog.Logger
func NewBadgerLogger(logger *slog.Logger) *BadgerLogger {
	return &BadgerLogger{Logger: logger}
}

// Errorf implements badger.Logger.
func (l *BadgerLogger) Errorf(message string, args ...any) {
	l.Logger.Error(fmt.Sprintf(message, args...))
}

// Infof implements badger.Logger.
func (l *BadgerLogger) Infof(message string, args ...any) {
	l.Logger.Info(fmt.Sprintf(message, args...))
}

// Warningf implements badger.Logger.
func (l *BadgerLogger) Warningf(message string, args ...any) {
	l.Logger.Warn(fmt.Sprintf(message, args...))
}

// Debugf implements the badger.Logger interface for BadgerLogger
func (l *BadgerLogger) Debugf(message string, args ...any) {
	l.Logger.Debug(fmt.Sprintf(message, args...))
}

// BadgerLoggingLevel represents the logging level for BadgerDB
type BadgerLoggingLevel int

const (
	BadgerDebug   BadgerLoggingLevel = iota // Debug logging level
	BadgerInfo                              // Info logging level
	BadgerWarning                           // Warning logging level
	BadgerError                             // Error logging level
)

// BadgerStorageConfig holds the configuration options for BadgerDB storage
type BadgerStorageConfig struct {
	LoggingLevel      BadgerLoggingLevel // 0=DEBUG, 1=INFO, 2=WARNING, 3=ERROR
	NumVersionsToKeep int                // Number of versions to keep for each key
	Path              string             // BadgerDB-specific: database path
	SyncWrites        bool               // BadgerDB-specific: sync writes to disk
}

// NewBadgerConfigWithDefaults returns a BadgerStorageConfig with sensible defaults
func NewBadgerConfigWithDefaults(path string) *BadgerStorageConfig {
	return &BadgerStorageConfig{
		LoggingLevel:      BadgerError, // Default to ERROR level
		NumVersionsToKeep: 1,
		Path:              path,
		SyncWrites:        true,
	}
}

// toBadgerOptions converts BadgerStorageConfig to badger.Options
func (c *BadgerStorageConfig) toBadgerOptions() badger.Options {
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

// BadgerKeyStorage is an implementation of keys.Storage that uses BadgerDB for persistent storage using the file system
type BadgerKeyStorage struct {
	db     *badger.DB
	config BadgerStorageConfig
	logger *slog.Logger
}

// NewBadgerStorage creates a new BadgerKeyStorage instance with the provided configuration and logger
// It initializes the BadgerDB database with the specified options and logger
// Returns an error if the database cannot be opened
func NewBadgerStorage(config BadgerStorageConfig, logger *slog.Logger) (*BadgerKeyStorage, errors.Error) {
	opts := config.toBadgerOptions().WithLogger(NewBadgerLogger(logger))

	db, err := badger.Open(opts)
	if err != nil {
		return nil, NewStorageError("badger-open-failed", "failed to open BadgerDB", err)
	}

	return &BadgerKeyStorage{db: db, config: config, logger: logger}, nil
}

// Close implements keys.Storage.
func (b *BadgerKeyStorage) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// Delete implements keys.Storage.
func (b *BadgerKeyStorage) Delete(key keys.Key) errors.Error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key.Bytes())
	})
	if err != nil {
		return NewStorageError("badger-delete-failed", "failed to delete key from BadgerDB", err).
			WithMetadata("key", key.String())
	}
	return nil
}

// Get implements keys.Storage.
func (b *BadgerKeyStorage) Get(key keys.Key) (*keys.Entry, bool, errors.Error) {
	var entryData []byte
	var found bool

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key.Bytes())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Not an error, just not found
			}
			return err
		}

		found = true
		entryData, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, false, NewStorageError("badger-get-failed", "failed to get key from BadgerDB", err).WithMetadata("key", key.String())
	}

	if !found {
		return nil, false, nil
	}

	// Deserialize the entry
	entry := &keys.Entry{}
	deserializedEntry, storageErr := entry.Deserialize(entryData)
	if storageErr != nil {
		return nil, false, storageErr
	}

	return deserializedEntry, true, nil
}

// Name implements keys.Storage.
func (b *BadgerKeyStorage) Name() string {
	return "badger"
}

// Put implements keys.Storage.
func (b *BadgerKeyStorage) Put(key keys.Key, value *keys.Entry) errors.Error {
	// Serialize the entry
	entryData, err := value.Serialize()
	if err != nil {
		return err
	}

	dbErr := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key.Bytes(), entryData)
	})
	if dbErr != nil {
		return NewStorageError("badger-put-failed", "failed to put key in BadgerDB", dbErr).
			WithMetadata("key", key.String())
	}
	return nil
}

// Scan implements keys.Storage.
func (b *BadgerKeyStorage) Scan(pattern string) (result map[keys.Key]*keys.Entry, err errors.Error) {
	result = make(map[keys.Key]*keys.Entry)
	prefixBytes := []byte(pattern)

	dbErr := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefixBytes); it.Valid(); it.Next() {
			item := it.Item()
			keyBytes := item.Key()

			// Check if key starts with pattern (simple prefix matching)
			if !hasPrefix(keyBytes, prefixBytes) {
				break
			}

			entryData, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			// Deserialize the entry
			entry := &keys.Entry{}
			deserializedEntry, storageErr := entry.Deserialize(entryData)
			if storageErr != nil {
				return storageErr
			}

			key := keys.Key(string(keyBytes))
			result[key] = deserializedEntry
		}
		return nil
	})

	if dbErr != nil {
		err = NewStorageError("badger-scan-failed", "failed to scan keys from BadgerDB", dbErr).
			WithMetadata("pattern", pattern)
	}

	return
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

var _ keys.Storage = (*BadgerKeyStorage)(nil)
