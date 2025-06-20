package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/William-Fernandes252/clavis/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	client := proto.NewClavisClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch os.Args[1] {
	case "put":
		_, err := client.Put(ctx, &proto.PutRequest{
			Key:   os.Args[2],
			Value: []byte(os.Args[3]),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Put successful")

	case "get":
		resp, err := client.Get(ctx, &proto.GetRequest{Key: os.Args[2]})
		if err != nil {
			log.Fatal(err)
		}
		if resp.Found {
			log.Printf("Value: %s", string(resp.Value))
		} else {
			log.Println("Key not found")
		}

	case "delete":
		_, err := client.Delete(ctx, &proto.DeleteRequest{Key: os.Args[2]})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Delete successful")

	default:
		log.Fatal("Unknown command. Usage: client [put|get|delete] [key] [value]?")
	}
}
