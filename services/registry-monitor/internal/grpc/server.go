// Package monitorServer implements the gRPC server for the registry-monitor service.
package monitorServer

import (
	"context"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/registry-monitor/internal/monitor"
)

type server struct {
	registry_monitor.UnimplementedRegistryMonitorServiceServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) CheckUpdates(ctx context.Context, req *registry_monitor.CheckUpdatesRequest) (*registry_monitor.CheckUpdatesResponse, error) {
	// Implement the logic to update the watchlist here
	return monitor.Monitor(req)
}
