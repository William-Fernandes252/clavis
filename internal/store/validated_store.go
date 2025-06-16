package store

import "fmt"

const ValueMaxSize = 100 * 1024 * 1024 // 100MB

// Store wrapper to provide data validation functionality
type ValidatedStore struct {
	inner          Store
	keyValidator   func(string) error
	valueValidator func(string, []byte) error
}

func ValidateNonEmptyKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	return nil
}

func ValidateKeyLength(maxLength int) func(string) error {
	return func(key string) error {
		if len(key) > maxLength {
			return fmt.Errorf("key too long: maximum %d characters, got %d", maxLength, len(key))
		}
		return nil
	}
}

func ValidateValueSize(maxSize int64) func(string, []byte) error {
	return func(key string, value []byte) error {
		if int64(len(value)) > maxSize {
			return fmt.Errorf("value too large: maximum %d bytes, got %d", maxSize, len(value))
		}
		return nil
	}
}

// ComposeKeyValidators combines multiple key validation functions
func ComposeKeyValidators(validators ...func(string) error) func(string) error {
	return func(key string) error {
		for _, validator := range validators {
			if err := validator(key); err != nil {
				return err
			}
		}
		return nil
	}
}

// ComposeValueValidators combines multiple value validation functions
func ComposeValueValidators(validators ...func(string, []byte) error) func(string, []byte) error {
	return func(key string, value []byte) error {
		for _, validator := range validators {
			if err := validator(key, value); err != nil {
				return err
			}
		}
		return nil
	}
}

// NewValidatedStore creates a store with functional validation
func NewValidatedStore(s Store, keyValidator func(string) error, valueValidator func(string, []byte) error) *ValidatedStore {
	return &ValidatedStore{
		inner:          s,
		keyValidator:   keyValidator,
		valueValidator: valueValidator,
	}
}

// NewDefaultValidatedStore creates a store with default functional validation
func NewDefaultValidatedStore(s Store) *ValidatedStore {
	keyValidator := ComposeKeyValidators(
		ValidateNonEmptyKey,
		ValidateKeyLength(1024),
	)

	valueValidator := ComposeValueValidators(
		ValidateValueSize(ValueMaxSize), // 100MB
	)

	return NewValidatedStore(s, keyValidator, valueValidator)
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

var _ Store = (*ValidatedStore)(nil)
