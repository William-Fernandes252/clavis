package proto

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/William-Fernandes252/clavis/api/proto"
	"github.com/William-Fernandes252/clavis/internal/server"
	"github.com/William-Fernandes252/clavis/internal/store"
	"google.golang.org/grpc"
)

// GRPCServerConfig defines the configuration for the gRPC server.
type GRPCServerConfig struct {
	Port string
}

var DefaultConfig = GRPCServerConfig{
	Port: ":50051",
}

// GRPCServer implements the server.Server interface for gRPC.
type GRPCServer struct {
	proto.UnimplementedClavisServer
	store  store.Store
	config *GRPCServerConfig
	server *grpc.Server
}

// New creates a new instance of GRPCServer with the provided store, configuration, and gRPC server.
func New(store store.Store, config *GRPCServerConfig, server *grpc.Server) (*GRPCServer, error) {
	return &GRPCServer{
		store:  store,
		config: config,
		server: server,
	}, nil
}

// Get retrieves the value associated with the key from the store.
func (s *GRPCServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	value, found, err := s.store.Get(req.Key)
	if err != nil {
		return nil, err
	}
	return &proto.GetResponse{Value: value, Found: found}, nil
}

// Put stores the value associated with the key in the store.
func (s *GRPCServer) Put(ctx context.Context, req *proto.PutRequest) (*proto.PutResponse, error) {
	if err := s.store.Put(req.Key, req.Value); err != nil {
		return nil, err
	}
	return &proto.PutResponse{}, nil
}

// Delete removes the key-value pair associated with the key from the store.
func (s *GRPCServer) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	if err := s.store.Delete(req.Key); err != nil {
		return nil, err
	}
	return &proto.DeleteResponse{}, nil
}

// Start initializes the gRPC server and starts listening for incoming connections.
// It also registers the server and sets up a shutdown handler.
// If any callbacks are provided, they will be executed after the server starts.
// The first callback, if provided, is executed immediately after the server starts.
func (s *GRPCServer) Start(callbacks ...func()) error {
	listener, err := s.listen(s.config.Port)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.register()

	go s.Shutdown()

	if len(callbacks) > 0 && callbacks[0] != nil {
		callbacks[0]()
	}

	if err := s.server.Serve(listener); err != nil {
		return err
	}

	return nil
}

func (s *GRPCServer) listen(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func (s *GRPCServer) register() {
	proto.RegisterClavisServer(s.server, s)
}

// Shutdown gracefully stops the gRPC server when a termination signal is received.
// It listens for SIGINT and SIGTERM signals and stops the server gracefully.
func (s *GRPCServer) Shutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down server...") // TODO: Use a logger instead of fmt
	s.server.GracefulStop()
}

// GetStore returns the store associated with the gRPC server.
func (s *GRPCServer) GetStore() (store.Store, error) {
	return s.store, nil
}

var _ server.Server = (*GRPCServer)(nil)
