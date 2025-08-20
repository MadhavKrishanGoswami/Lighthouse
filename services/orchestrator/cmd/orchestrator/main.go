package main

import (
	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/config"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/server"
)

func main() {
	// Load the configuration
	cfg := config.MustLoad()
	// Start the gRPC server
	server.StartServer(cfg) // Use the correct package path for Config
}
