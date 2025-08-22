package main

import (
	"context"

	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/config"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/server"
	"github.com/jackc/pgx/v5"
)

func main() {
	// Load the configuration
	cfg := config.MustLoad()

	// Initialize the database connection
	ctx := context.Background()
	dbConn, err := pgx.Connect(ctx, cfg.DataBaseURL)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer dbConn.Close(ctx) // Ensure the database connection is closed when done

	queries := db.New(dbConn) // Create a new Queries instance with the database connection

	// Create a new database instance

	// Start the gRPC server
	server.StartServer(cfg, queries) // Use the correct package path for Config
}
