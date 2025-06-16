# BadgerDB Store Implementation

This document describes the BadgerDB-based persistent key-value store implementation that satisfies the `Store` interface.

## Overview

The `BadgerStore` is a persistent, high-performance implementation of the `Store` interface built on top of [BadgerDB](https://github.com/dgraph-io/badger). BadgerDB is an embeddable, persistent, and fast key-value database written in pure Go.

## Features

- **Persistent storage**: Data survives application restarts
- **High performance**: Optimized for SSD storage with LSM-tree architecture
- **ACID transactions**: Supports atomic operations
- **Configurable**: Extensive configuration options for performance tuning
- **Dependency injection**: Supports custom configuration via `BadgerStoreConfig`
- **Thread-safe**: Safe for concurrent use across multiple goroutines

## Usage

### Basic Usage

```go
// Create with default configuration
store, err := badger.NewWithPath("/path/to/database")
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

### Dependency Injection with Custom Configuration

```go
// Create with custom configuration
config := &badger.BadgerStoreConfig{
    StoreConfig: store.StoreConfig{
        LoggingLevel:      1, // INFO level
        NumVersionsToKeep: 3, // Keep 3 versions of each key
    },
    Path:       "/path/to/database",
    SyncWrites: true, // Force sync to disk for durability
}

store, err := badger.New(config)
if err != nil {
    log.Fatal(err)
}
defer store.Close()
```

## Configuration

### BadgerStoreConfig

The `BadgerStoreConfig` struct provides configuration options specific to BadgerDB:

```go
type BadgerStoreConfig struct {
    store.StoreConfig        // Embedded common configuration
    Path              string // Database directory path
    SyncWrites        bool   // Sync writes to disk immediately
}
```

### Common StoreConfig (Embedded)

```go
type StoreConfig struct {
    LoggingLevel      int // 0=DEBUG, 1=INFO, 2=WARNING, 3=ERROR
    NumVersionsToKeep int // Number of versions to keep for each key
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Path` | string | Required | Directory path where BadgerDB files are stored |
| `SyncWrites` | bool | true | Whether to sync writes to disk immediately for durability |
| `LoggingLevel` | int | 3 (ERROR) | Logging verbosity: 0=DEBUG, 1=INFO, 2=WARNING, 3=ERROR |
| `NumVersionsToKeep` | int | 1 | Number of versions to keep for each key (affects storage size) |

### Default Configuration

```go
func DefaultConfig(path string) *BadgerStoreConfig {
    return &BadgerStoreConfig{
        StoreConfig: store.StoreConfig{
            LoggingLevel:      3, // ERROR level for production
            NumVersionsToKeep: 1, // Keep only latest version
        },
        Path:       path,
        SyncWrites: true, // Ensure durability
    }
}
```

## API Reference

### Constructor Functions

- `New(config *BadgerStoreConfig) (*BadgerStore, error)` - Creates a store with custom configuration
- `NewWithPath(path string) (*BadgerStore, error)` - Creates a store with default configuration at specified path

### Store Interface Methods

- `Get(key string) ([]byte, bool, error)` - Retrieves a value by key
- `Put(key string, value []byte) error` - Stores a key-value pair
- `Delete(key string) error` - Removes a key-value pair
- `Scan(prefix string) (map[string][]byte, error)` - Returns all keys with given prefix
- `Close() error` - Closes the database and releases resources

## Error Handling

The BadgerStore returns errors in the following cases:

- **Database errors**: File system issues, corruption, etc.
- **Configuration errors**: Invalid paths, permissions issues
- **Key not found**: `Get` returns `(nil, false, nil)` for non-existent keys
- **Nil configuration**: Creating with nil config returns "config cannot be nil" error

## Performance Characteristics

BadgerDB is optimized for:

- **Write-heavy workloads**: Uses LSM-tree for efficient writes
- **SSD storage**: Designed for modern SSD performance characteristics
- **Large datasets**: Can handle databases larger than RAM
- **High throughput**: Supports high concurrent read/write operations

### Performance Tips

1. **Batch operations**: Group multiple operations in transactions when possible
2. **Appropriate `NumVersionsToKeep`**: Lower values save space, higher values provide history
3. **SSD optimization**: BadgerDB performs best on SSD storage
4. **Memory settings**: Tune BadgerDB memory settings for your workload

## Durability and Reliability

- **WAL (Write-Ahead Log)**: Ensures durability even on system crashes
- **Checksums**: Built-in data integrity verification
- **Atomic operations**: All operations are atomic at the key level
- **Recovery**: Automatic recovery from crashes and corruption

## Storage Layout

BadgerDB stores data in:
- **Value log files**: Store the actual key-value data
- **LSM tree files**: Store sorted string tables for efficient lookups
- **Manifest files**: Store metadata about the database structure

## Monitoring and Debugging

### Logging Levels

- **DEBUG (0)**: Detailed operation logs
- **INFO (1)**: General operational information
- **WARNING (2)**: Warning conditions
- **ERROR (3)**: Error conditions only (recommended for production)

### Metrics

BadgerDB provides internal metrics that can be accessed for monitoring:
- Read/write throughput
- Memory usage
- Compaction statistics
- Cache hit ratios

## Migration and Backup

### Backup

```go
// BadgerDB provides built-in backup functionality
err := db.Backup(backupWriter, timestamp)
```

### Restore

```go
// Restore from backup
err := db.Load(backupReader, maxPendingWrites)
```

## Comparison with MemoryStore

| Feature | BadgerStore | MemoryStore |
|---------|-------------|-------------|
| Persistence | Yes | No |
| Performance | High | Higher |
| Memory Usage | Lower | Higher |
| Durability | High | None |
| Storage Limit | Disk Space | RAM |
| Use Case | Production | Testing/Caching |

## Use Cases

The BadgerStore is ideal for:

- **Production applications**: When data persistence is required
- **High-performance databases**: Applications requiring fast key-value access
- **Embedded databases**: Applications that need a local database without external dependencies
- **Time-series data**: With version keeping for historical data
- **Caching with persistence**: Long-term caching that survives restarts

## Dependencies

- **BadgerDB v4**: `github.com/dgraph-io/badger/v4`

## Testing

Run BadgerStore tests with:
```bash
go test ./internal/store/badger/... -v
```

The test suite includes:
- Configuration validation
- CRUD operations
- Error conditions
- Large value handling
- Concurrent access
- Database lifecycle

## Best Practices

1. **Always close the store**: Use `defer store.Close()` to ensure proper cleanup
2. **Handle errors properly**: BadgerDB operations can fail due to I/O issues
3. **Choose appropriate configuration**: Tune settings for your specific use case
4. **Monitor disk space**: BadgerDB databases can grow large with high write volumes
5. **Regular maintenance**: Consider periodic compaction for optimal performance
