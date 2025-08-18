// Handles Monitor Container logic
package agent

import (
	"context"
	"fmt"

	"github.com/moby/moby/client"
)

// MonitorContainer monitors the status of a Docker containers and returns its status.
func MonitorContainer(ctx context.Context, containerID string) (string, error) {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Get container status
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	// Return the status of the container
	status := containerJSON.State.Status
	return status, nil
}
