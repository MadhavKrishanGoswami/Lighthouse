package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/config"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"

	// Import your service implementations
	agentserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/agent"
	registryserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/registry-monitor"

	// Import your generated proto definitions
	agentpb "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	registrypb "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.MustLoad()

	// Initialize database connection
	ctx := context.Background()
	dbConn, err := pgx.Connect(ctx, cfg.DataBaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close(ctx)
	queries := db.New(dbConn)

	// --- Server Startup Logic ---
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a single gRPC server instance
	grpcServer := grpc.NewServer()

	// Create instances of your service implementations
	agentServer := agentserver.NewServer(queries)
	registryServer := registryserver.NewServer(queries)

	// Register BOTH services with the single gRPC server
	agentpb.RegisterHostAgentServiceServer(grpcServer, agentServer)
	registrypb.RegisterRegistryMonitorServiceServer(grpcServer, registryServer)

	// --- Graceful Shutdown Logic ---

	// Start the server in a separate goroutine
	go func() {
		log.Printf("gRPC server listening on %s", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	// Listen for SIGINT (Ctrl+C) and SIGTERM (sent by Docker, Kubernetes, etc.)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-quit
	log.Println("Shutting down gRPC server...")

	// Gracefully stop the server. This will wait for existing connections to finish.
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped gracefully.")
}
