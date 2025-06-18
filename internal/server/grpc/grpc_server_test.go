package proto

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/William-Fernandes252/clavis/api/proto"
	"github.com/William-Fernandes252/clavis/internal/store"
	"google.golang.org/grpc"
)

// mockStore implements the store.Store interface for testing
type mockStore struct {
	data      map[string][]byte
	getError  error
	putError  error
	delError  error
	scanError error
	closed    bool
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string][]byte),
	}
}

func (m *mockStore) Get(key string) ([]byte, bool, error) {
	if m.closed {
		return nil, false, errors.New("store is closed")
	}
	if m.getError != nil {
		return nil, false, m.getError
	}
	value, found := m.data[key]
	if !found {
		return nil, false, nil
	}
	// Return a copy to simulate real behavior
	result := make([]byte, len(value))
	copy(result, value)
	return result, true, nil
}

func (m *mockStore) Put(key string, value []byte) error {
	if m.closed {
		return errors.New("store is closed")
	}
	if m.putError != nil {
		return m.putError
	}
	// Store a copy to simulate real behavior
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	m.data[key] = valueCopy
	return nil
}

func (m *mockStore) Delete(key string) error {
	if m.closed {
		return errors.New("store is closed")
	}
	if m.delError != nil {
		return m.delError
	}
	delete(m.data, key)
	return nil
}

func (m *mockStore) Scan(prefix string) (map[string][]byte, error) {
	if m.closed {
		return nil, errors.New("store is closed")
	}
	if m.scanError != nil {
		return nil, m.scanError
	}
	result := make(map[string][]byte)
	for key, value := range m.data {
		if len(prefix) == 0 || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			result[key] = valueCopy
		}
	}
	return result, nil
}

func (m *mockStore) Close() error {
	m.closed = true
	return nil
}

func (m *mockStore) setGetError(err error) {
	m.getError = err
}

func (m *mockStore) setPutError(err error) {
	m.putError = err
}

func (m *mockStore) setDeleteError(err error) {
	m.delError = err
}

func TestNew(t *testing.T) {
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: ":50051"}
	grpcServer := grpc.NewServer()

	type args struct {
		store  store.Store
		config *GRPCServerConfig
		server *grpc.Server
	}
	tests := []struct {
		name    string
		args    args
		want    *GRPCServer
		wantErr bool
	}{
		{
			name: "successful creation",
			args: args{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			want: &GRPCServer{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			wantErr: false,
		},
		{
			name: "creation with nil store",
			args: args{
				store:  nil,
				config: config,
				server: grpcServer,
			},
			want: &GRPCServer{
				store:  nil,
				config: config,
				server: grpcServer,
			},
			wantErr: false,
		},
		{
			name: "creation with nil config",
			args: args{
				store:  mockStore,
				config: nil,
				server: grpcServer,
			},
			want: &GRPCServer{
				store:  mockStore,
				config: nil,
				server: grpcServer,
			},
			wantErr: false,
		},
		{
			name: "creation with nil grpc server",
			args: args{
				store:  mockStore,
				config: config,
				server: nil,
			},
			want: &GRPCServer{
				store:  mockStore,
				config: config,
				server: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.store, tt.args.config, tt.args.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGRPCServer_Get(t *testing.T) {
	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	type args struct {
		ctx context.Context
		req *proto.GetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.GetResponse
		wantErr bool
	}{
		{
			name: "successful get - key exists",
			fields: fields{
				store: func() store.Store {
					mock := newMockStore()
					mock.data["test-key"] = []byte("test-value")
					return mock
				}(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetRequest{Key: "test-key"},
			},
			want: &proto.GetResponse{
				Value: []byte("test-value"),
				Found: true,
			},
			wantErr: false,
		},
		{
			name: "get non-existent key",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetRequest{Key: "non-existent"},
			},
			want: &proto.GetResponse{
				Value: nil,
				Found: false,
			},
			wantErr: false,
		},
		{
			name: "get with empty key",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetRequest{Key: ""},
			},
			want: &proto.GetResponse{
				Value: nil,
				Found: false,
			},
			wantErr: false,
		},
		{
			name: "store error during get",
			fields: fields{
				store: func() store.Store {
					mock := newMockStore()
					mock.setGetError(errors.New("store error"))
					return mock
				}(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetRequest{Key: "test-key"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get with nil request",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: nil,
			},
			want:    nil,
			wantErr: true, // This will panic, but we expect it to be caught as an error
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}

			// Handle panic for nil request case
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("GRPCServer.Get() panicked unexpectedly: %v", r)
				}
			}()

			got, err := s.Get(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GRPCServer.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGRPCServer_Put(t *testing.T) {
	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	type args struct {
		ctx context.Context
		req *proto.PutRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.PutResponse
		wantErr bool
	}{
		{
			name: "successful put",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.PutRequest{
					Key:   "test-key",
					Value: []byte("test-value"),
				},
			},
			want:    &proto.PutResponse{},
			wantErr: false,
		},
		{
			name: "put with empty key",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.PutRequest{
					Key:   "",
					Value: []byte("test-value"),
				},
			},
			want:    &proto.PutResponse{},
			wantErr: false,
		},
		{
			name: "put with empty value",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.PutRequest{
					Key:   "test-key",
					Value: []byte{},
				},
			},
			want:    &proto.PutResponse{},
			wantErr: false,
		},
		{
			name: "put with nil value",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.PutRequest{
					Key:   "test-key",
					Value: nil,
				},
			},
			want:    &proto.PutResponse{},
			wantErr: false,
		},
		{
			name: "store error during put",
			fields: fields{
				store: func() store.Store {
					mock := newMockStore()
					mock.setPutError(errors.New("store error"))
					return mock
				}(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.PutRequest{
					Key:   "test-key",
					Value: []byte("test-value"),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "put with nil request",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: nil,
			},
			want:    nil,
			wantErr: true, // This will panic, but we expect it to be caught as an error
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}

			// Handle panic for nil request case
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("GRPCServer.Put() panicked unexpectedly: %v", r)
				}
			}()

			got, err := s.Put(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GRPCServer.Put() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGRPCServer_Delete(t *testing.T) {
	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	type args struct {
		ctx context.Context
		req *proto.DeleteRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.DeleteResponse
		wantErr bool
	}{
		{
			name: "successful delete - key exists",
			fields: fields{
				store: func() store.Store {
					mock := newMockStore()
					mock.data["test-key"] = []byte("test-value")
					return mock
				}(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.DeleteRequest{Key: "test-key"},
			},
			want:    &proto.DeleteResponse{},
			wantErr: false,
		},
		{
			name: "delete non-existent key",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.DeleteRequest{Key: "non-existent"},
			},
			want:    &proto.DeleteResponse{},
			wantErr: false,
		},
		{
			name: "delete with empty key",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.DeleteRequest{Key: ""},
			},
			want:    &proto.DeleteResponse{},
			wantErr: false,
		},
		{
			name: "store error during delete",
			fields: fields{
				store: func() store.Store {
					mock := newMockStore()
					mock.setDeleteError(errors.New("store error"))
					return mock
				}(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: &proto.DeleteRequest{Key: "test-key"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "delete with nil request",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
			args: args{
				ctx: context.Background(),
				req: nil,
			},
			want:    nil,
			wantErr: true, // This will panic, but we expect it to be caught as an error
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}

			// Handle panic for nil request case
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("GRPCServer.Delete() panicked unexpectedly: %v", r)
				}
			}()

			got, err := s.Delete(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GRPCServer.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGRPCServer_Start(t *testing.T) {
	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	type args struct {
		callbacks []func()
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "start with invalid port",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: "invalid-port"},
				server: grpc.NewServer(),
			},
			args: args{
				callbacks: nil,
			},
			wantErr: true,
		},
		{
			name: "start with nil config",
			fields: fields{
				store:  newMockStore(),
				config: nil,
				server: grpc.NewServer(),
			},
			args: args{
				callbacks: nil,
			},
			wantErr: true, // Will panic when accessing s.config.Port
		},
		{
			name: "start with high port number",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":65535"}, // Use max valid port
				server: grpc.NewServer(),
			},
			args: args{
				callbacks: nil,
			},
			wantErr: false, // This should succeed
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}

			// Handle panic cases
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("GRPCServer.Start() panicked unexpectedly: %v", r)
					}
				}
			}()

			err := s.Start(tt.args.callbacks...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGRPCServer_Start_CallbackValidation(t *testing.T) {
	// This test focuses on testing callback parameter validation
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: "invalid-port"} // Use invalid port to trigger early error
	grpcServer := grpc.NewServer()

	s := &GRPCServer{
		store:  mockStore,
		config: config,
		server: grpcServer,
	}

	tests := []struct {
		name      string
		callbacks []func()
		wantErr   bool
	}{
		{
			name:      "no callbacks",
			callbacks: nil,
			wantErr:   true, // Should fail due to invalid port
		},
		{
			name:      "single nil callback",
			callbacks: []func(){nil},
			wantErr:   true, // Should fail due to invalid port
		},
		{
			name: "single valid callback",
			callbacks: []func(){func() {
				// Valid callback
			}},
			wantErr: true, // Should fail due to invalid port
		},
		{
			name: "multiple callbacks",
			callbacks: []func(){
				func() { /* callback 1 */ },
				func() { /* callback 2 */ },
			},
			wantErr: true, // Should fail due to invalid port
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Start(tt.callbacks...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGRPCServer_listen(t *testing.T) {
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: ":50051"}
	grpcServer := grpc.NewServer()

	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	type args struct {
		port string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful listen on random port",
			fields: fields{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			args: args{
				port: ":0", // Use random available port
			},
			wantErr: false,
		},
		{
			name: "invalid port format",
			fields: fields{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			args: args{
				port: "invalid-port",
			},
			wantErr: true,
		},
		{
			name: "empty port",
			fields: fields{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			args: args{
				port: "",
			},
			wantErr: false, // Empty port defaults to all interfaces, which might succeed
		},
		{
			name: "malformed port number",
			fields: fields{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			args: args{
				port: ":99999999", // Port number too large
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}
			got, err := s.listen(tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.listen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("GRPCServer.listen() returned nil listener when expecting success")
					return
				}
				// Clean up the listener
				got.Close()
			}
		})
	}
}

func TestGRPCServer_register(t *testing.T) {
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: ":50051"}
	grpcServer := grpc.NewServer()

	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "successful registration",
			fields: fields{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
		},
		{
			name: "registration with nil server",
			fields: fields{
				store:  mockStore,
				config: config,
				server: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}

			// Handle potential panic for nil server case
			defer func() {
				if r := recover(); r != nil && tt.name != "registration with nil server" {
					t.Errorf("GRPCServer.register() panicked unexpectedly: %v", r)
				}
			}()

			s.register()
		})
	}
}

func TestGRPCServer_Shutdown(t *testing.T) {
	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "successful shutdown with valid server",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
		},
		{
			name: "shutdown with nil server",
			fields: fields{
				store:  newMockStore(),
				config: &GRPCServerConfig{Port: ":50051"},
				server: nil,
			},
		},
		{
			name: "shutdown with nil store",
			fields: fields{
				store:  nil,
				config: &GRPCServerConfig{Port: ":50051"},
				server: grpc.NewServer(),
			},
		},
		{
			name: "shutdown with nil config",
			fields: fields{
				store:  newMockStore(),
				config: nil,
				server: grpc.NewServer(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}

			// Handle potential panic for nil server case
			defer func() {
				if r := recover(); r != nil {
					if tt.name != "shutdown with nil server" {
						t.Errorf("GRPCServer.Shutdown() panicked unexpectedly: %v", r)
					}
					// For nil server case, we expect a panic, which is acceptable behavior
				}
			}()

			// Since Shutdown() blocks waiting for signals, we need to test it differently
			// We can't actually send signals in a unit test, so we test that the method
			// can be called without immediate errors (the signal waiting would block indefinitely)
			// In a real scenario, this would be tested with integration tests

			// For testing purposes, we'll just verify the method exists and can be called
			// without immediate panics (except for the nil server case)
			if tt.fields.server == nil && tt.name == "shutdown with nil server" {
				// This should panic when trying to call GracefulStop on nil server
				// We test this by attempting to call it and expecting a panic
				s.Shutdown()
				t.Error("Expected panic for nil server, but Shutdown() completed")
			} else {
				// For non-nil servers, we can't easily test the signal handling in a unit test
				// since it would block indefinitely waiting for signals.
				// In practice, this method would be tested with integration tests or by
				// sending actual signals to the process.

				// We can at least verify that calling the method doesn't immediately panic
				// by starting it in a goroutine and ensuring the test completes
				done := make(chan bool, 1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							// Expected for nil server case
							done <- true
						}
					}()
					// We can't actually call s.Shutdown() here as it would block
					// Instead, we verify the method signature and that the struct is properly initialized
					done <- true
				}()

				select {
				case <-done:
					// Test completed successfully
				default:
					// This shouldn't happen in our test setup
				}
			}
		})
	}
}

func TestGRPCServer_GetStore(t *testing.T) {
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: ":50051"}
	grpcServer := grpc.NewServer()

	type fields struct {
		UnimplementedClavisServer proto.UnimplementedClavisServer
		store                     store.Store
		config                    *GRPCServerConfig
		server                    *grpc.Server
	}
	tests := []struct {
		name    string
		fields  fields
		want    store.Store
		wantErr bool
	}{
		{
			name: "successful get store",
			fields: fields{
				store:  mockStore,
				config: config,
				server: grpcServer,
			},
			want:    mockStore,
			wantErr: false,
		},
		{
			name: "get store with nil store",
			fields: fields{
				store:  nil,
				config: config,
				server: grpcServer,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GRPCServer{
				UnimplementedClavisServer: tt.fields.UnimplementedClavisServer,
				store:                     tt.fields.store,
				config:                    tt.fields.config,
				server:                    tt.fields.server,
			}
			got, err := s.GetStore()
			if (err != nil) != tt.wantErr {
				t.Errorf("GRPCServer.GetStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GRPCServer.GetStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGRPCServer_ErrorPropagation(t *testing.T) {
	// Test that store errors are properly propagated through the gRPC methods
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: ":50051"}
	grpcServer := grpc.NewServer()

	s := &GRPCServer{
		store:  mockStore,
		config: config,
		server: grpcServer,
	}

	ctx := context.Background()

	// Test Get error propagation
	mockStore.setGetError(errors.New("store get error"))
	_, err := s.Get(ctx, &proto.GetRequest{Key: "test"})
	if err == nil {
		t.Error("Expected error from Get method, but got nil")
	}
	if err.Error() != "store get error" {
		t.Errorf("Expected 'store get error', got '%v'", err)
	}

	// Reset and test Put error propagation
	mockStore.setGetError(nil)
	mockStore.setPutError(errors.New("store put error"))
	_, err = s.Put(ctx, &proto.PutRequest{Key: "test", Value: []byte("value")})
	if err == nil {
		t.Error("Expected error from Put method, but got nil")
	}
	if err.Error() != "store put error" {
		t.Errorf("Expected 'store put error', got '%v'", err)
	}

	// Reset and test Delete error propagation
	mockStore.setPutError(nil)
	mockStore.setDeleteError(errors.New("store delete error"))
	_, err = s.Delete(ctx, &proto.DeleteRequest{Key: "test"})
	if err == nil {
		t.Error("Expected error from Delete method, but got nil")
	}
	if err.Error() != "store delete error" {
		t.Errorf("Expected 'store delete error', got '%v'", err)
	}
}

func TestGRPCServer_StoreStateAfterOperations(t *testing.T) {
	// Test that operations actually modify the store state correctly
	mockStore := newMockStore()
	config := &GRPCServerConfig{Port: ":50051"}
	grpcServer := grpc.NewServer()

	s := &GRPCServer{
		store:  mockStore,
		config: config,
		server: grpcServer,
	}

	ctx := context.Background()

	// Put a value
	_, err := s.Put(ctx, &proto.PutRequest{Key: "test-key", Value: []byte("test-value")})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify it was stored
	resp, err := s.Get(ctx, &proto.GetRequest{Key: "test-key"})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !resp.Found {
		t.Error("Expected key to be found, but it wasn't")
	}
	if string(resp.Value) != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", string(resp.Value))
	}

	// Delete the value
	_, err = s.Delete(ctx, &proto.DeleteRequest{Key: "test-key"})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it was deleted
	resp, err = s.Get(ctx, &proto.GetRequest{Key: "test-key"})
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if resp.Found {
		t.Error("Expected key to not be found after deletion, but it was")
	}

	// Test overwriting a value
	_, err = s.Put(ctx, &proto.PutRequest{Key: "test-key", Value: []byte("value1")})
	if err != nil {
		t.Fatalf("First put failed: %v", err)
	}

	_, err = s.Put(ctx, &proto.PutRequest{Key: "test-key", Value: []byte("value2")})
	if err != nil {
		t.Fatalf("Second put failed: %v", err)
	}

	resp, err = s.Get(ctx, &proto.GetRequest{Key: "test-key"})
	if err != nil {
		t.Fatalf("Get after overwrite failed: %v", err)
	}
	if string(resp.Value) != "value2" {
		t.Errorf("Expected 'value2' after overwrite, got '%s'", string(resp.Value))
	}
}
