package integration

import (
	"context"
	"testing"
	"time"
)

// Benchmark tests for the gRPC server integration

func BenchmarkGRPCServer_Put_SmallValues(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkPut(b, 32, 100) // 32-byte keys, 100-byte values
}

func BenchmarkGRPCServer_Put_MediumValues(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkPut(b, 64, 1024) // 64-byte keys, 1KB values
}

func BenchmarkGRPCServer_Put_LargeValues(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkPut(b, 128, 1024*1024) // 128-byte keys, 1MB values
}

func BenchmarkGRPCServer_Get_SmallValues(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkGet(b, 32, 100) // 32-byte keys, 100-byte values
}

func BenchmarkGRPCServer_Get_MediumValues(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkGet(b, 64, 1024) // 64-byte keys, 1KB values
}

func BenchmarkGRPCServer_Get_LargeValues(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkGet(b, 128, 1024*1024) // 128-byte keys, 1MB values
}

func BenchmarkGRPCServer_Delete_SmallKeys(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkDelete(b, 32) // 32-byte keys
}

func BenchmarkGRPCServer_Delete_LargeKeys(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)
	helper.BenchmarkDelete(b, 128) // 128-byte keys
}

// Parallel benchmark tests

func BenchmarkGRPCServer_Put_Parallel(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	b.RunParallel(func(pb *testing.PB) {
		client, conn := testServer.NewClient(b)
		defer func() {
			if err := conn.Close(); err != nil {
				b.Logf("Failed to close connection: %v", err)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		helper := NewBenchmarkHelper(client, ctx)

		i := 0
		for pb.Next() {
			helper.BenchmarkPut(&testing.B{}, 64, 1024) // Simulate single put operation
			i++
		}
	})
}

func BenchmarkGRPCServer_Get_Parallel(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	// Note: Pre-population setup would go here if needed for this benchmark
	// For this parallel benchmark, each goroutine will create its own client

	b.RunParallel(func(pb *testing.PB) {
		client, conn := testServer.NewClient(b)
		defer func() {
			if err := conn.Close(); err != nil {
				b.Logf("Failed to close connection: %v", err)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		helper := NewBenchmarkHelper(client, ctx)

		for pb.Next() {
			helper.BenchmarkGet(&testing.B{}, 64, 1024) // Simulate single get operation
		}
	})
}

// Mixed workload benchmarks

func BenchmarkGRPCServer_MixedWorkload_ReadHeavy(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Pre-populate data
	generator := &TestDataGenerator{}
	_ = generator.GenerateKeyValuePairs(100, "mixed-key", 512) // Would use for setup if needed

	helper := NewBenchmarkHelper(client, ctx)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 80% reads, 20% writes
		if i%5 == 0 {
			// Write operation
			helper.BenchmarkPut(&testing.B{}, 64, 512)
		} else {
			// Read operation
			helper.BenchmarkGet(&testing.B{}, 64, 512)
		}
	}
}

func BenchmarkGRPCServer_MixedWorkload_WriteHeavy(b *testing.B) {
	testServer := NewTestServer(b)
	defer testServer.Stop()
	testServer.Start(b)

	client, conn := testServer.NewClient(b)
	defer func() {
		if err := conn.Close(); err != nil {
			b.Logf("Failed to close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	helper := NewBenchmarkHelper(client, ctx)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 80% writes, 20% reads
		if i%5 == 4 {
			// Read operation
			helper.BenchmarkGet(&testing.B{}, 64, 512)
		} else {
			// Write operation
			helper.BenchmarkPut(&testing.B{}, 64, 512)
		}
	}
}
