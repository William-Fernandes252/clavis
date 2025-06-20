package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/William-Fernandes252/clavis/api/proto"
)

// BenchmarkHelper provides utilities for benchmarking integration tests
type BenchmarkHelper struct {
	client proto.ClavisClient
	ctx    context.Context
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(client proto.ClavisClient, ctx context.Context) *BenchmarkHelper {
	return &BenchmarkHelper{
		client: client,
		ctx:    ctx,
	}
}

// BenchmarkPut benchmarks the Put operation
func (bh *BenchmarkHelper) BenchmarkPut(b *testing.B, keySize, valueSize int) {
	key := generateString("key-", keySize)
	value := generateBytes(valueSize)

	b.ResetTimer()
	b.SetBytes(int64(len(key) + len(value)))

	for i := 0; i < b.N; i++ {
		req := &proto.PutRequest{
			Key:   fmt.Sprintf("%s-%d", key, i),
			Value: value,
		}
		_, err := bh.client.Put(bh.ctx, req)
		if err != nil {
			b.Fatalf("Put failed: %v", err)
		}
	}
}

// BenchmarkGet benchmarks the Get operation
func (bh *BenchmarkHelper) BenchmarkGet(b *testing.B, keySize, valueSize int) {
	// Pre-populate data
	key := generateString("get-key-", keySize)
	value := generateBytes(valueSize)

	req := &proto.PutRequest{
		Key:   key,
		Value: value,
	}
	_, err := bh.client.Put(bh.ctx, req)
	if err != nil {
		b.Fatalf("Setup Put failed: %v", err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(key) + len(value)))

	for i := 0; i < b.N; i++ {
		getReq := &proto.GetRequest{Key: key}
		_, err := bh.client.Get(bh.ctx, getReq)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

// BenchmarkDelete benchmarks the Delete operation
func (bh *BenchmarkHelper) BenchmarkDelete(b *testing.B, keySize int) {
	key := generateString("del-key-", keySize)

	// Pre-populate data
	for i := 0; i < b.N; i++ {
		req := &proto.PutRequest{
			Key:   fmt.Sprintf("%s-%d", key, i),
			Value: []byte("delete-test-value"),
		}
		_, err := bh.client.Put(bh.ctx, req)
		if err != nil {
			b.Fatalf("Setup Put failed: %v", err)
		}
	}

	b.ResetTimer()
	b.SetBytes(int64(len(key)))

	for i := 0; i < b.N; i++ {
		delReq := &proto.DeleteRequest{Key: fmt.Sprintf("%s-%d", key, i)}
		_, err := bh.client.Delete(bh.ctx, delReq)
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// generateString generates a string of specified length with a prefix
func generateString(prefix string, totalLength int) string {
	if len(prefix) >= totalLength {
		return prefix[:totalLength]
	}

	remaining := totalLength - len(prefix)
	suffix := make([]byte, remaining)
	for i := range suffix {
		suffix[i] = byte('a' + (i % 26))
	}

	return prefix + string(suffix)
}

// generateBytes generates a byte slice of specified length
func generateBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = byte(i % 256)
	}
	return bytes
}

// TestDataGenerator provides utilities for generating test data
type TestDataGenerator struct{}

// GenerateKeyValuePairs generates a specified number of key-value pairs
func (tdg *TestDataGenerator) GenerateKeyValuePairs(count int, keyPrefix string, valueSize int) []struct {
	Key   string
	Value []byte
} {
	pairs := make([]struct {
		Key   string
		Value []byte
	}, count)

	for i := 0; i < count; i++ {
		pairs[i].Key = fmt.Sprintf("%s-%d", keyPrefix, i)
		pairs[i].Value = generateBytes(valueSize)
		// Make each value unique
		copy(pairs[i].Value[:4], []byte(fmt.Sprintf("%04d", i)))
	}

	return pairs
}

// LoadTestHelper provides utilities for load testing
type LoadTestHelper struct {
	client proto.ClavisClient
	ctx    context.Context
}

// NewLoadTestHelper creates a new load test helper
func NewLoadTestHelper(client proto.ClavisClient, ctx context.Context) *LoadTestHelper {
	return &LoadTestHelper{
		client: client,
		ctx:    ctx,
	}
}

// RunConcurrentOperations runs multiple operations concurrently
func (lth *LoadTestHelper) RunConcurrentOperations(t *testing.T, numGoroutines int, operationsPerGoroutine int) {
	done := make(chan error, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < operationsPerGoroutine; i++ {
				key := fmt.Sprintf("load-test-%d-%d", goroutineID, i)
				value := []byte(fmt.Sprintf("value-%d-%d", goroutineID, i))

				// Put
				_, err := lth.client.Put(lth.ctx, &proto.PutRequest{
					Key:   key,
					Value: value,
				})
				if err != nil {
					done <- fmt.Errorf("goroutine %d put %d failed: %w", goroutineID, i, err)
					return
				}

				// Get
				resp, err := lth.client.Get(lth.ctx, &proto.GetRequest{Key: key})
				if err != nil {
					done <- fmt.Errorf("goroutine %d get %d failed: %w", goroutineID, i, err)
					return
				}
				if !resp.Found {
					done <- fmt.Errorf("goroutine %d: key %s not found", goroutineID, key)
					return
				}

				// Verify value
				if string(resp.Value) != string(value) {
					done <- fmt.Errorf("goroutine %d: value mismatch for key %s", goroutineID, key)
					return
				}

				// Delete
				_, err = lth.client.Delete(lth.ctx, &proto.DeleteRequest{Key: key})
				if err != nil {
					done <- fmt.Errorf("goroutine %d delete %d failed: %w", goroutineID, i, err)
					return
				}
			}
			done <- nil
		}(g)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Fatalf("Concurrent operation failed: %v", err)
		}
	}
}

// RetryHelper provides utilities for retry logic in tests
type RetryHelper struct{}

// RetryWithBackoff retries an operation with exponential backoff
func (rh *RetryHelper) RetryWithBackoff(operation func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i < maxRetries-1 {
			// Cap the shift to prevent overflow
			shift := i
			if shift > 10 { // 2^10 = 1024, reasonable cap
				shift = 10
			}
			delay := baseDelay * time.Duration(1<<shift)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}
