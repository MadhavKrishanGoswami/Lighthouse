package main

func main() {
	// gRPC server setup
	gRPCServer := NewGRPCServer(":9000")
	// Start the gRPC server
	gRPCServer.Run()
}
