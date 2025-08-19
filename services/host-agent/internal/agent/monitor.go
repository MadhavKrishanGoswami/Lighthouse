// Package agent provides functionality to monitor Docker containers.
package agent

import (
	"context"
	"fmt"

	"github.com/moby/moby/client"
)

// MonitorContainer monitors the status of a Docker containers and returns its status.
func MonitorContainer(cli *client.Client, ctx context.Context, containerID string) (string, error) {
	// Get container status
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	// Return the status of the container
	status := containerJSON.State.Status
	return status, nil
}
