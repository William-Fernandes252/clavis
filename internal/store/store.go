package store

import "io"

type Getter interface {
	// Get retrieves the value associated with the key. Returns the value, a boolean indicating if the key exists, and an error if any.
	Get(key string) ([]byte, bool, error)
}

type Putter interface {
	// Put stores the value associated with the key. Returns an error if any.
	Put(key string, value []byte) error
}

type Deleter interface {
	// Delete removes the key and its associated value from the store. Returns an error if any.
	Delete(key string) error
}

type Scanner interface {
	// Scan retrieves all key-value pairs that start with the given prefix. Returns a map of key-value pairs and an error if any.
	Scan(prefix string) (map[string][]byte, error)
}

// Store is an interface that defines methods for a key-value store.
type Store interface {
	io.Closer
	Getter
	Putter
	Deleter
	Scanner
}
