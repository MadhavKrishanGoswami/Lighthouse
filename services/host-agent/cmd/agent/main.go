package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/agent"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/client"
	dockerclient "github.com/moby/moby/client"
)

func main() {
	// This is the main entry point for the application.
	// The actual functionality is implemented in the agent package.
	// You can start the agent, monitor containers, or register with an orchestrator here.

	gRPCClient, clientConn, err := client.StartClient()
	if err != nil {
		panic(err)
	}
	defer func() {
		if clientConn != nil {
			if err := clientConn.Close(); err != nil {
				log.Printf("failed to close gRPC client connection: %v", err)
			} else {
				log.Println("gRPC client connection closed successfully")
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	// Start Heart Brate to send periodic updates to the orchestrator
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := agent.Heartbeat(cli, ctx, gRPCClient); err != nil {
					log.Printf("failed to send heartbeat: %v", err)
				} else {
					log.Println("Heartbeat sent successfully")
				}
			case <-ctx.Done():
				log.Println("Stopping heartbeat loop")
				return
			}
		}
	}()
	// Wait for a termination signal (e.g., Ctrl+C) to gracefully shut Down the agent
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("Termination signal received. Shutting down gracefully...")
}
