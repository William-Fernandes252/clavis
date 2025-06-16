package memory

import (
	"fmt"
	"strings"
	"sync"
)

// In-memory store that uses a map to manage key-value pairs.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func New(config *MemoryStoreConfig) (*MemoryStore, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &MemoryStore{
		data: make(map[string][]byte),
	}, nil
}

func NewWithDefaults() (*MemoryStore, error) {
	return New(DefaultConfig())
}

func (ms *MemoryStore) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Clear the map to help with garbage collection
	ms.data = nil
	return nil
}

// Get the value associated with the key
func (ms *MemoryStore) Get(key string) ([]byte, bool, error) {
	if key == "" {
		return nil, false, fmt.Errorf("key cannot be empty")
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.data == nil {
		return nil, false, fmt.Errorf("store is closed")
	}

	value, found := ms.data[key]
	if !found {
		return nil, false, nil
	}

	// Return a copy to prevent external modification of internal data
	result := make([]byte, len(value))
	copy(result, value)
	return result, true, nil
}

// Store the value associated with the key
func (ms *MemoryStore) Put(key string, value []byte) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.data == nil {
		return fmt.Errorf("store is closed")
	}

	// Store a copy to prevent external modification of internal data
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	ms.data[key] = valueCopy
	return nil
}

// Remove the key and its associated value from the store
func (ms *MemoryStore) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.data == nil {
		return fmt.Errorf("store is closed")
	}

	delete(ms.data, key)
	return nil
}

// Retrieve all key-value pairs that start with the given prefix
func (ms *MemoryStore) Scan(prefix string) (map[string][]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.data == nil {
		return nil, fmt.Errorf("store is closed")
	}

	result := make(map[string][]byte)

	for key, value := range ms.data {
		if strings.HasPrefix(key, prefix) {
			// Return a copy to prevent external modification of internal data
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			result[key] = valueCopy
		}
	}

	return result, nil
}
