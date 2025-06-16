package validation

import (
	"github.com/William-Fernandes252/clavis/internal/store"
)

const ValueMaxSize = 100 * 1024 * 1024 // 100MB

// Store wrapper to provide data validation functionality
type ValidatedStore struct {
	inner          store.Store
	keyValidator   func(string) error
	valueValidator func(string, []byte) error
}

// New creates a store with functional validation
func New(s store.Store, keyValidator func(string) error, valueValidator func(string, []byte) error) *ValidatedStore {
	return &ValidatedStore{
		inner:          s,
		keyValidator:   keyValidator,
		valueValidator: valueValidator,
	}
}

// NewWithDefaultValidators creates a store with default functional validation
func NewWithDefaultValidators(s store.Store) *ValidatedStore {
	keyValidator := ComposeKeyValidators(
		ValidateNonEmptyKey,
		ValidateKeyLength(1024),
	)

	valueValidator := ComposeValueValidators(
		ValidateValueSize(ValueMaxSize), // 100MB
	)

	return New(s, keyValidator, valueValidator)
}

func (f *ValidatedStore) Get(key string) ([]byte, bool, error) {
	if err := f.keyValidator(key); err != nil {
		return nil, false, err
	}
	return f.inner.Get(key)
}

func (f *ValidatedStore) Put(key string, value []byte) error {
	if err := f.keyValidator(key); err != nil {
		return err
	}
	if err := f.valueValidator(key, value); err != nil {
		return err
	}
	return f.inner.Put(key, value)
}

func (f *ValidatedStore) Delete(key string) error {
	if err := f.keyValidator(key); err != nil {
		return err
	}
	return f.inner.Delete(key)
}

func (f *ValidatedStore) Scan(prefix string) (map[string][]byte, error) {
	// Note: We could validate prefix here too if needed
	return f.inner.Scan(prefix)
}

func (f *ValidatedStore) Close() error {
	return f.inner.Close()
}

var _ store.Store = (*ValidatedStore)(nil)
