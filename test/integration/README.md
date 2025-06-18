# Integration Test Configuration

This directory contains comprehensive integration tests for the Clavis gRPC server implementation.

## Test Files

- `grpc_integration_test.go` - Core integration tests covering basic operations, validation, multiple clients, large data, error handling, and persistence
- `benchmarks_test.go` - Performance benchmarks for various operations and workload patterns
- `stress_test.go` - Stress tests and edge case testing
- `helpers.go` - Utility functions and helpers for integration testing

## Running Tests

### Basic Integration Tests
```bash
# Run all integration tests
go test ./test/integration/

# Run specific test
go test ./test/integration/ -run TestGRPCServer_Integration_BasicOperations

# Run with verbose output
go test -v ./test/integration/
```

### Benchmark Tests
```bash
# Run all benchmarks
go test -bench=. ./test/integration/

# Run specific benchmark
go test -bench=BenchmarkGRPCServer_Put ./test/integration/

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./test/integration/

# Run benchmarks multiple times for more stable results
go test -bench=. -count=5 ./test/integration/
```

### Stress Tests
```bash
# Run stress tests (these take longer)
go test ./test/integration/ -run TestGRPCServer_Integration_StressTest

# Skip stress tests in short mode
go test -short ./test/integration/
```

### Coverage
```bash
# Run tests with coverage
go test -cover ./test/integration/

# Generate detailed coverage report
go test -coverprofile=coverage.out ./test/integration/
go tool cover -html=coverage.out
```

## Test Categories

### 1. Basic Operations (`TestGRPCServer_Integration_BasicOperations`)
- Put, Get, Delete operations
- Non-existent key handling
- Basic data integrity

### 2. Validation Tests (`TestGRPCServer_Integration_ValidationErrors`)
- Empty key validation
- Oversized value validation (>100MB)
- Long key validation (>1024 chars)

### 3. Multi-client Tests (`TestGRPCServer_Integration_MultipleClients`)
- Concurrent operations from multiple clients
- Data isolation and consistency

### 4. Large Data Tests (`TestGRPCServer_Integration_LargeData`)
- 1KB, 1MB, and 10MB value handling
- Data integrity verification
- Performance with large payloads

### 5. Error Handling (`TestGRPCServer_Integration_ErrorHandling`)
- gRPC error code validation
- Edge case error scenarios

### 6. Persistence Tests (`TestGRPCServer_Integration_Persistence`)
- Data survives server restart
- BadgerDB durability validation

### 7. Stress Tests (`TestGRPCServer_Integration_StressTest`)
- High concurrency (50 goroutines, 100 ops each)
- System stability under load

### 8. Edge Cases (`TestGRPCServer_Integration_EdgeCases`)
- Special characters in keys
- Binary data handling
- Repeated operations
- Near-limit values

### 9. Timeout Handling (`TestGRPCServer_Integration_TimeoutHandling`)
- Short timeout scenarios
- Deadline exceeded handling

### 10. Data Integrity (`TestGRPCServer_Integration_DataIntegrity`)
- Large dataset integrity
- Random access patterns
- Data corruption detection

## Benchmarks

### Single Operation Benchmarks
- `BenchmarkGRPCServer_Put_*` - Put operation performance
- `BenchmarkGRPCServer_Get_*` - Get operation performance
- `BenchmarkGRPCServer_Delete_*` - Delete operation performance

### Parallel Benchmarks
- `BenchmarkGRPCServer_*_Parallel` - Concurrent operation performance

### Mixed Workload Benchmarks
- `BenchmarkGRPCServer_MixedWorkload_ReadHeavy` - 80% reads, 20% writes
- `BenchmarkGRPCServer_MixedWorkload_WriteHeavy` - 80% writes, 20% reads

## Test Infrastructure

### TestServer
- Manages temporary BadgerDB instances
- Handles server lifecycle (start/stop)
- Provides gRPC client creation
- Automatic cleanup of resources

### Helpers
- `BenchmarkHelper` - Utilities for performance testing
- `TestDataGenerator` - Generates test datasets
- `LoadTestHelper` - Concurrent operation testing
- `RetryHelper` - Retry logic for flaky operations

## Test Environment

Each test creates:
1. Temporary directory for BadgerDB data
2. BadgerDB store instance
3. Validated store wrapper (with default validators)
4. gRPC server on random available port
5. gRPC client connection

All resources are automatically cleaned up after each test.

## Validation Integration

Tests use the complete validation stack:
- Key validation (non-empty, max 1024 chars)
- Value validation (max 100MB)
- Error propagation through gRPC status codes

## Best Practices

1. Each test is independent and isolated
2. Temporary resources are always cleaned up
3. Timeouts are used to prevent hanging tests
4. Error cases are explicitly tested
5. Data integrity is verified at multiple levels
6. Concurrent access patterns are tested
7. Edge cases and boundary conditions are covered
