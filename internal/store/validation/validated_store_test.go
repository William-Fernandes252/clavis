package validation

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
)

func TestValidatedStore(t *testing.T) {
	baseStore := createTestStore(t)
	defer baseStore.Close()

	store := NewWithDefaultValidators(baseStore)
	defer store.Close()

	t.Run("AcceptValidData", func(t *testing.T) {
		err := store.Put("valid-key", []byte("valid-value"))
		if err != nil {
			t.Errorf("Expected valid data to be accepted: %v", err)
		}
	})

	t.Run("RejectEmptyKey", func(t *testing.T) {
		err := store.Put("", []byte("value"))
		if err == nil {
			t.Error("Expected empty key to be rejected")
		}
	})

	t.Run("RejectLongKey", func(t *testing.T) {
		longKey := strings.Repeat("a", 2000) // Longer than 1024 limit
		err := store.Put(longKey, []byte("value"))
		if err == nil {
			t.Error("Expected long key to be rejected")
		}
	})

	t.Run("RejectLargeValue", func(t *testing.T) {
		largeValue := make([]byte, 200*1024*1024) // Larger than 100MB limit
		err := store.Put("key", largeValue)
		if err == nil {
			t.Error("Expected large value to be rejected")
		}
	})
}

func TestValidatedStore_CustomValidators(t *testing.T) {
	baseStore := createTestStore(t)
	defer baseStore.Close()

	// Custom key validator that only allows keys starting with "test:"
	keyValidator := ComposeKeyValidators(
		ValidateNonEmptyKey,
		func(key string) error {
			if !strings.HasPrefix(key, "test:") {
				return fmt.Errorf("key must start with 'test:'")
			}
			return nil
		},
	)

	// Custom value validator that only allows JSON-like values
	valueValidator := ComposeValueValidators(
		func(key string, value []byte) error {
			str := strings.TrimSpace(string(value))
			if !strings.HasPrefix(str, "{") || !strings.HasSuffix(str, "}") {
				return fmt.Errorf("value must be JSON-like")
			}
			return nil
		},
	)

	store := New(baseStore, keyValidator, valueValidator)
	defer store.Close()

	t.Run("AcceptValidFormat", func(t *testing.T) {
		err := store.Put("test:data", []byte(`{"key": "value"}`))
		if err != nil {
			t.Errorf("Expected valid format to be accepted: %v", err)
		}
	})

	t.Run("RejectInvalidKeyPrefix", func(t *testing.T) {
		err := store.Put("invalid:data", []byte(`{"key": "value"}`))
		if err == nil {
			t.Error("Expected invalid key prefix to be rejected")
		}
	})

	t.Run("RejectInvalidValueFormat", func(t *testing.T) {
		err := store.Put("test:data", []byte(`plain text`))
		if err == nil {
			t.Error("Expected invalid value format to be rejected")
		}
	})
}

func TestValidatedStore_Composition(t *testing.T) {
	baseStore := createTestStore(t)
	defer baseStore.Close()

	// Compose multiple key validators
	keyValidator := ComposeKeyValidators(
		ValidateNonEmptyKey,
		ValidateKeyLength(20),
		func(key string) error {
			if strings.Contains(key, "banned") {
				return fmt.Errorf("key cannot contain 'banned'")
			}
			return nil
		},
		func(key string) error {
			if !strings.Contains(key, ":") {
				return fmt.Errorf("key must contain ':'")
			}
			return nil
		},
	)

	// Compose multiple value validators
	valueValidator := ComposeValueValidators(
		ValidateValueSize(100),
		func(key string, value []byte) error {
			if strings.Contains(string(value), "forbidden") {
				return fmt.Errorf("value cannot contain 'forbidden'")
			}
			return nil
		},
	)

	store := New(baseStore, keyValidator, valueValidator)
	defer store.Close()

	t.Run("AllValidatorsPass", func(t *testing.T) {
		err := store.Put("valid:key", []byte("valid value"))
		if err != nil {
			t.Errorf("Expected all validators to pass: %v", err)
		}
	})

	t.Run("FirstValidatorFails", func(t *testing.T) {
		err := store.Put("", []byte("value"))
		if err == nil {
			t.Error("Expected first validator (empty key) to fail")
		}
	})

	t.Run("SecondValidatorFails", func(t *testing.T) {
		longKey := strings.Repeat("a", 25)
		err := store.Put(longKey, []byte("value"))
		if err == nil {
			t.Error("Expected second validator (key length) to fail")
		}
	})

	t.Run("ThirdValidatorFails", func(t *testing.T) {
		err := store.Put("banned:key", []byte("value"))
		if err == nil {
			t.Error("Expected third validator (banned word) to fail")
		}
	})

	t.Run("FourthValidatorFails", func(t *testing.T) {
		err := store.Put("nocolon", []byte("value"))
		if err == nil {
			t.Error("Expected fourth validator (colon requirement) to fail")
		}
	})

	t.Run("ValueValidatorFails", func(t *testing.T) {
		err := store.Put("valid:key", []byte("forbidden content"))
		if err == nil {
			t.Error("Expected value validator to fail")
		}
	})
}

func TestValidatedStore_DomainSpecific(t *testing.T) {
	baseStore := createTestStore(t)
	defer baseStore.Close()

	// Create a user management store with domain-specific validation
	userKeyValidator := ComposeKeyValidators(
		ValidateNonEmptyKey,
		func(key string) error {
			if !strings.HasPrefix(key, "user:") {
				return fmt.Errorf("user keys must start with 'user:'")
			}

			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				return fmt.Errorf("user key format must be 'user:id'")
			}

			userId := parts[1]
			if len(userId) < 3 {
				return fmt.Errorf("user ID must be at least 3 characters")
			}

			return nil
		},
	)

	userValueValidator := ComposeValueValidators(
		ValidateValueSize(1024), // 1KB max
		func(key string, value []byte) error {
			// Simple validation: must contain email field
			if !strings.Contains(string(value), "email") {
				return fmt.Errorf("user data must contain email field")
			}
			return nil
		},
	)

	userStore := New(baseStore, userKeyValidator, userValueValidator)
	defer userStore.Close()

	t.Run("ValidUser", func(t *testing.T) {
		err := userStore.Put("user:john123", []byte(`{"name": "John", "email": "john@example.com"}`))
		if err != nil {
			t.Errorf("Expected valid user to be accepted: %v", err)
		}
	})

	t.Run("InvalidKeyFormat", func(t *testing.T) {
		err := userStore.Put("john123", []byte(`{"name": "John", "email": "john@example.com"}`))
		if err == nil {
			t.Error("Expected invalid key format to be rejected")
		}
	})

	t.Run("ShortUserId", func(t *testing.T) {
		err := userStore.Put("user:jo", []byte(`{"name": "Jo", "email": "jo@example.com"}`))
		if err == nil {
			t.Error("Expected short user ID to be rejected")
		}
	})

	t.Run("MissingEmail", func(t *testing.T) {
		err := userStore.Put("user:alice123", []byte(`{"name": "Alice"}`))
		if err == nil {
			t.Error("Expected missing email to be rejected")
		}
	})
}

func TestValidatedStore_ErrorMessages(t *testing.T) {
	baseStore := createTestStore(t)
	defer baseStore.Close()

	// Create validators with specific error messages
	keyValidator := func(key string) error {
		if key == "" {
			return fmt.Errorf("EMPTY_KEY: key cannot be empty")
		}
		if len(key) > 10 {
			return fmt.Errorf("KEY_TOO_LONG: maximum 10 characters, got %d", len(key))
		}
		return nil
	}

	valueValidator := func(key string, value []byte) error {
		if len(value) > 50 {
			return fmt.Errorf("VALUE_TOO_LARGE: maximum 50 bytes, got %d", len(value))
		}
		return nil
	}

	store := New(baseStore, keyValidator, valueValidator)
	defer store.Close()

	t.Run("EmptyKeyError", func(t *testing.T) {
		err := store.Put("", []byte("value"))
		if err == nil {
			t.Error("Expected error")
		} else if !strings.Contains(err.Error(), "EMPTY_KEY") {
			t.Errorf("Expected EMPTY_KEY error, got: %v", err)
		}
	})

	t.Run("LongKeyError", func(t *testing.T) {
		err := store.Put("very-long-key", []byte("value"))
		if err == nil {
			t.Error("Expected error")
		} else if !strings.Contains(err.Error(), "KEY_TOO_LONG") {
			t.Errorf("Expected KEY_TOO_LONG error, got: %v", err)
		}
	})

	t.Run("LargeValueError", func(t *testing.T) {
		largeValue := make([]byte, 100)
		err := store.Put("key", largeValue)
		if err == nil {
			t.Error("Expected error")
		} else if !strings.Contains(err.Error(), "VALUE_TOO_LARGE") {
			t.Errorf("Expected VALUE_TOO_LARGE error, got: %v", err)
		}
	})
}

// Benchmark to ensure validation doesn't add significant overhead
func BenchmarkValidatedStore(b *testing.B) {
	// Simple validation
	keyValidator := ComposeKeyValidators(
		ValidateNonEmptyKey,
		ValidateKeyLength(100),
	)

	valueValidator := ComposeValueValidators(
		ValidateValueSize(1024),
	)

	validatedStore := createTestStoreWithKeyAndValueValidators(b, keyValidator, valueValidator)
	defer validatedStore.Close()

	key := "benchmark-key"
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validatedStore.Put(key, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func createTestStoreWithKeyAndValueValidators(t testing.TB, keyValidator func(string) error, valueValidator func(string, []byte) error) *ValidatedStore {
	baseStore := createTestStore(t)
	defer baseStore.Close()

	store := New(baseStore, keyValidator, valueValidator)
	if store == nil {
		t.Fatal("Failed to create validated store")
	}

	return store
}

func createTestStore(t testing.TB) *ValidatedStore {
	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up will be handled by individual tests
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &badger.BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level for quiet tests
			NumVersionsToKeep: 1,
		},
		Path:       tempDir,
		SyncWrites: false, // Faster for tests
	}

	store, err := badger.New(config)
	if err != nil {
		t.Fatal(err)
	}

	return NewWithDefaultValidators(store)
}
