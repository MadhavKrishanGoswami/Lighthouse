// Package client provides functionality to connect to the gRPC server
package client

import (
	"log"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StartClient tries to connect to the gRPC server and waits until it's available.

// StartClient waits for the server to be ready using the new grpc.NewClient API.
func StartClient() (orchestrator.HostAgentServiceClient, *grpc.ClientConn, error) {
	addr := "localhost:50051"
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var conn *grpc.ClientConn
	var err error

	for {
		conn, err = grpc.NewClient(addr, opts...)
		if err == nil {
			break
		}
		log.Printf("Waiting for gRPC server at %s... retrying in 2s", addr)
		time.Sleep(2 * time.Second)
	}

	client := orchestrator.NewHostAgentServiceClient(conn)
	log.Printf("gRPC client connected to orchestrator at %s", addr)
	return client, conn, nil
}
