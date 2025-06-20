//go:build !codeanalysis

package examples

import (
	"fmt"
	"log"
	"os"

	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
)

func dependency_injection_example() {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "clavis-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Failed to remove temp directory: %v", err)
		}
	}()

	// Example 1: Using default configuration (backward compatibility)
	fmt.Println("=== Example 1: Default Configuration ===")
	defaultStore, err := badger.NewWithPath(tempDir + "/default")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := defaultStore.Close(); err != nil {
			log.Printf("Failed to close default store: %v", err)
		}
	}()
	fmt.Println("? Created store with default configuration")

	// Example 2: Custom configuration with dependency injection
	fmt.Println("\n=== Example 2: Custom Configuration ===")
	customConfig := &badger.BadgerStoreConfig{
		Path:       tempDir + "/custom",
		SyncWrites: false, // Better performance for non-critical data
		StoreConfig: store.StoreConfig{
			LoggingLevel:      1, // INFO level
			NumVersionsToKeep: 3, // Keep more versions
		},
	}

	customStore, err := badger.New(customConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := customStore.Close(); err != nil {
			log.Printf("Failed to close custom store: %v", err)
		}
	}()
	fmt.Println("? Created store with custom configuration")
	fmt.Printf("  - Path: %s\n", customConfig.Path)
	fmt.Printf("  - LoggingLevel: %d (INFO)\n", customConfig.LoggingLevel)
	fmt.Printf("  - SyncWrites: %t\n", customConfig.SyncWrites)
	fmt.Printf("  - NumVersionsToKeep: %d\n", customConfig.NumVersionsToKeep)

	// Example 3: Production configuration
	fmt.Println("\n=== Example 3: Production Configuration ===")
	prodConfig := &badger.BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level only
			NumVersionsToKeep: 1, // Minimal memory usage
		},
		Path:       tempDir + "/production",
		SyncWrites: true, // Ensure durability
	}

	prodStore, err := badger.New(prodConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := prodStore.Close(); err != nil {
			log.Printf("Failed to close production store: %v", err)
		}
	}()
	fmt.Println("? Created store with production configuration")
	fmt.Printf("  - Path: %s\n", prodConfig.Path)
	fmt.Printf("  - LoggingLevel: %d (ERROR)\n", prodConfig.LoggingLevel)
	fmt.Printf("  - SyncWrites: %t\n", prodConfig.SyncWrites)
	fmt.Printf("  - NumVersionsToKeep: %d\n", prodConfig.NumVersionsToKeep)

	// Test basic operations
	fmt.Println("\n=== Testing Basic Operations ===")
	key := "test-key"
	value := []byte("test-value")

	// Test Put
	if err := defaultStore.Put(key, value); err != nil {
		log.Fatal(err)
	}
	fmt.Println("? Put operation successful")

	// Test Get
	retrievedValue, found, err := defaultStore.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	if found {
		fmt.Printf("? Get operation successful: %s\n", string(retrievedValue))
	} else {
		fmt.Println("? Key not found")
	}

	// Test Delete
	if err := defaultStore.Delete(key); err != nil {
		log.Fatal(err)
	}
	fmt.Println("? Delete operation successful")

	// Verify deletion
	_, found, err = defaultStore.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	if !found {
		fmt.Println("? Verified key was deleted")
	} else {
		fmt.Println("? Key still exists after deletion")
	}

	fmt.Println("\n=== Dependency Injection Benefits Demonstrated ===")
	fmt.Println("1. ? Flexibility - Different configurations for different environments")
	fmt.Println("2. ? Testability - Easy to create test-specific configurations")
	fmt.Println("3. ? Maintainability - Configuration logic centralized")
	fmt.Println("4. ? Backward Compatibility - Existing API still works")
}
