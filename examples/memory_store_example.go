package examples

import (
	"fmt"
	"log"

	"github.com/William-Fernandes252/clavis/internal/store"
	"github.com/William-Fernandes252/clavis/internal/store/memory"
)

func RunMemoryStoreExample() {
	fmt.Println("=== Memory Store Example ===")

	// Create a memory store with default configuration
	fmt.Println("\n1. Creating memory store with default configuration...")
	memStore, err := memory.NewWithDefaults()
	if err != nil {
		log.Fatalf("Failed to create memory store: %v", err)
	}
	defer memStore.Close()

	// Create a memory store with custom configuration
	fmt.Println("\n2. Creating memory store with custom configuration...")
	customConfig := &memory.MemoryStoreConfig{
		StoreConfig: store.StoreConfig{
			LoggingLevel:      1, // INFO level
			NumVersionsToKeep: 5, // Keep more versions
		},
	}
	memStoreCustom, err := memory.New(customConfig)
	if err != nil {
		log.Fatalf("Failed to create custom memory store: %v", err)
	}
	defer memStoreCustom.Close()

	// Use the default store for the examples
	fmt.Println("\n3. Basic operations with memory store...")

	// Put some data
	fmt.Println("Storing key-value pairs...")
	data := map[string][]byte{
		"user:1":       []byte("alice@example.com"),
		"user:2":       []byte("bob@example.com"),
		"user:3":       []byte("charlie@example.com"),
		"product:1":    []byte("laptop"),
		"product:2":    []byte("mouse"),
		"config:debug": []byte("true"),
		"config:port":  []byte("8080"),
	}

	for key, value := range data {
		err := memStore.Put(key, value)
		if err != nil {
			log.Printf("Failed to put %s: %v", key, err)
		} else {
			fmt.Printf("  ? Stored: %s = %s\n", key, string(value))
		}
	}

	// Get individual values
	fmt.Println("\nRetrieving individual values...")
	testKeys := []string{"user:1", "product:1", "nonexistent:key"}

	for _, key := range testKeys {
		value, found, err := memStore.Get(key)
		if err != nil {
			log.Printf("Failed to get %s: %v", key, err)
		} else if found {
			fmt.Printf("  ? Found: %s = %s\n", key, string(value))
		} else {
			fmt.Printf("  ? Not found: %s\n", key)
		}
	}

	// Scan with prefix
	fmt.Println("\nScanning with prefixes...")
	prefixes := []string{"user:", "product:", "config:", "admin:"}

	for _, prefix := range prefixes {
		results, err := memStore.Scan(prefix)
		if err != nil {
			log.Printf("Failed to scan with prefix %s: %v", prefix, err)
		} else {
			fmt.Printf("  Prefix '%s' found %d items:\n", prefix, len(results))
			for key, value := range results {
				fmt.Printf("    %s = %s\n", key, string(value))
			}
		}
	}

	// Update a value
	fmt.Println("\nUpdating a value...")
	updateKey := "config:debug"
	newValue := []byte("false")
	err = memStore.Put(updateKey, newValue)
	if err != nil {
		log.Printf("Failed to update %s: %v", updateKey, err)
	} else {
		fmt.Printf("  ? Updated: %s = %s\n", updateKey, string(newValue))

		// Verify the update
		value, found, err := memStore.Get(updateKey)
		if err != nil {
			log.Printf("Failed to verify update: %v", err)
		} else if found {
			fmt.Printf("  ? Verified: %s = %s\n", updateKey, string(value))
		}
	}

	// Delete some keys
	fmt.Println("\nDeleting keys...")
	deleteKeys := []string{"user:2", "product:1"}

	for _, key := range deleteKeys {
		err := memStore.Delete(key)
		if err != nil {
			log.Printf("Failed to delete %s: %v", key, err)
		} else {
			fmt.Printf("  ? Deleted: %s\n", key)

			// Verify deletion
			_, found, err := memStore.Get(key)
			if err != nil {
				log.Printf("Failed to verify deletion: %v", err)
			} else if !found {
				fmt.Printf("    ? Confirmed: %s no longer exists\n", key)
			} else {
				fmt.Printf("    ? Error: %s still exists after deletion\n", key)
			}
		}
	}

	// Final scan to show remaining data
	fmt.Println("\nFinal state - all remaining data:")
	allData, err := memStore.Scan("")
	if err != nil {
		log.Printf("Failed to scan all data: %v", err)
	} else {
		fmt.Printf("  Total remaining items: %d\n", len(allData))
		for key, value := range allData {
			fmt.Printf("    %s = %s\n", key, string(value))
		}
	}

	// Demonstrate memory-specific features
	fmt.Println("\n4. Memory store specific features...")

	// Show that data is truly in-memory (lost after close)
	tempStore, err := memory.NewWithDefaults()
	if err != nil {
		log.Fatalf("Failed to create temp store: %v", err)
	}

	// Put some data
	err = tempStore.Put("temp:data", []byte("temporary"))
	if err != nil {
		log.Printf("Failed to put temp data: %v", err)
	} else {
		fmt.Println("  ? Stored temporary data")
	}

	// Verify it exists
	value, found, err := tempStore.Get("temp:data")
	if err != nil {
		log.Printf("Failed to get temp data: %v", err)
	} else if found {
		fmt.Printf("  ? Confirmed temp data exists: %s\n", string(value))
	}

	// Close the store
	tempStore.Close()

	// Try to access after close (should fail)
	_, found, err = tempStore.Get("temp:data")
	if err != nil {
		fmt.Printf("  ? Confirmed: Cannot access data after close (%v)\n", err)
	} else {
		fmt.Printf("  ? Unexpected: Data still accessible after close\n")
	}

	fmt.Println("\n=== Memory Store Example Complete ===")
}
