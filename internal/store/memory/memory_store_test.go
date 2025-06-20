package memory

import (
	"fmt"
	"testing"

	"github.com/William-Fernandes252/clavis/internal/store"
)

func TestMemoryStore_Configuration(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		// Test the backward-compatible constructor
		store, err := NewWithDefaults()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := store.Close(); err != nil {
				t.Logf("Failed to close store: %v", err)
			}
		}()

		// Test basic operations
		testKey := "test-key"
		testValue := []byte("test-value")

		err = store.Put(testKey, testValue)
		if err != nil {
			t.Errorf("Put failed: %v", err)
		}

		value, found, err := store.Get(testKey)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Key not found")
		}
		if string(value) != string(testValue) {
			t.Errorf("Expected %s, got %s", testValue, value)
		}
	})

	t.Run("CustomConfiguration", func(t *testing.T) {
		// Test the dependency injection constructor with custom config
		config := &MemoryStoreConfig{
			StoreConfig: store.StoreConfig{
				LoggingLevel:      1, // INFO level for tests
				NumVersionsToKeep: 3, // Keep more versions for tests
			},
		}

		store, err := New(config)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := store.Close(); err != nil {
				t.Logf("Failed to close store: %v", err)
			}
		}()

		// Test basic operations
		testKey := "custom-key"
		testValue := []byte("custom-value")

		err = store.Put(testKey, testValue)
		if err != nil {
			t.Errorf("Put failed: %v", err)
		}

		value, found, err := store.Get(testKey)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Key not found")
		}
		if string(value) != string(testValue) {
			t.Errorf("Expected %s, got %s", testValue, value)
		}
	})

	t.Run("NilConfigurationError", func(t *testing.T) {
		// Test that nil configuration returns an error
		_, err := New(nil)
		if err == nil {
			t.Error("Expected error for nil configuration")
		}
		if err.Error() != "config cannot be nil" {
			t.Errorf("Expected 'config cannot be nil', got '%s'", err.Error())
		}
	})
}

func TestMemoryStore_Get(t *testing.T) {
	store := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

	t.Run("GetExistingKey", func(t *testing.T) {
		key := "existing-key"
		expectedValue := []byte("existing-value")

		// First put the key
		err := store.Put(key, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		// Then get it
		value, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if string(value) != string(expectedValue) {
			t.Errorf("Expected %s, got %s", expectedValue, value)
		}
	})

	t.Run("GetNonExistentKey", func(t *testing.T) {
		key := "non-existent-key"

		value, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if found {
			t.Error("Expected key to not be found")
		}
		if value != nil {
			t.Errorf("Expected nil value, got %v", value)
		}
	})

	t.Run("GetEmptyKey", func(t *testing.T) {
		key := ""

		value, found, err := store.Get(key)
		if err == nil {
			t.Error("Expected error when getting with empty key")
		}
		if found {
			t.Error("Expected key to not be found")
		}
		if value != nil {
			t.Errorf("Expected nil value, got %v", value)
		}
	})

	t.Run("GetWithSpecialCharacters", func(t *testing.T) {
		specialKeys := []string{
			"key:with:colons",
			"key/with/slashes",
			"key with spaces",
			"key-with-dashes",
			"key_with_underscores",
			"key.with.dots",
			"key@with@at",
			"unicode-key-??",
		}

		for _, key := range specialKeys {
			expectedValue := []byte("value-for-" + key)

			err := store.Put(key, expectedValue)
			if err != nil {
				t.Errorf("Put failed for key '%s': %v", key, err)
				continue
			}

			value, found, err := store.Get(key)
			if err != nil {
				t.Errorf("Get failed for key '%s': %v", key, err)
				continue
			}
			if !found {
				t.Errorf("Expected key '%s' to be found", key)
				continue
			}
			if string(value) != string(expectedValue) {
				t.Errorf("For key '%s': Expected %s, got %s", key, expectedValue, value)
			}
		}
	})

	t.Run("GetEmptyValue", func(t *testing.T) {
		key := "empty-value-key"
		expectedValue := []byte{}

		err := store.Put(key, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		value, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if len(value) != 0 {
			t.Errorf("Expected empty value, got %v", value)
		}
	})

	t.Run("GetAfterClose", func(t *testing.T) {
		tempStore := createTestStore(t)
		key := "test-key"
		value := []byte("test-value")

		err := tempStore.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}

		// Close the store
		if err := tempStore.Close(); err != nil {
			t.Logf("Failed to close temp store: %v", err)
		}

		// Try to get after close
		_, found, err := tempStore.Get(key)
		if err == nil {
			t.Error("Expected error when getting from closed store")
		}
		if found {
			t.Error("Expected key to not be found in closed store")
		}
	})
}

func TestMemoryStore_Put(t *testing.T) {
	store := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

	t.Run("PutBasic", func(t *testing.T) {
		key := "basic-key"
		value := []byte("basic-value")

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put failed: %v", err)
		}

		// Verify the value was stored
		storedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if string(storedValue) != string(value) {
			t.Errorf("Expected %s, got %s", value, storedValue)
		}
	})

	t.Run("PutOverwrite", func(t *testing.T) {
		key := "overwrite-key"
		originalValue := []byte("original-value")
		newValue := []byte("new-value")

		// Put original value
		err := store.Put(key, originalValue)
		if err != nil {
			t.Fatal(err)
		}

		// Overwrite with new value
		err = store.Put(key, newValue)
		if err != nil {
			t.Errorf("Put overwrite failed: %v", err)
		}

		// Verify the new value was stored
		storedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if string(storedValue) != string(newValue) {
			t.Errorf("Expected %s, got %s", newValue, storedValue)
		}
	})

	t.Run("PutEmptyKey", func(t *testing.T) {
		key := ""
		value := []byte("value-for-empty-key")

		err := store.Put(key, value)
		if err == nil {
			t.Error("Expected error when putting with empty key")
		}
	})

	t.Run("PutEmptyValue", func(t *testing.T) {
		key := "empty-value-key"
		value := []byte{}

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put failed for empty value: %v", err)
		}

		// Verify the empty value was stored
		storedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if len(storedValue) != 0 {
			t.Errorf("Expected empty value, got %v", storedValue)
		}
	})

	t.Run("PutNilValue", func(t *testing.T) {
		key := "nil-value-key"
		var value []byte = nil

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put failed for nil value: %v", err)
		}

		// Verify the nil value was stored as empty
		storedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if len(storedValue) != 0 {
			t.Errorf("Expected empty value, got %v", storedValue)
		}
	})

	t.Run("PutLargeValue", func(t *testing.T) {
		key := "large-value-key"
		// Create a 1MB value
		value := make([]byte, 1024*1024)
		for i := range value {
			value[i] = byte(i % 256)
		}

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put failed for large value: %v", err)
		}

		// Verify the large value was stored correctly
		storedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if len(storedValue) != len(value) {
			t.Errorf("Expected value length %d, got %d", len(value), len(storedValue))
		}
		for i := range value {
			if storedValue[i] != value[i] {
				t.Errorf("Value mismatch at index %d: expected %d, got %d", i, value[i], storedValue[i])
				break
			}
		}
	})

	t.Run("PutAfterClose", func(t *testing.T) {
		tempStore := createTestStore(t)
		key := "test-key"
		value := []byte("test-value")

		// Close the store
		if err := tempStore.Close(); err != nil {
			t.Logf("Failed to close temp store: %v", err)
		}

		// Try to put after close
		err := tempStore.Put(key, value)
		if err == nil {
			t.Error("Expected error when putting to closed store")
		}
	})

	t.Run("PutDataIsolation", func(t *testing.T) {
		// Test that modifying the original slice doesn't affect stored data
		key := "isolation-key"
		originalValue := []byte("original")

		err := store.Put(key, originalValue)
		if err != nil {
			t.Fatal(err)
		}

		// Modify the original slice
		originalValue[0] = 'X'

		// Get the stored value and verify it's unchanged
		storedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if string(storedValue) != "original" {
			t.Errorf("Expected 'original', got %s", storedValue)
		}

		// Also test that modifying the returned value doesn't affect stored data
		storedValue[0] = 'Y'

		// Get again and verify it's still unchanged
		storedValue2, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if string(storedValue2) != "original" {
			t.Errorf("Expected 'original', got %s", storedValue2)
		}
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	store := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

	t.Run("DeleteExistingKey", func(t *testing.T) {
		key := "key-to-delete"
		value := []byte("value-to-delete")

		// First put the key
		err := store.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}

		// Verify it exists
		_, found, err := store.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("Key should exist before deletion")
		}

		// Delete the key
		err = store.Delete(key)
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		// Verify it no longer exists
		_, found, err = store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if found {
			t.Error("Key should not exist after deletion")
		}
	})

	t.Run("DeleteNonExistentKey", func(t *testing.T) {
		key := "non-existent-key"

		// Delete a key that doesn't exist (should not error)
		err := store.Delete(key)
		if err != nil {
			t.Errorf("Delete of non-existent key failed: %v", err)
		}
	})

	t.Run("DeleteEmptyKey", func(t *testing.T) {
		key := ""

		err := store.Delete(key)
		if err == nil {
			t.Error("Expected error when deleting with empty key")
		}
	})

	t.Run("DeleteAfterClose", func(t *testing.T) {
		tempStore := createTestStore(t)
		key := "test-key"

		// Close the store
		if err := tempStore.Close(); err != nil {
			t.Logf("Failed to close temp store: %v", err)
		}

		// Try to delete after close
		err := tempStore.Delete(key)
		if err == nil {
			t.Error("Expected error when deleting from closed store")
		}
	})

	t.Run("DeleteMultipleTimes", func(t *testing.T) {
		key := "multiple-delete-key"
		value := []byte("value")

		// Put and delete multiple times
		for i := 0; i < 3; i++ {
			err := store.Put(key, value)
			if err != nil {
				t.Errorf("Put iteration %d failed: %v", i, err)
			}

			err = store.Delete(key)
			if err != nil {
				t.Errorf("Delete iteration %d failed: %v", i, err)
			}

			// Verify it's deleted
			_, found, err := store.Get(key)
			if err != nil {
				t.Errorf("Get iteration %d failed: %v", i, err)
			}
			if found {
				t.Errorf("Key should not exist after deletion iteration %d", i)
			}
		}
	})
}

func TestMemoryStore_Scan(t *testing.T) {
	store := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

	// Setup test data
	testData := map[string][]byte{
		"user:1":       []byte("alice"),
		"user:2":       []byte("bob"),
		"user:3":       []byte("charlie"),
		"product:1":    []byte("laptop"),
		"product:2":    []byte("mouse"),
		"config:debug": []byte("true"),
		"config:port":  []byte("8080"),
		"other":        []byte("data"),
	}

	// Put all test data
	for key, value := range testData {
		err := store.Put(key, value)
		if err != nil {
			t.Fatalf("Failed to put test data: %v", err)
		}
	}

	t.Run("ScanWithPrefix", func(t *testing.T) {
		result, err := store.Scan("user:")
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		expected := map[string][]byte{
			"user:1": []byte("alice"),
			"user:2": []byte("bob"),
			"user:3": []byte("charlie"),
		}

		if len(result) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(result))
		}

		for key, expectedValue := range expected {
			value, found := result[key]
			if !found {
				t.Errorf("Expected key %s not found in results", key)
				continue
			}
			if string(value) != string(expectedValue) {
				t.Errorf("For key %s: expected %s, got %s", key, expectedValue, value)
			}
		}
	})

	t.Run("ScanWithEmptyPrefix", func(t *testing.T) {
		result, err := store.Scan("")
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		// Should return all keys
		if len(result) != len(testData) {
			t.Errorf("Expected %d results, got %d", len(testData), len(result))
		}

		for key, expectedValue := range testData {
			value, found := result[key]
			if !found {
				t.Errorf("Expected key %s not found in results", key)
				continue
			}
			if string(value) != string(expectedValue) {
				t.Errorf("For key %s: expected %s, got %s", key, expectedValue, value)
			}
		}
	})

	t.Run("ScanWithNonExistentPrefix", func(t *testing.T) {
		result, err := store.Scan("nonexistent:")
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 results, got %d", len(result))
		}
	})

	t.Run("ScanAfterClose", func(t *testing.T) {
		tempStore := createTestStore(t)

		// Put some data
		err := tempStore.Put("test:key", []byte("value"))
		if err != nil {
			t.Fatal(err)
		}

		// Close the store
		if err := tempStore.Close(); err != nil {
			t.Logf("Failed to close temp store: %v", err)
		}

		// Try to scan after close
		_, err = tempStore.Scan("test:")
		if err == nil {
			t.Error("Expected error when scanning closed store")
		}
	})

	t.Run("ScanDataIsolation", func(t *testing.T) {
		// Test that modifying returned values doesn't affect stored data
		result, err := store.Scan("user:")
		if err != nil {
			t.Fatal(err)
		}

		// Modify a returned value
		if value, found := result["user:1"]; found {
			value[0] = 'X' // Modify the returned slice
		}

		// Scan again and verify original data is unchanged
		result2, err := store.Scan("user:")
		if err != nil {
			t.Fatal(err)
		}

		if value, found := result2["user:1"]; found {
			if string(value) != "alice" {
				t.Errorf("Expected 'alice', got %s", value)
			}
		} else {
			t.Error("Expected user:1 to be found")
		}
	})
}

func TestMemoryStore_Concurrency(t *testing.T) {
	store := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Failed to close store: %v", err)
		}
	}()

	t.Run("ConcurrentOperations", func(t *testing.T) {
		// This test verifies that concurrent operations don't cause data races
		numGoroutines := 10
		numOperations := 100

		done := make(chan bool, numGoroutines)

		// Start multiple goroutines performing operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("key-%d-%d", id, j)
					value := []byte(fmt.Sprintf("value-%d-%d", id, j))

					// Put
					err := store.Put(key, value)
					if err != nil {
						t.Errorf("Put failed: %v", err)
						return
					}

					// Get
					_, found, err := store.Get(key)
					if err != nil {
						t.Errorf("Get failed: %v", err)
						return
					}
					if !found {
						t.Errorf("Key %s not found", key)
						return
					}

					// Delete
					err = store.Delete(key)
					if err != nil {
						t.Errorf("Delete failed: %v", err)
						return
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

// Helper function to create a test store
func createTestStore(t *testing.T) *MemoryStore {
	config := &MemoryStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level for quiet tests
			NumVersionsToKeep: 1,
		},
	}

	store, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	return store
}

// TestMemoryStoreImplementsInterface verifies that MemoryStore implements the Store interface
func TestMemoryStoreImplementsInterface(t *testing.T) {
	memStore := createTestStore(t)
	defer func() {
		if err := memStore.Close(); err != nil {
			t.Logf("Failed to close memory store: %v", err)
		}
	}()

	// Test that MemoryStore implements Store interface
	var _ store.Store = memStore

	// Test that MemoryStore implements individual interfaces
	var _ store.Getter = memStore
	var _ store.Putter = memStore
	var _ store.Deleter = memStore
	var _ store.Scanner = memStore

	// Test basic interface compliance
	key := "interface-test"
	value := []byte("interface-value")

	// Test Putter interface
	err := memStore.Put(key, value)
	if err != nil {
		t.Errorf("Put method failed: %v", err)
	}

	// Test Getter interface
	retrievedValue, found, err := memStore.Get(key)
	if err != nil {
		t.Errorf("Get method failed: %v", err)
	}
	if !found {
		t.Error("Key should be found")
	}
	if string(retrievedValue) != string(value) {
		t.Errorf("Expected %s, got %s", value, retrievedValue)
	}

	// Test Scanner interface
	results, err := memStore.Scan("interface")
	if err != nil {
		t.Errorf("Scan method failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test Deleter interface
	err = memStore.Delete(key)
	if err != nil {
		t.Errorf("Delete method failed: %v", err)
	}

	// Verify deletion
	_, found, err = memStore.Get(key)
	if err != nil {
		t.Errorf("Get after delete failed: %v", err)
	}
	if found {
		t.Error("Key should not be found after deletion")
	}
}
