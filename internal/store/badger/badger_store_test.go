package badger

import (
	"os"
	"testing"

	"github.com/William-Fernandes252/clavis/internal/store"
)

func TestBadgerStore_Configuration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("DefaultConfiguration", func(t *testing.T) {
		// Test the backward-compatible constructor
		store, err := NewWithPath(tempDir + "/default")
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()

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
		config := &BadgerStoreConfig{
			StoreConfig: store.StoreConfig{
				LoggingLevel:      1, // INFO level for tests
				NumVersionsToKeep: 3, // Keep more versions for tests
			},
			Path:       tempDir + "/custom",
			SyncWrites: false, // Faster for tests
		}

		store, err := New(config)
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()

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

func TestBadgerStore_Get(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

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
		expectedValue := []byte("empty-key-value")

		// Put with empty key
		err := store.Put(key, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		// Get with empty key
		value, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected empty key to be found")
		}
		if string(value) != string(expectedValue) {
			t.Errorf("Expected %s, got %s", expectedValue, value)
		}
	})

	t.Run("GetWithSpecialCharacters", func(t *testing.T) {
		key := "key/with/special:chars@#$%"
		expectedValue := []byte("special-chars-value")

		err := store.Put(key, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		value, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key with special characters to be found")
		}
		if string(value) != string(expectedValue) {
			t.Errorf("Expected %s, got %s", expectedValue, value)
		}
	})
}

func TestBadgerStore_Put(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	t.Run("PutNewKey", func(t *testing.T) {
		key := "new-key"
		value := []byte("new-value")

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put failed: %v", err)
		}

		// Verify it was stored
		retrievedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found after put")
		}
		if string(retrievedValue) != string(value) {
			t.Errorf("Expected %s, got %s", value, retrievedValue)
		}
	})

	t.Run("PutUpdateExistingKey", func(t *testing.T) {
		key := "update-key"
		originalValue := []byte("original-value")
		updatedValue := []byte("updated-value")

		// Put original value
		err := store.Put(key, originalValue)
		if err != nil {
			t.Fatal(err)
		}

		// Update with new value
		err = store.Put(key, updatedValue)
		if err != nil {
			t.Errorf("Put update failed: %v", err)
		}

		// Verify updated value
		retrievedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found after update")
		}
		if string(retrievedValue) != string(updatedValue) {
			t.Errorf("Expected %s, got %s", updatedValue, retrievedValue)
		}
	})

	t.Run("PutEmptyValue", func(t *testing.T) {
		key := "empty-value-key"
		value := []byte("")

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put with empty value failed: %v", err)
		}

		retrievedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key with empty value to be found")
		}
		if len(retrievedValue) != 0 {
			t.Errorf("Expected empty value, got %v", retrievedValue)
		}
	})

	t.Run("PutNilValue", func(t *testing.T) {
		key := "nil-value-key"
		var value []byte = nil

		err := store.Put(key, value)
		if err != nil {
			t.Errorf("Put with nil value failed: %v", err)
		}

		retrievedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key with nil value to be found")
		}
		if retrievedValue == nil || len(retrievedValue) != 0 {
			t.Errorf("Expected empty byte slice, got %v", retrievedValue)
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
			t.Errorf("Put with large value failed: %v", err)
		}

		retrievedValue, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key with large value to be found")
		}
		if len(retrievedValue) != len(value) {
			t.Errorf("Expected value length %d, got %d", len(value), len(retrievedValue))
		}
		// Check first and last bytes
		if retrievedValue[0] != value[0] || retrievedValue[len(value)-1] != value[len(value)-1] {
			t.Error("Large value content mismatch")
		}
	})
}

func TestBadgerStore_Delete(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	t.Run("DeleteExistingKey", func(t *testing.T) {
		key := "delete-key"
		value := []byte("delete-value")

		// Put the key first
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
			t.Fatal("Expected key to exist before deletion")
		}

		// Delete the key
		err = store.Delete(key)
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		// Verify it's gone
		_, found, err = store.Get(key)
		if err != nil {
			t.Errorf("Get after delete failed: %v", err)
		}
		if found {
			t.Error("Expected key to be deleted")
		}
	})

	t.Run("DeleteNonExistentKey", func(t *testing.T) {
		key := "non-existent-delete-key"

		// Try to delete a key that doesn't exist
		err := store.Delete(key)
		if err != nil {
			t.Errorf("Delete of non-existent key should not fail: %v", err)
		}
	})

	t.Run("DeleteEmptyKey", func(t *testing.T) {
		key := ""
		value := []byte("empty-key-value")

		// Put empty key
		err := store.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}

		// Delete empty key
		err = store.Delete(key)
		if err != nil {
			t.Errorf("Delete of empty key failed: %v", err)
		}

		// Verify it's gone
		_, found, err := store.Get(key)
		if err != nil {
			t.Errorf("Get after delete failed: %v", err)
		}
		if found {
			t.Error("Expected empty key to be deleted")
		}
	})

	t.Run("DeleteMultipleKeys", func(t *testing.T) {
		keys := []string{"multi1", "multi2", "multi3"}
		value := []byte("multi-value")

		// Put multiple keys
		for _, key := range keys {
			err := store.Put(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Delete all keys
		for _, key := range keys {
			err := store.Delete(key)
			if err != nil {
				t.Errorf("Delete of key %s failed: %v", key, err)
			}
		}

		// Verify all are gone
		for _, key := range keys {
			_, found, err := store.Get(key)
			if err != nil {
				t.Errorf("Get after delete failed for key %s: %v", key, err)
			}
			if found {
				t.Errorf("Expected key %s to be deleted", key)
			}
		}
	})
}

func TestBadgerStore_Scan(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	// Setup test data
	testData := map[string][]byte{
		"user:1":    []byte("user1-data"),
		"user:2":    []byte("user2-data"),
		"user:3":    []byte("user3-data"),
		"product:1": []byte("product1-data"),
		"product:2": []byte("product2-data"),
		"config":    []byte("config-data"),
		"userinfo":  []byte("userinfo-data"), // This should not match "user:" prefix
	}

	// Put all test data
	for key, value := range testData {
		err := store.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("ScanWithPrefix", func(t *testing.T) {
		prefix := "user:"
		result, err := store.Scan(prefix)
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		expectedKeys := []string{"user:1", "user:2", "user:3"}
		if len(result) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(result))
		}

		for _, key := range expectedKeys {
			if value, exists := result[key]; !exists {
				t.Errorf("Expected key %s not found in scan result", key)
			} else if string(value) != string(testData[key]) {
				t.Errorf("Value mismatch for key %s: expected %s, got %s", key, testData[key], value)
			}
		}

		// Verify that keys not matching the prefix are not included
		if _, exists := result["userinfo"]; exists {
			t.Error("Key 'userinfo' should not match prefix 'user:'")
		}
		if _, exists := result["config"]; exists {
			t.Error("Key 'config' should not match prefix 'user:'")
		}
	})

	t.Run("ScanWithProductPrefix", func(t *testing.T) {
		prefix := "product:"
		result, err := store.Scan(prefix)
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		expectedKeys := []string{"product:1", "product:2"}
		if len(result) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(result))
		}

		for _, key := range expectedKeys {
			if value, exists := result[key]; !exists {
				t.Errorf("Expected key %s not found in scan result", key)
			} else if string(value) != string(testData[key]) {
				t.Errorf("Value mismatch for key %s: expected %s, got %s", key, testData[key], value)
			}
		}
	})

	t.Run("ScanWithNonExistentPrefix", func(t *testing.T) {
		prefix := "nonexistent:"
		result, err := store.Scan(prefix)
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 keys for non-existent prefix, got %d", len(result))
		}
	})

	t.Run("ScanWithEmptyPrefix", func(t *testing.T) {
		prefix := ""
		result, err := store.Scan(prefix)
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		// Should return all keys
		if len(result) != len(testData) {
			t.Errorf("Expected %d keys for empty prefix, got %d", len(testData), len(result))
		}

		for key, expectedValue := range testData {
			if value, exists := result[key]; !exists {
				t.Errorf("Expected key %s not found in scan result", key)
			} else if string(value) != string(expectedValue) {
				t.Errorf("Value mismatch for key %s: expected %s, got %s", key, expectedValue, value)
			}
		}
	})

	t.Run("ScanAfterDelete", func(t *testing.T) {
		// Delete one of the user keys
		err := store.Delete("user:2")
		if err != nil {
			t.Fatal(err)
		}

		prefix := "user:"
		result, err := store.Scan(prefix)
		if err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		expectedKeys := []string{"user:1", "user:3"} // user:2 should be missing
		if len(result) != len(expectedKeys) {
			t.Errorf("Expected %d keys after delete, got %d", len(expectedKeys), len(result))
		}

		for _, key := range expectedKeys {
			if _, exists := result[key]; !exists {
				t.Errorf("Expected key %s not found in scan result after delete", key)
			}
		}

		// Verify deleted key is not in result
		if _, exists := result["user:2"]; exists {
			t.Error("Deleted key 'user:2' should not be in scan result")
		}
	})
}

func TestBadgerStore_Close(t *testing.T) {
	store := createTestStore(t)

	// Put some data
	err := store.Put("test-key", []byte("test-value"))
	if err != nil {
		t.Fatal(err)
	}

	// Close the store
	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Try to use the store after closing (should fail)
	_, _, err = store.Get("test-key")
	if err == nil {
		t.Error("Expected error when using store after close")
	}
}

// Helper function to create a test store with isolated temporary directory
func createTestStore(t *testing.T) *BadgerStore {
	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up will be handled by individual tests
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level for quiet tests
			NumVersionsToKeep: 1,
		},
		Path:       tempDir,
		SyncWrites: false, // Faster for tests
	}

	store, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	return store
}
