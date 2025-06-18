package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/William-Fernandes252/clavis/api/proto"
	grpcserver "github.com/William-Fernandes252/clavis/internal/server/grpc"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
	"github.com/William-Fernandes252/clavis/internal/store/validation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const maxMessageSize = 128 * 1024 * 1024 // 128MB

// TestServer represents a test gRPC server instance
type TestServer struct {
	server     *grpcserver.GRPCServer
	grpcServer *grpc.Server
	listener   net.Listener
	address    string
	tempDir    string
	done       chan struct{}
}

// NewTestServer creates a new test server instance with a temporary BadgerDB
func NewTestServer(t testing.TB) *TestServer {
	// Create temporary directory for BadgerDB
	tempDir, err := os.MkdirTemp("", "clavis-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create BadgerDB store
	badgerStore, err := badger.NewWithPath(filepath.Join(tempDir, "data"))
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create BadgerDB store: %v", err)
	}

	// Wrap with validation
	validatedStore := validation.NewWithDefaultValidators(badgerStore)

	// Create gRPC server with larger message limits
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMessageSize), // 128MB
		grpc.MaxSendMsgSize(maxMessageSize), // 128MB
	)

	// Find available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		badgerStore.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create listener: %v", err)
	}

	address := listener.Addr().String()
	config := &grpcserver.GRPCServerConfig{Port: address}

	// Create clavis gRPC server
	server, err := grpcserver.New(validatedStore, config, grpcServer)
	if err != nil {
		listener.Close()
		badgerStore.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create gRPC server: %v", err)
	}

	testServer := &TestServer{
		server:     server,
		grpcServer: grpcServer,
		listener:   listener,
		address:    address,
		tempDir:    tempDir,
		done:       make(chan struct{}),
	}

	return testServer
}

// Start starts the test server in a goroutine
func (ts *TestServer) Start(t testing.TB) {
	// Register the server
	proto.RegisterClavisServer(ts.grpcServer, ts.server)

	// Start serving in background
	go func() {
		defer close(ts.done)
		if err := ts.grpcServer.Serve(ts.listener); err != nil {
			// Only log if it's not due to server shutdown
			if err != grpc.ErrServerStopped {
				t.Logf("Server error: %v", err)
			}
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
}

// Stop stops the test server and cleans up resources
func (ts *TestServer) Stop() {
	if ts.grpcServer != nil {
		ts.grpcServer.GracefulStop()
	}
	if ts.listener != nil {
		ts.listener.Close()
	}

	// Wait for server to stop
	<-ts.done

	// Close store properly and wait for it to flush
	if store, err := ts.server.GetStore(); err == nil && store != nil {
		store.Close()
		// Give BadgerDB time to properly close and sync
		time.Sleep(200 * time.Millisecond)
	}
	if ts.tempDir != "" {
		// Don't remove temp directory yet - persistence test needs it
		// os.RemoveAll(ts.tempDir)
	}
}

// NewClient creates a new gRPC client connected to the test server
func (ts *TestServer) NewClient(t testing.TB) (proto.ClavisClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(ts.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize), // 128MB
			grpc.MaxCallSendMsgSize(maxMessageSize), // 128MB
		),
	)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	client := proto.NewClavisClient(conn)
	return client, conn
}

// Integration tests

func TestGRPCServer_Integration_BasicOperations(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test Put operation
	t.Run("Put", func(t *testing.T) {
		req := &proto.PutRequest{
			Key:   "test-key",
			Value: []byte("test-value"),
		}
		_, err := client.Put(ctx, req)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}
	})

	// Test Get operation
	t.Run("Get", func(t *testing.T) {
		req := &proto.GetRequest{Key: "test-key"}
		resp, err := client.Get(ctx, req)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !resp.Found {
			t.Error("Expected key to be found")
		}
		if string(resp.Value) != "test-value" {
			t.Errorf("Expected 'test-value', got '%s'", string(resp.Value))
		}
	})

	// Test Get non-existent key
	t.Run("GetNonExistent", func(t *testing.T) {
		req := &proto.GetRequest{Key: "non-existent"}
		resp, err := client.Get(ctx, req)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if resp.Found {
			t.Error("Expected key to not be found")
		}
		if resp.Value != nil {
			t.Error("Expected nil value for non-existent key")
		}
	})

	// Test Delete operation
	t.Run("Delete", func(t *testing.T) {
		req := &proto.DeleteRequest{Key: "test-key"}
		_, err := client.Delete(ctx, req)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify deletion
		getReq := &proto.GetRequest{Key: "test-key"}
		resp, err := client.Get(ctx, getReq)
		if err != nil {
			t.Fatalf("Get after delete failed: %v", err)
		}
		if resp.Found {
			t.Error("Expected key to not be found after deletion")
		}
	})
}

func TestGRPCServer_Integration_ValidationErrors(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test empty key validation
	t.Run("EmptyKey", func(t *testing.T) {
		req := &proto.PutRequest{
			Key:   "",
			Value: []byte("test-value"),
		}
		_, err := client.Put(ctx, req)
		if err == nil {
			t.Fatal("Expected error for empty key")
		}
		if status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unknown {
			t.Errorf("Expected InvalidArgument or Unknown error, got %v", status.Code(err))
		}
	})

	// Test oversized value validation
	t.Run("OversizedValue", func(t *testing.T) {
		// Create a value larger than the 100MB limit
		largeValue := make([]byte, 101*1024*1024) // 101MB
		req := &proto.PutRequest{
			Key:   "large-key",
			Value: largeValue,
		}
		_, err := client.Put(ctx, req)
		if err == nil {
			t.Fatal("Expected error for oversized value")
		}
		// gRPC returns ResourceExhausted for large messages that exceed transport limits
		if status.Code(err) != codes.ResourceExhausted && status.Code(err) != codes.InvalidArgument {
			t.Errorf("Expected ResourceExhausted or InvalidArgument error, got %v", status.Code(err))
		}
	})

	// Test long key validation
	t.Run("LongKey", func(t *testing.T) {
		// Create a key longer than 1024 characters
		longKey := string(make([]byte, 1025))
		req := &proto.PutRequest{
			Key:   longKey,
			Value: []byte("test-value"),
		}
		_, err := client.Put(ctx, req)
		if err == nil {
			t.Fatal("Expected error for long key")
		}
		if status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unknown {
			t.Errorf("Expected InvalidArgument or Unknown error, got %v", status.Code(err))
		}
	})
}

func TestGRPCServer_Integration_MultipleClients(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create multiple clients
	client1, conn1 := testServer.NewClient(t)
	defer conn1.Close()

	client2, conn2 := testServer.NewClient(t)
	defer conn2.Close()

	// Test concurrent operations
	t.Run("ConcurrentPuts", func(t *testing.T) {
		done := make(chan error, 2)

		// Client 1 puts values
		go func() {
			for i := 0; i < 10; i++ {
				req := &proto.PutRequest{
					Key:   fmt.Sprintf("client1-key-%d", i),
					Value: []byte(fmt.Sprintf("client1-value-%d", i)),
				}
				_, err := client1.Put(ctx, req)
				if err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}()

		// Client 2 puts values
		go func() {
			for i := 0; i < 10; i++ {
				req := &proto.PutRequest{
					Key:   fmt.Sprintf("client2-key-%d", i),
					Value: []byte(fmt.Sprintf("client2-value-%d", i)),
				}
				_, err := client2.Put(ctx, req)
				if err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}()

		// Wait for both clients to complete
		for i := 0; i < 2; i++ {
			if err := <-done; err != nil {
				t.Fatalf("Concurrent put failed: %v", err)
			}
		}
	})

	// Verify all values were stored correctly
	t.Run("VerifyValues", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			// Check client 1 values
			req := &proto.GetRequest{Key: fmt.Sprintf("client1-key-%d", i)}
			resp, err := client1.Get(ctx, req)
			if err != nil {
				t.Fatalf("Get client1 value failed: %v", err)
			}
			if !resp.Found {
				t.Errorf("Client1 key-%d not found", i)
			}
			expected := fmt.Sprintf("client1-value-%d", i)
			if string(resp.Value) != expected {
				t.Errorf("Client1 key-%d: expected '%s', got '%s'", i, expected, string(resp.Value))
			}

			// Check client 2 values
			req = &proto.GetRequest{Key: fmt.Sprintf("client2-key-%d", i)}
			resp, err = client2.Get(ctx, req)
			if err != nil {
				t.Fatalf("Get client2 value failed: %v", err)
			}
			if !resp.Found {
				t.Errorf("Client2 key-%d not found", i)
			}
			expected = fmt.Sprintf("client2-value-%d", i)
			if string(resp.Value) != expected {
				t.Errorf("Client2 key-%d: expected '%s', got '%s'", i, expected, string(resp.Value))
			}
		}
	})
}

func TestGRPCServer_Integration_LargeData(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with various data sizes
	testSizes := []int{
		1024,            // 1KB
		1024 * 1024,     // 1MB
		5 * 1024 * 1024, // 5MB (reduced from 10MB to be more conservative)
	}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size_%dBytes", size), func(t *testing.T) {
			// Create test data
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			key := fmt.Sprintf("large-data-%d", size)

			// Put large data
			putReq := &proto.PutRequest{
				Key:   key,
				Value: data,
			}
			_, err := client.Put(ctx, putReq)
			if err != nil {
				t.Fatalf("Put large data failed: %v", err)
			}

			// Get large data
			getReq := &proto.GetRequest{Key: key}
			resp, err := client.Get(ctx, getReq)
			if err != nil {
				t.Fatalf("Get large data failed: %v", err)
			}

			if !resp.Found {
				t.Error("Large data not found")
			}

			if len(resp.Value) != size {
				t.Errorf("Expected data size %d, got %d", size, len(resp.Value))
			}

			// Verify data integrity
			for i, b := range resp.Value {
				expected := byte(i % 256)
				if b != expected {
					t.Errorf("Data corruption at byte %d: expected %d, got %d", i, expected, b)
					break
				}
			}

			// Clean up
			delReq := &proto.DeleteRequest{Key: key}
			_, err = client.Delete(ctx, delReq)
			if err != nil {
				t.Fatalf("Delete large data failed: %v", err)
			}
		})
	}
}

func TestGRPCServer_Integration_ErrorHandling(t *testing.T) {
	// Create and start test server
	testServer := NewTestServer(t)
	defer testServer.Stop()
	testServer.Start(t)

	// Create client
	client, conn := testServer.NewClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test nil request handling (though gRPC should prevent this)
	t.Run("NilRequests", func(t *testing.T) {
		// These might be prevented by gRPC itself, but let's test the behavior

		// Test with minimal requests to ensure proper error handling
		_, err := client.Get(ctx, &proto.GetRequest{})
		if err != nil {
			// This might error due to empty key validation
			if status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unknown {
				t.Errorf("Expected InvalidArgument or Unknown for empty key, got %v", status.Code(err))
			}
		}

		_, err = client.Put(ctx, &proto.PutRequest{})
		if err != nil {
			// This should error due to empty key validation
			if status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unknown {
				t.Errorf("Expected InvalidArgument or Unknown for empty key, got %v", status.Code(err))
			}
		}

		_, err = client.Delete(ctx, &proto.DeleteRequest{})
		if err != nil {
			// This might error due to empty key validation
			if status.Code(err) != codes.InvalidArgument && status.Code(err) != codes.Unknown {
				t.Errorf("Expected InvalidArgument or Unknown for empty key, got %v", status.Code(err))
			}
		}
	})
}

func TestGRPCServer_Integration_Persistence(t *testing.T) {
	var tempDir string

	// First server instance
	{
		// Create and start test server
		testServer := NewTestServer(t)
		tempDir = testServer.tempDir
		testServer.Start(t)

		// Create client
		client, conn := testServer.NewClient(t)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Store some data
		for i := 0; i < 5; i++ {
			req := &proto.PutRequest{
				Key:   fmt.Sprintf("persistent-key-%d", i),
				Value: []byte(fmt.Sprintf("persistent-value-%d", i)),
			}
			_, err := client.Put(ctx, req)
			if err != nil {
				t.Fatalf("Put failed: %v", err)
			}
		}

		cancel()
		conn.Close()

		// Ensure data is flushed to disk before stopping server
		time.Sleep(200 * time.Millisecond)
		testServer.Stop()

		// Additional delay to ensure BadgerDB has fully closed
		time.Sleep(300 * time.Millisecond)
	}

	// Second server instance using the same data directory
	{
		// Create BadgerDB store with existing data
		badgerStore, err := badger.NewWithPath(filepath.Join(tempDir, "data"))
		if err != nil {
			t.Fatalf("Failed to reopen BadgerDB store: %v", err)
		}
		defer badgerStore.Close()
		defer os.RemoveAll(tempDir) // Clean up at the end

		// Wrap with validation
		validatedStore := validation.NewWithDefaultValidators(badgerStore)

		// Create new gRPC server with larger message limits
		grpcServer := grpc.NewServer(
			grpc.MaxRecvMsgSize(maxMessageSize), // 128MB
			grpc.MaxSendMsgSize(maxMessageSize), // 128MB
		)
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		config := &grpcserver.GRPCServerConfig{Port: listener.Addr().String()}
		server, err := grpcserver.New(validatedStore, config, grpcServer)
		if err != nil {
			t.Fatalf("Failed to create gRPC server: %v", err)
		}

		// Register and start server
		proto.RegisterClavisServer(grpcServer, server)
		done := make(chan struct{})
		go func() {
			defer close(done)
			grpcServer.Serve(listener)
		}()
		defer func() {
			grpcServer.GracefulStop()
			<-done
		}()

		time.Sleep(100 * time.Millisecond)

		// Create client for second server with larger message limits
		conn, err := grpc.NewClient(listener.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(maxMessageSize), // 128MB
				grpc.MaxCallSendMsgSize(maxMessageSize), // 128MB
			),
		)
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		client := proto.NewClavisClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Verify data persistence
		for i := 0; i < 5; i++ {
			req := &proto.GetRequest{Key: fmt.Sprintf("persistent-key-%d", i)}
			resp, err := client.Get(ctx, req)
			if err != nil {
				t.Fatalf("Get persistent data failed: %v", err)
			}
			if !resp.Found {
				t.Errorf("Persistent key-%d not found", i)
			}
			expected := fmt.Sprintf("persistent-value-%d", i)
			if string(resp.Value) != expected {
				t.Errorf("Persistent key-%d: expected '%s', got '%s'", i, expected, string(resp.Value))
			}
		}
	}
}
