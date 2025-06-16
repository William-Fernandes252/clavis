# Store Package

This package provides the low-level storage layer for Clavis.

## Overview

The store package implements a clean interface-based architecture that supports multiple storage backends and composable validation. It follows the dependency injection pattern and provides both persistent and in-memory storage options.

## Architecture

```
???????????????????
?   Store Interface   ?  ? Common interface for all implementations
???????????????????
         ?
         ? implements
    ???????????????????????????????????????
    ?         ?            ?              ?
????????? ??????? ???????????? ????????????
?Memory ? ?Badger? ?Validated ? ?  Future  ?
?Store  ? ?Store ? ?  Store   ? ?  Stores  ?
????????? ??????? ???????????? ????????????
```

## Core Interface

All store implementations satisfy the `Store` interface:

```go
type Store interface {
    io.Closer
    Get(key string) ([]byte, bool, error)
    Put(key string, value []byte) error
    Delete(key string) error
    Scan(prefix string) (map[string][]byte, error)
}
```

## Available Implementations

### 1. Memory Store (`/memory`)
- **Type**: In-memory storage
- **Persistence**: No
- **Performance**: Highest
- **Use Cases**: Testing, caching, temporary data
- **Thread Safety**: Yes (sync.RWMutex)

[?? Memory Store Documentation](./memory/README.md)

### 2. BadgerDB Store (`/badger`)
- **Type**: Persistent storage (LSM-tree)
- **Persistence**: Yes
- **Performance**: High
- **Use Cases**: Production applications, embedded databases
- **Thread Safety**: Yes (BadgerDB handles concurrency)

[?? BadgerDB Store Documentation](./badger/README.md)

### 3. Validation Store (`/validation`)
- **Type**: Decorator/Wrapper
- **Purpose**: Add validation to any store implementation
- **Features**: Composable validators, custom validation rules
- **Use Cases**: Input validation, data integrity, compliance

[?? Validation Store Documentation](./validation/README.md)

## Quick Start

### Basic Usage

```go
import (
    "github.com/William-Fernandes252/clavis/internal/store/memory"
    "github.com/William-Fernandes252/clavis/internal/store/badger"
    "github.com/William-Fernandes252/clavis/internal/store/validation"
)

// Memory store for testing
memStore, err := memory.NewWithDefaults()
if err != nil {
    log.Fatal(err)
}
defer memStore.Close()

// Persistent BadgerDB store
badgerStore, err := badger.NewWithPath("/path/to/database")
if err != nil {
    log.Fatal(err)
}
defer badgerStore.Close()

// Add validation to any store
validatedStore := validation.NewWithDefaultValidators(memStore)
defer validatedStore.Close()
```

### Dependency Injection Pattern

```go
// Configuration-driven store creation
func createStore(config StoreConfig) (Store, error) {
    var baseStore Store
    var err error

    switch config.Type {
    case "memory":
        baseStore, err = memory.New(config.Memory)
    case "badger":
        baseStore, err = badger.New(config.Badger)
    default:
        return nil, fmt.Errorf("unknown store type: %s", config.Type)
    }

    if err != nil {
        return nil, err
    }

    if config.EnableValidation {
        return validation.NewWithDefaultValidators(baseStore), nil
    }

    return baseStore, nil
}
```

## Configuration

### Common Configuration (`StoreConfig`)

All store implementations embed a common configuration structure:

```go
type StoreConfig struct {
    LoggingLevel      int // 0=DEBUG, 1=INFO, 2=WARNING, 3=ERROR
    NumVersionsToKeep int // Number of versions to keep for each key
}
```

### Store-Specific Configuration

Each implementation extends the common configuration:

- **MemoryStoreConfig**: No additional fields (just embeds StoreConfig)
- **BadgerStoreConfig**: Adds `Path` and `SyncWrites` fields
- **ValidatedStore**: Uses functional configuration with validator functions

## Operation Examples

### Basic CRUD Operations

```go
// Put (Create/Update)
err := store.Put("user:123", []byte(`{"name":"Alice","email":"alice@example.com"}`))
if err != nil {
    log.Printf("Put failed: %v", err)
}

// Get (Read)
value, found, err := store.Get("user:123")
if err != nil {
    log.Printf("Get failed: %v", err)
} else if found {
    fmt.Printf("User data: %s\n", string(value))
} else {
    fmt.Println("User not found")
}

// Delete
err = store.Delete("user:123")
if err != nil {
    log.Printf("Delete failed: %v", err)
}

// Scan (List with prefix)
users, err := store.Scan("user:")
if err != nil {
    log.Printf("Scan failed: %v", err)
} else {
    fmt.Printf("Found %d users\n", len(users))
    for key, value := range users {
        fmt.Printf("  %s: %s\n", key, string(value))
    }
}
```

### Error Handling

All store operations return errors for:
- Invalid inputs (empty keys, oversized values)
- Storage backend errors (file I/O, database corruption)
- Validation failures (when using ValidatedStore)

```go
value, found, err := store.Get("some-key")
if err != nil {
    // Handle storage error
    log.Printf("Storage error: %v", err)
    return
}

if !found {
    // Handle missing key
    log.Println("Key not found")
    return
}

// Use value
fmt.Printf("Value: %s\n", string(value))
```

## Performance Characteristics

| Operation | Memory Store | BadgerDB Store | Notes |
|-----------|--------------|----------------|-------|
| Get | O(1) | O(log n) | Memory is fastest |
| Put | O(1) | O(log n) | BadgerDB uses LSM-tree |
| Delete | O(1) | O(log n) | Similar to Put |
| Scan | O(n) | O(k + log n) | k = result size |
| Memory Usage | High | Low | Memory stores all data in RAM |
| Persistence | None | Full | BadgerDB survives restarts |

## Testing Support

The store package provides excellent testing support:

### Test Utilities

```go
// Create isolated test stores
func createTestMemoryStore(t *testing.T) Store {
    store, err := memory.NewWithDefaults()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { store.Close() })
    return store
}

func createTestBadgerStore(t *testing.T) Store {
    tempDir, err := os.MkdirTemp("", "badger-test-*")
    if err != nil {
        t.Fatal(err)
    }

    store, err := badger.NewWithPath(tempDir)
    if err != nil {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        store.Close()
        os.RemoveAll(tempDir)
    })

    return store
}
```

### Interface Compliance Testing

```go
func TestStoreCompliance(t *testing.T) {
    stores := []Store{
        createTestMemoryStore(t),
        createTestBadgerStore(t),
    }

    for _, store := range stores {
        t.Run(fmt.Sprintf("%T", store), func(t *testing.T) {
            testBasicOperations(t, store)
        })
    }
}
```

## Design Patterns

### Decorator Pattern (Validation)

```go
// Stack multiple decorators
baseStore := createMemoryStore()
validatedStore := validation.NewWithDefaultValidators(baseStore)
// Could add more decorators: metrics, caching, etc.
```

### Factory Pattern

```go
type StoreFactory interface {
    CreateStore(config Config) (Store, error)
}

type MemoryStoreFactory struct{}

func (f *MemoryStoreFactory) CreateStore(config Config) (Store, error) {
    return memory.New(config.Memory)
}
```

### Strategy Pattern

```go
type StorageStrategy interface {
    Store
}

type Application struct {
    storage StorageStrategy
}

func (app *Application) SetStorageStrategy(strategy StorageStrategy) {
    if app.storage != nil {
        app.storage.Close()
    }
    app.storage = strategy
}
```

## Best Practices

### 1. Resource Management

```go
// Always close stores
store, err := badger.NewWithPath("/path/to/db")
if err != nil {
    return err
}
defer store.Close() // Ensure cleanup

// Or with explicit error handling
defer func() {
    if err := store.Close(); err != nil {
        log.Printf("Failed to close store: %v", err)
    }
}()
```

### 2. Error Handling

```go
// Check for specific error types if needed
value, found, err := store.Get(key)
if err != nil {
    if isTemporaryError(err) {
        // Retry logic
        return retryOperation()
    }
    return fmt.Errorf("permanent error: %w", err)
}
```

### 3. Configuration

```go
// Use environment-specific configurations
func createProductionStore() (Store, error) {
    config := &badger.BadgerStoreConfig{
        StoreConfig: store.StoreConfig{
            LoggingLevel: 3, // ERROR only in production
            NumVersionsToKeep: 1,
        },
        Path: "/var/lib/myapp/database",
        SyncWrites: true, // Ensure durability
    }
    return badger.New(config)
}

func createTestStore() (Store, error) {
    return memory.NewWithDefaults() // Fast for tests
}
```

### 4. Validation

```go
// Compose validators for specific use cases
userKeyValidator := validation.ComposeKeyValidators(
    validation.ValidateNonEmptyKey,
    validation.ValidateKeyLength(256),
    func(key string) error {
        if !strings.HasPrefix(key, "user:") {
            return fmt.Errorf("key must start with 'user:'")
        }
        return nil
    },
)
```

## Migration Between Stores

```go
func migrateStore(source, destination Store) error {
    // Get all data from source
    allData, err := source.Scan("")
    if err != nil {
        return fmt.Errorf("failed to scan source: %w", err)
    }

    // Put all data into destination
    for key, value := range allData {
        if err := destination.Put(key, value); err != nil {
            return fmt.Errorf("failed to migrate key %s: %w", key, err)
        }
    }

    return nil
}
```

## Extensions and Future Work

The store package is designed for extensibility:

### Planned Extensions
- **Redis Store**: Network-based distributed storage
- **Metrics Store**: Wrapper for operation metrics and monitoring
- **Caching Store**: Multi-level caching with TTL support
- **Encrypted Store**: Transparent encryption/decryption
- **Replicated Store**: Master-slave replication

### Plugin Architecture
```go
type StorePlugin interface {
    Wrap(Store) Store
}

type PluginChain []StorePlugin

func (chain PluginChain) Apply(base Store) Store {
    result := base
    for _, plugin := range chain {
        result = plugin.Wrap(result)
    }
    return result
}
```

## Contributing

When adding new store implementations:

1. **Implement the Store interface**: Ensure full compliance
2. **Add comprehensive tests**: Follow existing test patterns
3. **Document thoroughly**: Create README with examples
4. **Follow naming conventions**: Use consistent package and type names
5. **Support dependency injection**: Accept configuration structs

### Testing New Implementations

```go
func TestNewStoreImplementation(t *testing.T) {
    store := createYourNewStore(t)

    // Run standard compliance tests
    testStoreCompliance(t, store)

    // Add implementation-specific tests
    testSpecificFeatures(t, store)
}
```

## Dependencies

- **Core**: No external dependencies (uses only Go standard library)
- **BadgerDB**: `github.com/dgraph-io/badger/v4`
- **Testing**: `testing` package for comprehensive test suites

## License

This store package is part of the Clavis project. See the project's LICENSE file for details.
