// Package registryserver implements the gRPC server for the RegistryMonitorService.
package registryserver

import (
	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
)

type server struct {
	orchestrator.UnimplementedRegistryMonitorServiceServer
	db *db.Queries // Database queries instance
}

func NewServer(queries *db.Queries) *server {
	return &server{
		db: queries, // Initialize the database queries instance
	}
}
