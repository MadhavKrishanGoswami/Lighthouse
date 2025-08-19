// gRPC server setup and handler wiring
package client

import (
	"log"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartClient() {
	// This function will be used to start the gRPC client
	// It will connect to the server and perform operations
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// check if the connection is successful
	client := orchestrator.NewHostAgentServiceClient(conn)
	if client == nil {
		log.Fatalf("failed to create client: %v", err)
	}
	log.Println("gRPC client connected to server at localhost:50051")
}
