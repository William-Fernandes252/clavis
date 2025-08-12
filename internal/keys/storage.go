package keys

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"sync"

	"github.com/William-Fernandes252/clavis/internal/data"
	"github.com/William-Fernandes252/clavis/internal/errors"
)

const storageErrorType = "storage"

// NewStorageError creates a new storage error with the given code, message, and cause.
func NewStorageError(code, message string, cause error) *errors.BaseError {
	return &errors.BaseError{
		ErrorData: errors.NewErrorData(storageErrorType, code, message),
		Cause:     &cause,
	}
}

// Entry represents a key-value pair in the storage system.
// It includes the key, value, type of data, and expiration time.
type Entry struct {
	// Key is the unique identifier for the entry
	value []byte

	key []byte

	// Type is type of the data stored in the entry.
	datatype data.Type
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for Entry.
func (e Entry) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s:%s:%s", e.value, e.key, e.datatype)
	return b.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for Entry.
func (e *Entry) UnmarshalBinary(value []byte) error {
	var datatype data.Type
	var key []byte
	n, err := fmt.Sscanf(string(value), "%s:%s:%s", &value, &key, &datatype)
	if err != nil || n != 4 {
		return NewStorageError("deserialization-failed", "Failed to un-marshal entry", err)
	}
	e.datatype = datatype
	e.value = value
	e.key = key
	return nil
}

// NewEntry creates a new Entry with the given value, key, type and expiration time.
func NewEntry(key Key, value []byte, datatype data.Type) *Entry {
	return &Entry{
		key:      key.Bytes(),
		value:    value,
		datatype: datatype,
	}
}

// Value returns the value of the entry as a byte slice.
// It is the responsibility of the caller to ensure that the value is interpreted correctly based on its datatype.
func (e *Entry) Value() []byte {
	return e.value
}

// Key returns the key of the entry as a byte slice.
func (e *Entry) Key() Key {
	return Key(string(e.key))
}

// Type returns the type of the entry as a string.
// This should match the type used when retrieving the value from the entry.
// It is used to determine how to deserialize the value.
// For example, if the entry is a list, this should return "list".
func (e *Entry) Type() data.Type {
	return e.datatype
}

// Serialize converts the Entry into a byte slice for storage.
func (e *Entry) Serialize() ([]byte, errors.Error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(e); err != nil {
		return nil, NewStorageError("serialization-failed", "Failed to serialize entry", err)
	}
	return buf.Bytes(), nil
}

// Deserialize reads the bytes back into the Entry.
func (e *Entry) Deserialize(data []byte) (*Entry, errors.Error) {
	if len(data) == 0 {
		return nil, nil // No data to deserialize
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(e); err != nil {
		return nil, NewStorageError("deserialization-failed", "Failed to deserialize entry", err)
	}
	return e, nil
}

// Getter is an interface for retrieving values from a key-value store.
type Getter[T any] interface {
	// Get retrieves the value associated with the key. Returns the value, a boolean indicating if the key exists, and an error if any.
	Get(key Key) (T, bool, errors.Error)
}

// Putter is an interface for storing values in a key-value store.
type Putter[T any] interface {
	// Put stores the value associated with the key. Returns an error if any.
	Put(key Key, value T) errors.Error
}

// Deleter is an interface for deleting keys from a key-value store.
type Deleter interface {
	// Delete removes the key and its associated value from the store. Returns an error if any.
	Delete(key Key) errors.Error
}

// Scanner is an interface for scanning keys in a key-value store.
type Scanner[T any] interface {
	// Scan retrieves all key-value pairs that match the given glob pattern. Returns a map of key-value pairs and an error if any.
	Scan(pattern string) (map[Key]T, errors.Error)
}

// Storage is an interface that defines methods for a key-value store.
type Storage interface {
	io.Closer
	Getter[*Entry]
	Putter[*Entry]
	Deleter
	Scanner[*Entry]

	// Name returns the name of the store.
	Name() string
}

// entryKeySpace is a concrete implementation of the KeySpace that should be used as a base to interact with the keys store.
type entryKeySpace struct {
	mu      sync.RWMutex
	storage Storage // The underlying storage for key-value pairs
}

// NewEntryKeySpace creates a new KeySpace instance that uses the provided storage for managing keys and values.
func NewEntryKeySpace(storage Storage) KeySpace[*Entry] {
	return &entryKeySpace{
		mu:      sync.RWMutex{},
		storage: storage,
	}
}

// Set stores a key with its value.
func (ks *entryKeySpace) Set(key Key, value *Entry) errors.Error {
	if err := key.Validate(); err != nil {
		return NewInvalidKeyError(key, err)
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()

	if err := ks.storage.Put(key, value); err != nil {
		return NewKeyError("store-error", "Failed to store key-value pair", err).
			WithMetadata("key", key.String())
	}

	return nil
}

// Get retrieves the value for a given key.
func (ks *entryKeySpace) Get(key Key) (*Entry, errors.Error) {
	if err := key.Validate(); err != nil {
		return nil, NewInvalidKeyError(key, err)
	}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	entry, exists, err := ks.storage.Get(key)
	if err != nil {
		return nil, NewKeyError("store-error", "failed to retrieve key from store", err).
			WithMetadata("key", key.String())
	}
	if !exists {
		return nil, NewKeyNotFoundError(key, nil)
	}

	return entry, nil
}

// Del removes the specified keys from the key space.
func (ks *entryKeySpace) Del(keys ...Key) (int, errors.Error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	deletedCount := 0

	for _, key := range keys {
		if err := key.Validate(); err != nil {
			continue // Skip invalid keys
		}

		if err := ks.storage.Delete(key); err != nil {
			return deletedCount, NewKeyError("store-error", "failed to delete key from store", err).
				WithMetadata("key", key.String())
		}

		deletedCount++
	}

	return deletedCount, nil
}

// Keys retrieves all keys matching a given pattern.
func (ks *entryKeySpace) Keys(pattern string) ([]Key, errors.Error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keyMap, err := ks.storage.Scan(pattern)
	if err != nil {
		return nil, NewKeyError("scan-error", "failed to scan keys from store", err).
			WithMetadata("pattern", pattern)
	}

	var matchingKeys []Key
	for key := range keyMap {
		matchingKeys = append(matchingKeys, key)
	}

	return matchingKeys, nil
}

// Scan retrieves a limited number of keys matching a given pattern.
func (ks *entryKeySpace) Scan(pattern string, count int) ([]Key, errors.Error) {
	if count <= 0 {
		return nil, NewKeyError("invalid-count", "count must be positive", nil).WithMetadata("count", count)
	}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keyMap, err := ks.storage.Scan(pattern)
	if err != nil {
		return nil, NewKeyError("scan-error", "failed to scan keys from store", err).WithMetadata("pattern", pattern)
	}

	var matchingKeys []Key
	for key := range keyMap {
		matchingKeys = append(matchingKeys, key)
	}
	if len(matchingKeys) > count {
		matchingKeys = matchingKeys[:count]
	}

	return matchingKeys, nil
}

var _ KeySpace[*Entry] = (*entryKeySpace)(nil)
