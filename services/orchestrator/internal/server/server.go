// Package server provides a gRPC server implementation for handling requests.
package server

import (
	"context"
	"log"
	"net"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"google.golang.org/grpc"
)

type server struct {
	orchestrator.UnimplementedHostAgentServiceServer
}

func newServer() *server {
	return &server{}
}

func (s *server) RegisterHost(ctx context.Context, req *orchestrator.RegisterHostRequest) (*orchestrator.RegisterHostResponse, error) {
	// Implement the logic to register a host here
	log.Printf("Received request to register host: %s", req.Host)

	// For now, we just return a success response
	return &orchestrator.RegisterHostResponse{
		Success: true,
		Message: "Host registered successfully",
	}, nil
}

func StartServer() {
	// Start the gRPC server until it is stopped
	log.Println("Starting gRPC server...")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	// Listen on port 50051 for incoming gRPC requests
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Create a new gRPC server instance

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	orchestrator.RegisterHostAgentServiceServer(grpcServer, newServer())

	// Log the address where the server is listening
	log.Printf("gRPC server is listening on %s", lis.Addr().String())

	// Gracefully shutdown the server when it is stopped
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("gRPC server stopped gracefully")
	// Close the listener
	lis.Close()
	log.Println("Listener closed")
	log.Println("Server shutdown complete")
}
