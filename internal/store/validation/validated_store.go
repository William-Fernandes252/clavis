package validation

import (
	"github.com/William-Fernandes252/clavis/internal/store"
)

const ValueMaxSize = 100 * 1024 * 1024 // 100MB

// Store wrapper to provide data validation functionality
type ValidatedStore struct {
	inner          store.Store
	keyValidator   *StoreKeyValidator
	valueValidator *StoreValueValidator
}

// New creates a store with validation using the new validation framework
func New(s store.Store, keyValidator *StoreKeyValidator, valueValidator *StoreValueValidator) *ValidatedStore {
	return &ValidatedStore{
		inner:          s,
		keyValidator:   keyValidator,
		valueValidator: valueValidator,
	}
}

// NewWithDefaultValidators creates a store with default validation
func NewWithDefaultValidators(s store.Store) *ValidatedStore {
	keyValidator := NewStoreKeyValidator(DefaultKeyValidators()...)
	valueValidator := NewStoreValueValidator(DefaultValueValidators()...)

	return New(s, keyValidator, valueValidator)
}

func (f *ValidatedStore) Get(key string) ([]byte, bool, error) {
	if err := f.keyValidator.Validate(key); err != nil {
		return nil, false, err
	}
	return f.inner.Get(key)
}

func (f *ValidatedStore) Put(key string, value []byte) error {
	if err := f.keyValidator.Validate(key); err != nil {
		return err
	}
	if err := f.valueValidator.Validate(key, value); err != nil {
		return err
	}
	return f.inner.Put(key, value)
}

func (f *ValidatedStore) Delete(key string) error {
	if err := f.keyValidator.Validate(key); err != nil {
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
