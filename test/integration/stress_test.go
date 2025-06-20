package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/William-Fernandes252/clavis/api/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestGRPCServer_Integration_StressTest performs stress testing with high load
func TestGRPCServer_Integration_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run load test
	loadHelper := NewLoadTestHelper(client, ctx)

	t.Run("HighConcurrency", func(t *testing.T) {
		// Test with 50 goroutines, 100 operations each
		loadHelper.RunConcurrentOperations(t, 50, 100)
	})
}

// TestGRPCServer_Integration_EdgeCases tests various edge cases
func TestGRPCServer_Integration_EdgeCases(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with special characters in keys
	t.Run("SpecialCharacterKeys", func(t *testing.T) {
		specialKeys := []string{
			"key-with-hyphens",
			"key_with_underscores",
			"key.with.dots",
			"key/with/slashes",
			"key with spaces",
			"key@with#special$chars",
			"unicode-key-spanish",
			"emoji-key-symbols",
		}

		for i, key := range specialKeys {
			value := []byte(fmt.Sprintf("value-for-special-key-%d", i))

			// Put
			_, err := client.Put(ctx, &proto.PutRequest{
				Key:   key,
				Value: value,
			})
			if err != nil {
				t.Errorf("Put failed for key '%s': %v", key, err)
				continue
			}

			// Get
			resp, err := client.Get(ctx, &proto.GetRequest{Key: key})
			if err != nil {
				t.Errorf("Get failed for key '%s': %v", key, err)
				continue
			}

			if !resp.Found {
				t.Errorf("Key '%s' not found", key)
				continue
			}

			if string(resp.Value) != string(value) {
				t.Errorf("Value mismatch for key '%s': expected '%s', got '%s'", key, string(value), string(resp.Value))
			}

			// Delete
			_, err = client.Delete(ctx, &proto.DeleteRequest{Key: key})
			if err != nil {
				t.Errorf("Delete failed for key '%s': %v", key, err)
			}
		}
	})

	// Test with binary data
	t.Run("BinaryData", func(t *testing.T) {
		binaryData := [][]byte{
			{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
			make([]byte, 0),          // Empty binary data
			{0x7F, 0x80, 0x81, 0x82}, // ASCII boundary values
		}

		for i, data := range binaryData {
			key := fmt.Sprintf("binary-key-%d", i)

			// Put
			_, err := client.Put(ctx, &proto.PutRequest{
				Key:   key,
				Value: data,
			})
			if err != nil {
				t.Errorf("Put failed for binary data %d: %v", i, err)
				continue
			}

			// Get
			resp, err := client.Get(ctx, &proto.GetRequest{Key: key})
			if err != nil {
				t.Errorf("Get failed for binary data %d: %v", i, err)
				continue
			}

			if !resp.Found {
				t.Errorf("Binary data %d not found", i)
				continue
			}

			if len(resp.Value) != len(data) {
				t.Errorf("Binary data %d length mismatch: expected %d, got %d", i, len(data), len(resp.Value))
				continue
			}

			for j := range data {
				if resp.Value[j] != data[j] {
					t.Errorf("Binary data %d byte %d mismatch: expected %02x, got %02x", i, j, data[j], resp.Value[j])
					break
				}
			}
		}
	})

	// Test with repeated operations on same key
	t.Run("RepeatedOperations", func(t *testing.T) {
		key := "repeated-ops-key"

		// Multiple puts with different values
		for i := 0; i < 10; i++ {
			value := []byte(fmt.Sprintf("value-%d", i))
			_, err := client.Put(ctx, &proto.PutRequest{
				Key:   key,
				Value: value,
			})
			if err != nil {
				t.Fatalf("Put %d failed: %v", i, err)
			}

			// Verify the value was updated
			resp, err := client.Get(ctx, &proto.GetRequest{Key: key})
			if err != nil {
				t.Fatalf("Get %d failed: %v", i, err)
			}

			if string(resp.Value) != string(value) {
				t.Errorf("Value %d mismatch: expected '%s', got '%s'", i, string(value), string(resp.Value))
			}
		}

		// Multiple deletes (first should succeed, rest should be no-ops)
		for i := 0; i < 3; i++ {
			_, err := client.Delete(ctx, &proto.DeleteRequest{Key: key})
			if err != nil {
				t.Errorf("Delete %d failed: %v", i, err)
			}
		}

		// Verify key is gone
		resp, err := client.Get(ctx, &proto.GetRequest{Key: key})
		if err != nil {
			t.Fatalf("Final get failed: %v", err)
		}
		if resp.Found {
			t.Error("Key should not be found after deletion")
		}
	})

	// Test near-limit values
	t.Run("NearLimitValues", func(t *testing.T) {
		// Test with key near the size limit (1024 chars)
		longKey := strings.Repeat("a", 1023) // Just under the limit
		_, err := client.Put(ctx, &proto.PutRequest{
			Key:   longKey,
			Value: []byte("value-for-long-key"),
		})
		if err != nil {
			t.Errorf("Put failed for long key: %v", err)
		}

		// Test with large value (5MB - well under 100MB limit)
		largeValue := make([]byte, 5*1024*1024)
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}

		_, err = client.Put(ctx, &proto.PutRequest{
			Key:   "large-value-key",
			Value: largeValue,
		})
		if err != nil {
			t.Errorf("Put failed for large value: %v", err)
		}

		// Verify large value
		resp, err := client.Get(ctx, &proto.GetRequest{Key: "large-value-key"})
		if err != nil {
			t.Errorf("Get failed for large value: %v", err)
		} else if len(resp.Value) != len(largeValue) {
			t.Errorf("Large value size mismatch: expected %d, got %d", len(largeValue), len(resp.Value))
		}
	})
}

// TestGRPCServer_Integration_TimeoutHandling tests timeout scenarios
func TestGRPCServer_Integration_TimeoutHandling(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}()

	// Test with very short timeout
	t.Run("ShortTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// This should timeout quickly
		_, err := client.Put(ctx, &proto.PutRequest{
			Key:   "timeout-test-key",
			Value: []byte("timeout-test-value"),
		})

		if err == nil {
			t.Error("Expected timeout error, but operation succeeded")
		} else if status.Code(err) != codes.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded error, got %v", status.Code(err))
		}
	})

	// Test with reasonable timeout
	t.Run("ReasonableTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// This should succeed
		_, err := client.Put(ctx, &proto.PutRequest{
			Key:   "normal-key",
			Value: []byte("normal-value"),
		})
		if err != nil {
			t.Errorf("Operation failed with reasonable timeout: %v", err)
		}
	})
}

// TestGRPCServer_Integration_DataIntegrity tests data integrity under various conditions
func TestGRPCServer_Integration_DataIntegrity(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate test data
	generator := &TestDataGenerator{}
	testPairs := generator.GenerateKeyValuePairs(1000, "integrity-test", 512)

	// Store all data
	t.Run("StoreData", func(t *testing.T) {
		for i, pair := range testPairs {
			_, err := client.Put(ctx, &proto.PutRequest{
				Key:   pair.Key,
				Value: pair.Value,
			})
			if err != nil {
				t.Fatalf("Put failed for pair %d: %v", i, err)
			}
		}
	})

	// Verify all data
	t.Run("VerifyData", func(t *testing.T) {
		for i, pair := range testPairs {
			resp, err := client.Get(ctx, &proto.GetRequest{Key: pair.Key})
			if err != nil {
				t.Errorf("Get failed for pair %d: %v", i, err)
				continue
			}

			if !resp.Found {
				t.Errorf("Pair %d not found", i)
				continue
			}

			if len(resp.Value) != len(pair.Value) {
				t.Errorf("Pair %d length mismatch: expected %d, got %d", i, len(pair.Value), len(resp.Value))
				continue
			}

			for j := range pair.Value {
				if resp.Value[j] != pair.Value[j] {
					t.Errorf("Pair %d byte %d mismatch: expected %02x, got %02x", i, j, pair.Value[j], resp.Value[j])
					break
				}
			}
		}
	})

	// Random access pattern
	t.Run("RandomAccess", func(t *testing.T) {
		// Access data in non-sequential order
		indices := []int{500, 100, 800, 200, 900, 50, 750, 300}

		for _, idx := range indices {
			if idx >= len(testPairs) {
				continue
			}

			pair := testPairs[idx]
			resp, err := client.Get(ctx, &proto.GetRequest{Key: pair.Key})
			if err != nil {
				t.Errorf("Random access get failed for index %d: %v", idx, err)
				continue
			}

			if !resp.Found {
				t.Errorf("Random access: pair %d not found", idx)
				continue
			}

			if string(resp.Value) != string(pair.Value) {
				t.Errorf("Random access: pair %d value mismatch", idx)
			}
		}
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		for i, pair := range testPairs {
			_, err := client.Delete(ctx, &proto.DeleteRequest{Key: pair.Key})
			if err != nil {
				t.Errorf("Delete failed for pair %d: %v", i, err)
			}
		}

		// Verify all data is gone
		for i, pair := range testPairs {
			resp, err := client.Get(ctx, &proto.GetRequest{Key: pair.Key})
			if err != nil {
				t.Errorf("Get after delete failed for pair %d: %v", i, err)
				continue
			}

			if resp.Found {
				t.Errorf("Pair %d still found after deletion", i)
			}
		}
	})
}
