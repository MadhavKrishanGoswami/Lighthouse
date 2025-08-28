package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/agent"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/client"
	dockerclient "github.com/moby/moby/client"
)

func main() {
	// --- 1. Start gRPC client and wait until orchestrator server is ready ---
	gRPCClient, clientConn, err := client.StartClient()
	if err != nil {
		log.Fatalf("Failed to start gRPC client: %v", err)
	}
	defer func() {
		if clientConn != nil {
			if err := clientConn.Close(); err != nil {
				log.Printf("Failed to close gRPC client connection: %v", err)
			} else {
				log.Println("gRPC client connection closed successfully")
			}
		}
	}()
	log.Println("gRPC client connected to orchestrator.")

	// --- 2. Docker client setup ---
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// --- 3. Context for all goroutines ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- 4. Register agent with orchestrator ---
	regCtx, regCancel := context.WithTimeout(ctx, 10*time.Second)
	defer regCancel()
	if err := agent.RegisterAgent(cli, regCtx, gRPCClient); err != nil {
		log.Fatalf("Failed to register agent: %v", err)
	}
	log.Println("Agent registered successfully with orchestrator.")

	// --- 5. WaitGroup for goroutines ---
	var wg sync.WaitGroup

	// --- 6. Heartbeat goroutine ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := agent.Heartbeat(cli, ctx, gRPCClient); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
				} else {
					log.Println("Heartbeat sent successfully.")
				}
			case <-ctx.Done():
				log.Println("Stopping heartbeat loop.")
				return
			}
		}
	}()

	// --- 7. UpdateContainer gRPC stream goroutine ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := agent.UpdateContainerStream(cli, ctx, gRPCClient); err != nil && ctx.Err() == nil {
			log.Printf("Failed to start UpdateContainer stream: %v", err)
		}
	}()

	// --- 8. Handle OS signals for graceful shutdown ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	log.Println("Termination signal received. Shutting down gracefully...")
	cancel()  // cancel context to stop goroutines
	wg.Wait() // wait for all goroutines to finish
	log.Println("All goroutines stopped. Agent shutdown complete.")
}
