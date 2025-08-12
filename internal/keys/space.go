package keys

import (
	"github.com/William-Fernandes252/clavis/internal/data"
	"github.com/William-Fernandes252/clavis/internal/errors"
)

// KeySpace defines the interface for the current keys managed by the system.
type KeySpace[T any] interface {
	// Set stores a key with its value.
	// If the key already exists, it will be overwritten.
	// Returns an error if the key is invalid, or if the value cannot be stored.
	Set(key Key, value T) errors.Error

	// Get retrieves the value for a given key.
	Get(key Key) (T, errors.Error)

	// Del removes the specified keys from the key space.
	// If a key does not exist, it is ignored.
	// The operation is atomic, meaning it will either succeed for all keys or fail without partial deletion.
	// Returns the number of keys that were removed.
	Del(keys ...Key) (int, errors.Error)

	// Keys retrieves all keys matching a given glob pattern.
	Keys(pattern string) ([]Key, errors.Error)

	// Scan retrieves a limited number of keys matching a given glob pattern.
	Scan(pattern string, count int) ([]Key, errors.Error)
}

// DataKeySpace is wrapper around KeySpace to access the key values as structured data.
type DataKeySpace[T any] struct {
	// The key space where keys are associated with their data.
	inner KeySpace[*Entry]

	// The codec used to serialize/deserialize the data.
	codec data.Codec[T]
}

// NewDataKeySpace creates a new DataKeySpace instance that uses the provided KeySpace for managing keys and values.
func NewDataKeySpace[T any](ks KeySpace[*Entry], codec data.Codec[T]) *DataKeySpace[T] {
	return &DataKeySpace[T]{
		inner: ks,
		codec: codec,
	}
}

// Set stores a key with its value.
// The value is serialized using the codec before being stored.
func (dks *DataKeySpace[T]) Set(key Key, value T) errors.Error {
	// Serialize the value using the codec
	serializedValue, err := dks.codec.Serialize(value)
	if err != nil {
		return err
	}

	// Create an entry with the serialized value
	entry := NewEntry(key, serializedValue, dks.codec.Type())

	// Store the entry in the inner key space
	return dks.inner.Set(key, entry)
}

// Get retrieves the value for a given key and deserializes it using the codec.
func (dks *DataKeySpace[T]) Get(key Key) (T, errors.Error) {
	// zero is the default value for type T, returned in case of error.
	var zero T

	// Get the entry from the inner key space
	entry, err := dks.inner.Get(key)
	if err != nil {
		return zero, err
	}

	// Deserialize the value using the codec
	value, deserializeErr := dks.codec.Deserialize(entry.Value())
	if deserializeErr != nil {
		return zero, deserializeErr
	}

	return value, nil
}

// Del removes the specified keys from the key space.
func (dks *DataKeySpace[T]) Del(keys ...Key) (int, errors.Error) {
	return dks.inner.Del(keys...)
}

// Keys retrieves all keys matching a given glob pattern.
func (dks *DataKeySpace[T]) Keys(pattern string) ([]Key, errors.Error) {
	return dks.inner.Keys(pattern)
}

// Scan retrieves a limited number of keys matching a given glob pattern.
func (dks *DataKeySpace[T]) Scan(pattern string, count int) ([]Key, errors.Error) {
	return dks.inner.Scan(pattern, count)
}

var _ KeySpace[any] = (*DataKeySpace[any])(nil)
