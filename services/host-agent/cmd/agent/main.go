package main

import (
	"context"
	"log"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/agent"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/client"
	dockerclient "github.com/moby/moby/client"
)

func main() {
	// This is the main entry point for the application.
	// The actual functionality is implemented in the agent package.
	// You can start the agent, monitor containers, or register with an orchestrator here.

	gRPCClient, clientConn, err := client.StartClient()

	// Gracefull shudown of the gRPC client connection
	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Printf("failed to close gRPC client connection: %v", err)
		} else {
			log.Println("gRPC client connection closed successfully")
		}
	}()
	if err != nil {
		panic(err)
	}
	// You can now use the client `c` to interact with the gRPC server.
	ctx := context.Background()
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	// Register the agent with the orchestrator

	if err := agent.RegisterAgent(cli, ctx, gRPCClient); err != nil {
		log.Fatalf("failed to register agent: %v", err)
	}
	log.Println("Agent registered successfully with the orchestrator")
}
