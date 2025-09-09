package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	// Internal package imports
	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/config"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	agentserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/agent"
	registryclient "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/registry-monitor"
	tuiserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/tui"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/monitor"

	// Proto definitions
	agentpb "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	// TUI proto package for registration
	"github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"

	// External dependencies
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

func main() {
	// --- 1. Configuration Loading ---
	// Load configuration from environment variables or a config file.
	// The application will exit if the configuration is invalid.
	cfg := config.MustLoad()
	log.Println("Configuration loaded successfully.")

	// --- 2. Database Connection ---
	// Establish a connection to the PostgreSQL database.
	// The connection is closed gracefully when the main function exits.
	ctx := context.Background()
	dbConn, err := pgx.Connect(ctx, cfg.DataBaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close(ctx)
	queries := db.New(dbConn)
	log.Println("Database connection established.")

	// --- 3. Registry Monitor Client Connection ---
	registryMonitorClient, clientConn, err := registryclient.StartClient()
	if err != nil {
		log.Fatalf("Failed to start registry monitor client: %v", err)
	}
	defer func() {
		if clientConn != nil {
			log.Println("Closing gRPC client connection to Registry Monitor...")
			if err := clientConn.Close(); err != nil {
				log.Printf("Error closing gRPC client connection: %v", err)
			}
		}
	}()
	log.Println("gRPC client for Registry Monitor started.")

	// --- 4. gRPC Server Setup ---
	// Set up the main gRPC server that will host our services.
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("Failed to listen on address %s: %v", cfg.Addr, err)
	}
	grpcServer := grpc.NewServer()

	// --- 5. Service Registration ---
	agentServer := agentserver.NewServer(queries)
	agentpb.RegisterHostAgentServiceServer(grpcServer, agentServer)
	log.Println("HostAgentService registered.")

	// TUI server (for logs + snapshots)
	// Initialize and register the TUI gRPC service so that TUI clients can connect.
	// Without this registration clients receive Unimplemented errors.
	// We also hook the standard logger so logs are streamed to connected TUI log streams.
	{
		// local scope to avoid accidental reuse of tuiSrv below; if needed later move outside block
		tuiSrv := tuiserver.NewServer(queries)
		tuiSrv.HookStandardLogger()
		// Register service with gRPC server
		tui.RegisterTUIServiceServer(grpcServer, tuiSrv)
		log.Println("TUIService registered and logging hook installed.")
	}

	// --- 6. Server Start & Graceful Shutdown ---
	// Start the server in a background goroutine so it doesn't block.
	go func() {
		log.Printf("gRPC server starting on %s", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()
	// -----starting cron job for monitoring after host agentserver is connected-----
	go func() {
		log.Println("Starting cron job for monitoring...")
		// Wire dependencies so SetCronTime can restart correctly
		monitor.SetRuntimeDeps(registryMonitorClient, queries, agentServer)
		monitor.SetCronTimeInHours(1) // Set to 1 hour, can be made configurable
		monitor.StartCronJob(registryMonitorClient, queries, agentServer)
	}()
	// Wait for a shutdown signal (e.g., Ctrl+C).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, initiating graceful shutdown...")

	// Gracefully stop the server. This allows ongoing requests to complete.
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped gracefully.")
}
