# Memory Store Implementation

This document describes the in-memory key-value store implementation that satisfies the `Store` interface.

## Overview

The `MemoryStore` is a thread-safe, in-memory implementation of the `Store` interface. It uses a Go map for storage and provides all the standard operations: Get, Put, Delete, Scan, and Close.

## Features

- **Thread-safe**: Uses `sync.RWMutex` for concurrent read/write operations
- **Data isolation**: Returns copies of data to prevent external modification
- **Memory efficient**: Automatically cleaned up when closed
- **Fast operations**: All operations are O(1) or O(n) for scanning
- **Configurable**: Supports dependency injection with configuration

## Usage

### Basic Usage

```go
// Create with default configuration
store, err := memory.NewWithDefaults()
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Store a value
err = store.Put("user:1", []byte("alice@example.com"))
if err != nil {
    log.Fatal(err)
}

// Retrieve a value
value, found, err := store.Get("user:1")
if err != nil {
    log.Fatal(err)
}
if found {
    fmt.Printf("Found: %s\n", string(value))
}
```

### Dependency Injection

```go
// Create with custom configuration
config := &memory.MemoryStoreConfig{
    StoreConfig: store.StoreConfig{
        LoggingLevel:      1, // INFO level
        NumVersionsToKeep: 5,
    },
}

store, err := memory.New(config)
if err != nil {
    log.Fatal(err)
}
defer store.Close()
```

## API Reference

### Constructor Functions

- `New(config *MemoryStoreConfig) (*MemoryStore, error)` - Creates a store with custom configuration
- `NewWithDefaults() (*MemoryStore, error)` - Creates a store with default configuration

### Store Interface Methods

- `Get(key string) ([]byte, bool, error)` - Retrieves a value by key
- `Put(key string, value []byte) error` - Stores a key-value pair
- `Delete(key string) error` - Removes a key-value pair
- `Scan(prefix string) (map[string][]byte, error)` - Returns all keys with given prefix
- `Close() error` - Closes the store and clears memory

## Configuration

The `MemoryStoreConfig` embeds `store.StoreConfig` to provide common configuration options:

```go
type MemoryStoreConfig struct {
    store.StoreConfig // Embedded configuration
}

type StoreConfig struct {
    LoggingLevel      int // Logging verbosity level
    NumVersionsToKeep int // Number of versions to keep (used by underlying systems)
}
```

## Error Handling

The memory store returns errors in the following cases:

- **Empty key**: Operations with empty string keys return an error
- **Closed store**: Operations on a closed store return "store is closed" error
- **Nil configuration**: Creating with nil config returns "config cannot be nil" error

## Thread Safety

The `MemoryStore` is fully thread-safe:

- **Read operations** (Get, Scan) use read locks (`RLock()`)
- **Write operations** (Put, Delete, Close) use write locks (`Lock()`)
- **Data isolation** ensures returned data cannot affect internal storage

## Memory Management

- Data is stored as copies to prevent external modification
- Memory is automatically released when the store is closed
- Large values are handled efficiently without memory leaks

## Performance Characteristics

- **Get**: O(1) average case
- **Put**: O(1) average case
- **Delete**: O(1) average case
- **Scan**: O(n) where n is the total number of keys
- **Memory usage**: O(k + v) where k is total key size and v is total value size

## Testing

The memory store includes comprehensive tests covering:

- Basic CRUD operations
- Error conditions
- Thread safety
- Data isolation
- Interface compliance
- Edge cases (empty values, large values, special characters)

Run tests with:
```bash
go test ./internal/store/memory/... -v
```

## Comparison with BadgerStore

| Feature | MemoryStore | BadgerStore |
|---------|-------------|-------------|
| Persistence | No | Yes |
| Performance | Faster | Good |
| Memory Usage | Higher | Lower |
| Durability | None | High |
| Use Case | Testing, Caching | Production Storage |

## Use Cases

The memory store is ideal for:

- **Unit testing**: Fast, isolated test environments
- **Caching**: Temporary data storage
- **Development**: Quick prototyping without persistence
- **In-memory databases**: When persistence is not required
