// gRPC server setup and handler wiring
package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func SetupServer() {
	// This fuction will setup and listen for gRPC requests.
	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("Failed to listen on port 50051: %v\n", err)
		return
	}
	fmt.Println("gRPC server is listening on port 50051...")
	server := grpc.NewServer()
	// Register your service handlers here
}
