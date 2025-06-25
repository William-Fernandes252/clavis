package validation

import (
	"os"
	"strings"
	"testing"

	"github.com/William-Fernandes252/clavis/internal/model/validation/validators"
	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
)

func TestValidatedStore(t *testing.T) {
	baseStore := createTestStoreBase(t)
	defer func() {
		if err := baseStore.Close(); err != nil {
			t.Logf("Failed to close base store: %v", err)
		}
	}()

	store := NewWithDefaultValidators(baseStore)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close validated store: %v", err)
		}
	}()

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
	baseStore := createTestStoreBase(t)
	defer func() {
		if err := baseStore.Close(); err != nil {
			t.Logf("Failed to close base store: %v", err)
		}
	}()

	// Custom key validator that only allows keys starting with "test:"
	keyValidator := NewStoreKeyValidator(
		NonEmptyKeyValidator(),
		validators.Custom(func(key string) bool {
			return strings.HasPrefix(key, "test:")
		}, "key must start with 'test:'").WithName("test-prefix"),
	)

	// Custom value validator that only allows JSON-like values
	valueValidator := NewStoreValueValidator(
		ValueContentValidator(func(value []byte) bool {
			str := strings.TrimSpace(string(value))
			return strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")
		}, "value must be JSON-like"),
	)

	store := New(baseStore, keyValidator, valueValidator)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

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
	baseStore := createTestStoreBase(t)
	defer func() {
		if err := baseStore.Close(); err != nil {
			t.Logf("Failed to close base store: %v", err)
		}
	}()

	// Compose multiple key validators
	keyValidator := NewStoreKeyValidator(
		NonEmptyKeyValidator(),
		KeyLengthValidator(20),
		validators.Custom(func(key string) bool {
			return !strings.Contains(key, "banned")
		}, "key cannot contain 'banned'").WithName("no-banned"),
		validators.Custom(func(key string) bool {
			return strings.Contains(key, ":")
		}, "key must contain ':'").WithName("requires-colon"),
	)

	// Compose multiple value validators
	valueValidator := NewStoreValueValidator(
		ValueSizeValidator(100),
		ValueContentValidator(func(value []byte) bool {
			return !strings.Contains(string(value), "forbidden")
		}, "value cannot contain 'forbidden'"),
	)

	store := New(baseStore, keyValidator, valueValidator)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

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
	baseStore := createTestStoreBase(t)
	defer func() {
		if err := baseStore.Close(); err != nil {
			t.Logf("Failed to close base store: %v", err)
		}
	}()

	// Create a user management store with domain-specific validation
	userKeyValidator := NewStoreKeyValidator(
		NonEmptyKeyValidator(),
		validators.Custom(func(key string) bool {
			if !strings.HasPrefix(key, "user:") {
				return false
			}
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				return false
			}
			userId := parts[1]
			return len(userId) >= 3
		}, "invalid user key format").WithName("user-key-format"),
	)

	userValueValidator := NewStoreValueValidator(
		ValueSizeValidator(1024), // 1KB max
		ValueContentValidator(func(value []byte) bool {
			// Simple validation: must contain email field
			return strings.Contains(string(value), "email")
		}, "user data must contain email field"),
	)

	userStore := New(baseStore, userKeyValidator, userValueValidator)
	defer func() {
		if err := userStore.Close(); err != nil {
			t.Logf("Failed to close user store: %v", err)
		}
	}()

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
	baseStore := createTestStoreBase(t)
	defer func() {
		if err := baseStore.Close(); err != nil {
			t.Logf("Failed to close base store: %v", err)
		}
	}()

	// Create validators with specific error handling
	keyValidator := NewStoreKeyValidator(
		NonEmptyKeyValidator(),
		KeyLengthValidator(10),
	)

	valueValidator := NewStoreValueValidator(
		ValueSizeValidator(50),
	)

	store := New(baseStore, keyValidator, valueValidator)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

	t.Run("EmptyKeyError", func(t *testing.T) {
		err := store.Put("", []byte("value"))
		if err == nil {
			t.Error("Expected empty key to be rejected")
		} else if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Expected error message to mention 'empty', got: %v", err.Error())
		}
	})

	t.Run("LongKeyError", func(t *testing.T) {
		err := store.Put("very-long-key", []byte("value"))
		if err == nil {
			t.Error("Expected long key to be rejected")
		} else if !strings.Contains(err.Error(), "characters") {
			t.Errorf("Expected error message to mention 'characters', got: %v", err.Error())
		}
	})

	t.Run("LargeValueError", func(t *testing.T) {
		largeValue := make([]byte, 100)
		err := store.Put("key", largeValue)
		if err == nil {
			t.Error("Expected large value to be rejected")
		} else if !strings.Contains(err.Error(), "large") {
			t.Errorf("Expected error message to mention 'large', got: %v", err.Error())
		}
	})
}

// Benchmark to ensure validation doesn't add significant overhead
func BenchmarkValidatedStore(b *testing.B) {
	// Simple validation
	keyValidator := NewStoreKeyValidator(
		NonEmptyKeyValidator(),
		KeyLengthValidator(100),
	)

	valueValidator := NewStoreValueValidator(
		ValueSizeValidator(1024),
	)

	baseStore := createTestStoreBase(b)
	defer func() {
		if err := baseStore.Close(); err != nil {
			b.Logf("Failed to close base store: %v", err)
		}
	}()

	validatedStore := New(baseStore, keyValidator, valueValidator)
	defer func() {
		if err := validatedStore.Close(); err != nil {
			b.Logf("Failed to close validated store: %v", err)
		}
	}()

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

func createTestStore(t testing.TB) store.Store {
	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up will be handled by individual tests
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	})

	config := &badger.BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level for quiet tests
			NumVersionsToKeep: 1,
		},
		Path:       tempDir,
		SyncWrites: false, // Faster for tests
	}

	baseStore, err := badger.New(config)
	if err != nil {
		t.Fatal(err)
	}

	return baseStore
}

func createTestStoreBase(t testing.TB) store.Store {
	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up will be handled by individual tests
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	})

	config := &badger.BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level for quiet tests
			NumVersionsToKeep: 1,
		},
		Path:       tempDir,
		SyncWrites: false, // Faster for tests
	}

	baseStore, err := badger.New(config)
	if err != nil {
		t.Fatal(err)
	}

	return baseStore
}
