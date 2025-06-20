//go:build !codeanalysis

package examples

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
	"github.com/William-Fernandes252/clavis/internal/store/validation"
)

func functional_validation_example() {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "functional-validation-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create base BadgerStore
	config := &badger.BadgerStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      3, // ERROR level for quiet demo
			NumVersionsToKeep: 1,
		},
		Path:       tempDir,
		SyncWrites: false, // Faster for demo
	}

	baseStore, err := badger.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := baseStore.Close(); err != nil {
			log.Printf("Failed to close base store: %v", err)
		}
	}()

	// Example 1: Basic validation
	fmt.Println("=== Example 1: Basic Validation ===")
	basicStore := validation.NewWithDefaultValidators(baseStore)
	defer func() {
		if err := basicStore.Close(); err != nil {
			log.Printf("Failed to close basic store: %v", err)
		}
	}()

	// This should work
	err = basicStore.Put("valid-key", []byte("valid-value"))
	if err != nil {
		fmt.Printf("? Unexpected error: %v\n", err)
	} else {
		fmt.Println("? Valid key/value accepted")
	}

	// This should fail (empty key)
	err = basicStore.Put("", []byte("value"))
	if err != nil {
		fmt.Printf("? Empty key rejected: %v\n", err)
	} else {
		fmt.Println("? Empty key should have been rejected")
	}

	// Example 2: Custom validation for a user management system
	fmt.Println("\n=== Example 2: User Management System ===")

	// Custom validators for user keys
	userKeyValidator := validation.ComposeKeyValidators(
		validation.ValidateNonEmptyKey,
		validation.ValidateKeyLength(50),
		validateUserKeyFormat,
		validateUserIdRange,
	)

	// Custom validators for user data
	userValueValidator := validation.ComposeValueValidators(
		validation.ValidateValueSize(10*1024), // 10KB max user data
		validateUserDataFormat,
	)

	userStore := validation.New(baseStore, userKeyValidator, userValueValidator)
	defer func() {
		if err := userStore.Close(); err != nil {
			log.Printf("Failed to close user store: %v", err)
		}
	}()

	// Test user validation
	testUserValidation(userStore)

	// Example 3: Configuration system with different rules
	fmt.Println("\n=== Example 3: Configuration System ===")

	configKeyValidator := validation.ComposeKeyValidators(
		validation.ValidateNonEmptyKey,
		validation.ValidateKeyLength(100),
		validateConfigKeyFormat,
	)

	configValueValidator := validation.ComposeValueValidators(
		validation.ValidateValueSize(1024*1024), // 1MB max config
		validateJSONFormat,
	)

	configStore := validation.New(baseStore, configKeyValidator, configValueValidator)
	defer func() {
		if err := configStore.Close(); err != nil {
			log.Printf("Failed to close config store: %v", err)
		}
	}()

	// Test config validation
	testConfigValidation(configStore)

	// Example 4: Composing different validation strategies
	fmt.Println("\n=== Example 4: Multi-tenant System ===")

	// Different validation for different tenants
	tenantValidators := map[string]store.Store{
		"premium": createPremiumTenantStore(baseStore),
		"basic":   createBasicTenantStore(baseStore),
		"free":    createFreeTenantStore(baseStore),
	}

	testMultiTenantValidation(tenantValidators)
}

// Custom validation functions for user management
func validateUserKeyFormat(key string) error {
	if !strings.HasPrefix(key, "user:") {
		return fmt.Errorf("user keys must start with 'user:'")
	}

	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return fmt.Errorf("user key format must be 'user:id'")
	}

	return nil
}

func validateUserIdRange(key string) error {
	if !strings.HasPrefix(key, "user:") {
		return nil // Not a user key, skip this validation
	}

	parts := strings.Split(key, ":")
	if len(parts) == 2 {
		userId := parts[1]
		if len(userId) < 3 || len(userId) > 20 {
			return fmt.Errorf("user ID must be between 3-20 characters")
		}
	}

	return nil
}

func validateUserDataFormat(key string, value []byte) error {
	// Simple check: user data should contain basic fields
	data := string(value)
	if !strings.Contains(data, "name") {
		return fmt.Errorf("user data must contain 'name' field")
	}
	return nil
}

// Custom validation functions for configuration
func validateConfigKeyFormat(key string) error {
	if !strings.HasPrefix(key, "config:") {
		return fmt.Errorf("config keys must start with 'config:'")
	}
	return nil
}

func validateJSONFormat(key string, value []byte) error {
	// Simple JSON validation (just check for braces)
	data := strings.TrimSpace(string(value))
	if !strings.HasPrefix(data, "{") || !strings.HasSuffix(data, "}") {
		return fmt.Errorf("config values must be valid JSON objects")
	}
	return nil
}

// Test functions
func testUserValidation(userStore store.Store) {
	// Valid user
	err := userStore.Put("user:john123", []byte(`{"name": "John Doe", "email": "john@example.com"}`))
	if err != nil {
		fmt.Printf("? Valid user rejected: %v\n", err)
	} else {
		fmt.Println("? Valid user accepted")
	}

	// Invalid format
	err = userStore.Put("john123", []byte(`{"name": "John"}`))
	if err != nil {
		fmt.Printf("? Invalid format rejected: %v\n", err)
	} else {
		fmt.Println("? Invalid format should have been rejected")
	}

	// Short user ID
	err = userStore.Put("user:jo", []byte(`{"name": "Jo"}`))
	if err != nil {
		fmt.Printf("? Short user ID rejected: %v\n", err)
	} else {
		fmt.Println("? Short user ID should have been rejected")
	}

	// Missing name field
	err = userStore.Put("user:alice123", []byte(`{"email": "alice@example.com"}`))
	if err != nil {
		fmt.Printf("? Invalid user data rejected: %v\n", err)
	} else {
		fmt.Println("? Invalid user data should have been rejected")
	}
}

func testConfigValidation(configStore store.Store) {
	// Valid config
	err := configStore.Put("config:database", []byte(`{"host": "localhost", "port": 5432}`))
	if err != nil {
		fmt.Printf("? Valid config rejected: %v\n", err)
	} else {
		fmt.Println("? Valid config accepted")
	}

	// Invalid format
	err = configStore.Put("database", []byte(`{"host": "localhost"}`))
	if err != nil {
		fmt.Printf("? Invalid config format rejected: %v\n", err)
	} else {
		fmt.Println("? Invalid config format should have been rejected")
	}

	// Invalid JSON
	err = configStore.Put("config:redis", []byte(`host=localhost,port=6379`))
	if err != nil {
		fmt.Printf("? Invalid JSON rejected: %v\n", err)
	} else {
		fmt.Println("? Invalid JSON should have been rejected")
	}
}

// Multi-tenant validation examples
func createPremiumTenantStore(baseStore store.Store) store.Store {
	keyValidator := validation.ComposeKeyValidators(
		validation.ValidateNonEmptyKey,
		validation.ValidateKeyLength(200), // Longer keys allowed
	)

	valueValidator := validation.ComposeValueValidators(
		validation.ValidateValueSize(10 * 1024 * 1024), // 10MB for premium
	)

	return validation.New(baseStore, keyValidator, valueValidator)
}

func createBasicTenantStore(baseStore store.Store) store.Store {
	keyValidator := validation.ComposeKeyValidators(
		validation.ValidateNonEmptyKey,
		validation.ValidateKeyLength(100), // Medium key length
	)

	valueValidator := validation.ComposeValueValidators(
		validation.ValidateValueSize(1 * 1024 * 1024), // 1MB for basic
	)

	return validation.New(baseStore, keyValidator, valueValidator)
}

func createFreeTenantStore(baseStore store.Store) store.Store {
	keyValidator := validation.ComposeKeyValidators(
		validation.ValidateNonEmptyKey,
		validation.ValidateKeyLength(50), // Short keys only
		func(key string) error {
			// Free tier restrictions
			if strings.Contains(key, "premium") || strings.Contains(key, "enterprise") {
				return fmt.Errorf("free tier cannot use premium/enterprise keys")
			}
			return nil
		},
	)

	valueValidator := validation.ComposeValueValidators(
		validation.ValidateValueSize(100 * 1024), // 100KB for free
	)

	return validation.New(baseStore, keyValidator, valueValidator)
}

func testMultiTenantValidation(tenantStores map[string]store.Store) {
	// Test premium tenant
	premiumStore := tenantStores["premium"]
	defer func() {
		if err := premiumStore.Close(); err != nil {
			log.Printf("Failed to close premium store: %v", err)
		}
	}()

	largeValue := make([]byte, 5*1024*1024) // 5MB
	err := premiumStore.Put("premium:large-data", largeValue)
	if err != nil {
		fmt.Printf("? Premium store rejected large data: %v\n", err)
	} else {
		fmt.Println("? Premium store accepted large data")
	}

	// Test free tenant with same data
	freeStore := tenantStores["free"]
	defer func() {
		if err := freeStore.Close(); err != nil {
			log.Printf("Failed to close free store: %v", err)
		}
	}()

	err = freeStore.Put("premium:large-data", largeValue)
	if err != nil {
		fmt.Printf("? Free store rejected premium key and large data: %v\n", err)
	} else {
		fmt.Println("? Free store should have rejected premium key and large data")
	}

	// Test free tenant with appropriate data
	err = freeStore.Put("user:data", []byte("small data"))
	if err != nil {
		fmt.Printf("? Free store rejected small data: %v\n", err)
	} else {
		fmt.Println("? Free store accepted appropriate data")
	}
}
