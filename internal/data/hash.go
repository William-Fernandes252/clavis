package data

import (
	"encoding/json"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

const hashDataType = Type("hash")

type Hash[T any] interface {
	Get(key string) T
	Set(key string, value T)
}

// hash represents a map of string keys to byte slices.
// Its the most low-level hash data type.
type hash map[string][]byte

// NewBytesHash creates a new HashData instance.
func NewBytesHash() Hash[[]byte] {
	return new(hash).Init()
}

// Init initializes the Hash instance.
func (h *hash) Init() *hash {
	*h = make(hash, 0)
	return h
}

// Set adds a key-value pair to the hash.
// If the hash is nil, it initializes a new Hash.
// If the key already exists, it updates the value.
func (h *hash) Set(key string, value []byte) {
	if h == nil {
		h = NewBytesHash().(*hash)
	}
	(*h)[key] = value
}

// Get retrieves the value associated with the key.
// If the hash is nil or the key does not exist, it returns nil.
// It returns a pointer to the value if it exists, or nil if it does not.
func (h *hash) Get(key string) []byte {
	if h == nil {
		return nil
	}
	value, exists := (*h)[key]
	if !exists {
		return nil
	}
	return value
}

var _ Hash[[]byte] = (*hash)(nil)

type DataHash[T any] struct {
	// The underlying hash data structure.
	inner Hash[[]byte]
	// The codec used to serialize/deserialize the data.
	codec Codec[T]
}

// NewDataHash creates a new DataHash instance that uses the provided Hash for managing keys and values.
func NewDataHash[T any](codec Codec[T]) *DataHash[T] {
	return &DataHash[T]{
		inner: NewBytesHash(),
		codec: codec,
	}
}

// Set stores a key with its value.
// The value is serialized using the codec before being stored.
func (dh *DataHash[T]) Set(key string, value T) {
	// Serialize the value using the codec
	serializedValue, _ := dh.codec.Serialize(value)

	// Store the serialized value in the inner hash
	dh.inner.Set(key, serializedValue)
}

// Get retrieves the value for a given key and deserializes it using the codec.
func (dh *DataHash[T]) Get(key string) T {
	var zero T

	// Get the serialized value from the inner hash
	serializedValue := dh.inner.Get(key)
	if serializedValue == nil {
		return zero // Key does not exist
	}

	// Deserialize the value using the codec
	value, err := dh.codec.Deserialize(serializedValue)
	if err != nil {
		return zero // Return zero value if deserialization fails
	}

	return value
}

// HashCodec represents a hash data type. It implements the Data interface for hash values.
type HashCodec[T any] struct {
	// The codec used to serialize/deserialize the hash data.
	codec Codec[T]
}

// NewHashCodec creates a new HashCodec instance that uses the provided Codec for managing hash data.
func NewHashCodec[T any](codec Codec[T]) *HashCodec[T] {
	return &HashCodec[T]{codec: codec}
}

// Type returns the type name ("hash")
func (h *HashCodec[T]) Type() Type {
	return hashDataType
}

// Serialize converts the hash value into a byte slice for storage
func (h *HashCodec[T]) Serialize(value Hash[T]) ([]byte, errors.Error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, NewDataError("serialization-failed", "Failed to serialize hash data", err)
	}
	return data, nil
}

// Deserialize reads the bytes back into the hash data type
// It returns an empty hash if the input is nil
func (h *HashCodec[T]) Deserialize(raw []byte) (Hash[T], errors.Error) {
	if raw == nil {
		return NewDataHash(h.codec), nil
	}
	var value DataHash[T]
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewDataError("deserialization-failed", "Failed to deserialize hash data", err)
	}
	return &value, nil
}

var _ Codec[Hash[any]] = (*HashCodec[any])(nil)
