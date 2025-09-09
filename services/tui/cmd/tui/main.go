// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/client"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/config"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/ui"
)

func main() {
	// Configuration loading
	cfg := config.MustLoad()
	// Initialize gRPC client
	c, clientConn, err := client.StartClient(*cfg)
	if err != nil {
		log.Fatalf("gRPC client init failed: %v", err)
	}

	// Create TUI app with client
	app := ui.NewApp(c, clientConn)
	defer app.Close()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		app.Stop()
	}()

	// Run TUI app (blocks until exit)
	if err := app.Run(); err != nil {
		log.Fatalf("TUI app failed: %v", err)
	}
}
