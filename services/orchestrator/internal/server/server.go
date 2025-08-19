// Package server provides a gRPC server implementation for handling requests.
package server

import (
	"log"
	"net"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"google.golang.org/grpc"
)

type server struct {
	orchestrator.UnimplementedHostAgentServiceServer
}

// constructor
func newServer() *server {
	return &server{}
}

func StartServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	orchestrator.RegisterHostAgentServiceServer(grpcServer, newServer())

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
