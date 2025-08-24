// Package registryclient provides functionality to connect to the gRPC server
package registryclient

import (
	"log"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartClient() (registry_monitor.RegistryMonitorServiceClient, *grpc.ClientConn, error) {
	// This function will be used to start the gRPC client
	// It will connect to the server and perform operations
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient("localhost:50052", opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, nil, err
	}

	// check if the connection is successful
	client := registry_monitor.NewRegistryMonitorServiceClient(conn)
	if client == nil {
		log.Fatalf("failed to create client: %v", err)
		return nil, nil, err
	}
	log.Println("gRPC client connected to orchestrator at localhost:50051")

	// Return the client to be used for further operations
	return client, conn, nil
}
