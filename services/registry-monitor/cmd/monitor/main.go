package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	// 1. ADD THIS IMPORT for the generated protobuf code
	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	monitorServer "github.com/MadhavKrishanGoswami/Lighthouse/services/registry-monitor/internal/grpc"
	"google.golang.org/grpc"
)

func main() {
	// --- Server Startup Logic ---
	lis, err := net.Listen("tcp", "0.0.0.0:50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// 2. CREATE and REGISTER your service implementation
	// Create an instance of your service.
	monitorServiceServer := monitorServer.NewServer()
	// Register that service with the main gRPC server.
	registry_monitor.RegisterRegistryMonitorServiceServer(grpcServer, monitorServiceServer)

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
