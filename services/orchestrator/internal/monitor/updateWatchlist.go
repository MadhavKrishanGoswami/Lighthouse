package monitor

import (
	"context"
	"log"
	"strings"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
)

func UpdateWatchlistinRegistryMonitor(ctx context.Context, grpcClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries) error {
	// Get containers from the database where watchlist is true
	containers, err := queries.GetallContainersWhereWatched(ctx)
	if err != nil {
		log.Printf("Failed to get containers from database: %v", err)
	}
	if len(containers) == 0 {
		log.Println("No containers found in the watchlist.")
		return nil
	}
	// Send containers to the registry monitor service
	var containerInfos []*registry_monitor.ImageInfo
	for _, c := range containers {
		// Extract repository name from image (assuming image is in the format "repository:tag" or "repository")
		repository := c.Name
		if idx := strings.Index(c.Image, ":"); idx != -1 {
			repository = c.Image[:idx]
		}
		// also get tag after :
		tagfind := c.Name
		if idx := strings.Index(c.Image, ":"); idx != -1 {
			tagfind = c.Image[idx+1:]
		} else {
			tagfind = "latest" // default tag
		}
		// Create a ContainerInfo message
		containerInfo := &registry_monitor.ImageInfo{
			Repository: repository,
			Tag:        tagfind,
		}
		containerInfos = append(containerInfos, containerInfo)
	}
	req := &registry_monitor.UpdateWatchlistRequest{
		Images: containerInfos,
	}
	_, err = grpcClient.UpdateWatchlist(ctx, req)
	if err != nil {
		log.Printf("Failed to update watchlist in registry monitor: %v", err)
		return err
	}
	log.Printf("Successfully updated watchlist in registry monitor with %d containers.", len(containerInfos))
	return nil
}
