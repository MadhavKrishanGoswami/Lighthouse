// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/client"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/data"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/ui"
)

func main() {
	// Initialize gRPC client
	gRPCClient, clientConn, err := client.StartClient()
	if err != nil {
		log.Fatalf("gRPC client init failed: %v", err)
	}
	defer clientConn.Close()

	// Create TUI app
	app := ui.NewApp()

	// Create data manager
	dataManager := data.NewDataManager(gRPCClient, app)

	// Start data streaming
	err = dataManager.StartDataStream()
	if err != nil {
		log.Fatalf("Failed to start data stream: %v", err)
	}
	defer dataManager.Stop()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		dataManager.Stop()
		app.Stop()
	}()

	// Run TUI app (blocks until exit)
	if err := app.Run(); err != nil {
		log.Fatalf("TUI app failed: %v", err)
	}
}
