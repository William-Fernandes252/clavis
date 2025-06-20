package main

import (
	"log"

	proto "github.com/William-Fernandes252/clavis/internal/server/grpc"
	"github.com/William-Fernandes252/clavis/internal/store/badger"
	"google.golang.org/grpc"
)

const (
	port     = ":50051"
	dataPath = "./data"
)

func main() {
	// Initialize storage
	kvStore, err := badger.NewWithPath(dataPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		if err := kvStore.Close(); err != nil {
			log.Printf("Failed to close storage: %v", err)
		}
	}()

	// Create the gRPC server
	grpcServer := grpc.NewServer()

	server, err := proto.New(kvStore, &proto.GRPCServerConfig{Port: port}, grpcServer)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	if err := server.Start(func() {
		log.Printf("Server is running on %s", port)
	}); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
