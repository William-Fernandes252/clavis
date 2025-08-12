package storage

import (
	"strings"
	"sync"

	"github.com/William-Fernandes252/clavis/internal/errors"
	"github.com/William-Fernandes252/clavis/internal/keys"
)

// MemoryKeyStorage is an in-memory storage that uses a hash to manage key-value pairs
type MemoryKeyStorage struct {
	mu   sync.RWMutex
	data map[string]*keys.Entry
}

// NewMemoryKeyStorage creates a new instance of MemoryKeyStorage
func NewMemoryKeyStorage() *MemoryKeyStorage {
	return &MemoryKeyStorage{
		data: make(map[string]*keys.Entry),
	}
}

// Close implements keys.Storage.
func (m *MemoryKeyStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear the map to help with garbage collection
	m.data = nil
	return nil
}

// Delete implements keys.Storage.
func (m *MemoryKeyStorage) Delete(key keys.Key) errors.Error {
	if err := key.Validate(); err != nil {
		return NewStorageError("invalid_key", "key validation failed", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		return NewStorageError("store_closed", "store is closed", nil)
	}

	delete(m.data, key.String())
	return nil
}

// Get implements keys.Storage.
func (m *MemoryKeyStorage) Get(key keys.Key) (*keys.Entry, bool, errors.Error) {
	if err := key.Validate(); err != nil {
		return nil, false, NewStorageError("invalid_key", "key validation failed", err)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		return nil, false, NewStorageError("store_closed", "store is closed", nil)
	}

	entry, found := m.data[key.String()]
	if !found {
		return nil, false, nil
	}

	return entry, true, nil
}

// Name implements keys.Storage.
func (m *MemoryKeyStorage) Name() string {
	return "memory"
}

// Put implements keys.Storage.
func (m *MemoryKeyStorage) Put(key keys.Key, value *keys.Entry) errors.Error {
	if err := key.Validate(); err != nil {
		return NewStorageError("invalid_key", "key validation failed", err)
	}

	if value == nil {
		return NewStorageError("invalid_value", "value cannot be nil", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		return NewStorageError("store_closed", "store is closed", nil)
	}

	// Store the entry directly (no need to copy since Entry is already a value type)
	m.data[key.String()] = value
	return nil
}

// Scan implements keys.Storage.
func (m *MemoryKeyStorage) Scan(pattern string) (map[keys.Key]*keys.Entry, errors.Error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		return nil, NewStorageError("store_closed", "store is closed", nil)
	}

	result := make(map[keys.Key]*keys.Entry)

	for keyStr, entry := range m.data {
		// Simple prefix matching (similar to memory_store.go)
		if len(pattern) == 0 || strings.HasPrefix(keyStr, pattern) {
			key := keys.Key(keyStr)
			result[key] = entry
		}
	}

	return result, nil
}

var _ keys.Storage = (*MemoryKeyStorage)(nil)
