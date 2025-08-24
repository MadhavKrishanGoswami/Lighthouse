package monitorServer

import (
	"context"
	"log"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
)

type server struct {
	registry_monitor.UnimplementedRegistryMonitorServiceServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) UpdateWatchlist(ctx context.Context, req *registry_monitor.UpdateWatchlistRequest) (*registry_monitor.WatchlistUpatedResponse, error) {
	// Implement the logic to update the watchlist here
	log.Printf("Received request to update watchlist: %v", req.Images)

	// Here, you would typically update the watchlist in your database or in-memory store.
	// For demonstration purposes, we'll just log the received watchlist.

	log.Printf("Watchlist updated successfully")

	return &registry_monitor.WatchlistUpatedResponse{
		Success: true,
		Message: "Watchlist updated successfully",
	}, nil
}
