// Package monitor provides functionalities for monitoring container updates.
package monitor

import (
	"context"
	"log"
	"strings"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
)

// CheckForUpdates queries the database for watched containers and asks the
// registry-monitor service to check if updates are available for them.
func CheckForUpdates(ctx context.Context, grpcClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries) (registry_monitor.CheckUpdatesResponse, error) {
	// Get containers from the database where watchlist is true.
	containers, err := queries.GetallContainersWhereWatched(ctx)
	if err != nil {
		log.Printf("Failed to get containers from database: %v", err)
		// Return the error to the caller instead of continuing with a nil slice.
		return registry_monitor.CheckUpdatesResponse{}, err
	}

	if len(containers) == 0 {
		log.Println("No containers in watchlist")
		return registry_monitor.CheckUpdatesResponse{}, nil
	}

	// Prepare the request for the registry monitor service.
	var containerInfos []*registry_monitor.ImageInfo
	for _, c := range containers {
		// Correctly parse the repository and tag from the image string.
		// An image string can be in formats like:
		// - "ubuntu" (implies "docker.io/library/ubuntu:latest")
		// - "ubuntu:22.04"
		// - "myregistry:5000/my-app:1.0"
		imageStr := c.Image
		repository := imageStr
		tag := "latest" // Assume "latest" if no tag is specified.

		// A tag is present if a colon exists after the last slash.
		// This correctly handles registry URLs with port numbers (e.g., myregistry:5000/...).
		lastColon := strings.LastIndex(imageStr, ":")
		lastSlash := strings.LastIndex(imageStr, "/")

		if lastColon > lastSlash {
			repository = imageStr[:lastColon]
			tag = imageStr[lastColon+1:]
		}

		// Create an ImageInfo message.
		// The c.Digest field is passed directly from the database.
		// As you noted, if the digest is empty in the database, it will be empty
		// in this request. The logic for populating the initial digest should be
		// handled when the container is first added to the database.
		log.Printf("Check image repository=%s tag=%s", repository, tag)
		imageInfo := &registry_monitor.ImageInfo{
			ContainerUid: c.ContainerUid,
			Repository:   repository,
			Tag:          tag,
		}
		containerInfos = append(containerInfos, imageInfo)
	}

	req := &registry_monitor.CheckUpdatesRequest{
		Images: containerInfos,
	}

	log.Printf("Checking for updates for %d containers", len(containerInfos))
	resp, err := grpcClient.CheckUpdates(ctx, req)
	if err != nil {
		log.Printf("gRPC call to CheckUpdates failed: %v", err)
		return registry_monitor.CheckUpdatesResponse{}, err
	}

	log.Printf("%d containers have updates", len(resp.ImagestoUpdate))
	return *resp, nil
}
