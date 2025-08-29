package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// UpdateContainer safely updates a container and rolls back on failure
func UpdateContainer(cli *dockerclient.Client, ctx context.Context, update *orchestrator.UpdateContainerCommand, stream orchestrator.HostAgentService_ConnectAgentStreamClient) error {
	log.Printf("Starting update for container %s -> image %s", update.ContainerUID, update.Image)

	// 1. Inspect the existing container to get its configuration
	inspect, err := cli.ContainerInspect(ctx, update.ContainerUID)
	if err != nil {
		log.Printf("Inspect failed for %s: %v", update.ContainerUID, err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Container not found: %v", err))
		return err
	}

	originalConfig := inspect.Config
	originalHostConfig := inspect.HostConfig
	originalName := strings.TrimPrefix(inspect.Name, "/")
	originalImage := inspect.Config.Image

	// Defer the rollback function to execute if any subsequent step fails
	var newContainerID string
	rollbackNeeded := true
	defer func() {
		if rollbackNeeded {
			rollbackChanges(cli, ctx, stream, update, originalConfig, originalHostConfig, originalName, newContainerID)
		}
	}()

	// 2. Pull the new Docker image
	if err := pullImage(cli, ctx, stream, update); err != nil {
		return err // Rollback will be triggered by defer
	}

	// 3. Stop and remove the old container
	if err := stopAndRemoveContainer(cli, ctx, stream, update, update.ContainerUID); err != nil {
		return err // Rollback will be triggered by defer
	}

	// 4. Create the new container
	newContainerID, err = createNewContainer(cli, ctx, stream, update, originalName, &inspect)
	if err != nil {
		return err // Rollback will be triggered by defer
	}

	// 5. Start the new container
	if err := startNewContainer(cli, ctx, stream, update, newContainerID); err != nil {
		return err // Rollback will be triggered by defer
	}

	// If all steps succeed, disable the rollback
	rollbackNeeded = false

	// 6. Send completion status and clean up the old image
	log.Printf("Update completed. New container ID: %s", newContainerID)
	sendStatus(stream, update, orchestrator.UpdateStatus_COMPLETED, fmt.Sprintf("Container updated successfully. New ID: %s", newContainerID))

	go cleanupOldImage(cli, context.Background(), originalImage)

	return nil
}

// rollbackChanges attempts to restore the original container if the update fails
func rollbackChanges(cli *dockerclient.Client, ctx context.Context, stream orchestrator.HostAgentService_ConnectAgentStreamClient, update *orchestrator.UpdateContainerCommand, originalConfig *container.Config, originalHostConfig *container.HostConfig, originalName, newContainerID string) {
	log.Printf("Starting rollback for %s", update.ContainerUID)
	sendStatus(stream, update, orchestrator.UpdateStatus_ROLLBACK, "Update failed, attempting to roll back.")

	// If a new container was created, try to remove it
	if newContainerID != "" {
		log.Printf("Rollback: removing failed new container %s", newContainerID)
		if err := cli.ContainerRemove(ctx, newContainerID, container.RemoveOptions{Force: true}); err != nil {
			log.Printf("Rollback: failed to remove new container %s: %v", newContainerID, err)
		}
	}

	// Try to recreate the original container
	log.Printf("Rollback: re-creating original container '%s' (image %s)", originalName, originalConfig.Image)
	resp, err := cli.ContainerCreate(ctx, originalConfig, originalHostConfig, nil, nil, originalName)
	if err != nil {
		log.Printf("Rollback failed: could not re-create original container: %v", err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Rollback failed: could not re-create container: %v", err))
		return
	}

	// Try to start the restored container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Printf("Rollback failed: could not start restored container: %v", err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Rollback failed: could not start restored container: %v", err))
		return
	}

	log.Printf("Rollback successful: restored original container (ID %s)", resp.ID)
	sendStatus(stream, update, orchestrator.UpdateStatus_COMPLETED, "Rollback successful. Original container is running.")
}

// pullImage pulls the new Docker image
func pullImage(cli *dockerclient.Client, ctx context.Context, stream orchestrator.HostAgentService_ConnectAgentStreamClient, update *orchestrator.UpdateContainerCommand) error {
	sendStatus(stream, update, orchestrator.UpdateStatus_PULLING, "Pulling new image")
	out, err := cli.ImagePull(ctx, update.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Pull failed for image %s: %v", update.Image, err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Failed to pull image: %v", err))
		return err
	}
	defer out.Close()
	io.Copy(io.Discard, out)
	log.Printf("Pulled image %s", update.Image)
	return nil
}

// stopAndRemoveContainer stops and removes the specified container
func stopAndRemoveContainer(cli *dockerclient.Client, ctx context.Context, stream orchestrator.HostAgentService_ConnectAgentStreamClient, update *orchestrator.UpdateContainerCommand, containerID string) error {
	sendStatus(stream, update, orchestrator.UpdateStatus_STARTING, "Stopping existing container")
	stopTimeout := 10
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &stopTimeout}); err != nil {
		log.Printf("Stop failed for container %s: %v", containerID, err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Failed to stop container: %v", err))
		return err
	}

	log.Printf("Removing old container %s", containerID)
	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: false}); err != nil {
		log.Printf("Remove failed for container %s: %v", containerID, err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Failed to remove container: %v", err))
		return err
	}
	return nil
}

// createNewContainer creates a new container based on provided configurations
func createNewContainer(cli *dockerclient.Client, ctx context.Context, stream orchestrator.HostAgentService_ConnectAgentStreamClient, update *orchestrator.UpdateContainerCommand, name string, inspect *container.InspectResponse) (string, error) {
	sendStatus(stream, update, orchestrator.UpdateStatus_RUNNING, "Creating new container")

	// Prepare configurations
	newConfig, newHostConfig := prepareConfigs(update, inspect)
	networkConfig := prepareNetworkConfig(inspect, newHostConfig)

	resp, err := cli.ContainerCreate(ctx, newConfig, newHostConfig, networkConfig, nil, name)
	if err != nil {
		log.Printf("Create failed: %v", err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Failed to create container: %v", err))
		return "", err
	}
	log.Printf("Created new container ID: %s", resp.ID)
	return resp.ID, nil
}

// startNewContainer starts the newly created container
func startNewContainer(cli *dockerclient.Client, ctx context.Context, stream orchestrator.HostAgentService_ConnectAgentStreamClient, update *orchestrator.UpdateContainerCommand, containerID string) error {
	log.Printf("Starting new container %s", containerID)
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		log.Printf("Start failed for container %s: %v", containerID, err)
		sendStatus(stream, update, orchestrator.UpdateStatus_FAILED, fmt.Sprintf("Failed to start container: %v", err))
		return err
	}
	return nil
}

// prepareConfigs prepares container and host configurations
func prepareConfigs(update *orchestrator.UpdateContainerCommand, inspect *container.InspectResponse) (*container.Config, *container.HostConfig) {
	// Base configs from original container
	newConfig := inspect.Config
	newHostConfig := inspect.HostConfig

	// Apply overrides
	newConfig.Image = update.Image
	if len(update.OverrideEnvVars) > 0 {
		newConfig.Env = update.OverrideEnvVars
	}
	if len(update.OverrideVolumes) > 0 {
		newHostConfig.Binds = update.OverrideVolumes
		newHostConfig.Mounts = nil // Binds and Mounts can conflict
	}
	if update.OverrideNetwork != "" {
		newHostConfig.NetworkMode = container.NetworkMode(update.OverrideNetwork)
	}

	// Handle port overrides
	if len(update.OverridePorts) > 0 {
		portSet, portMap := processPortOverrides(update.OverridePorts)
		newConfig.ExposedPorts = portSet
		newHostConfig.PortBindings = portMap
	}

	return newConfig, newHostConfig
}

// processPortOverrides converts protobuf PortMapping to Docker's format
func processPortOverrides(ports []*orchestrator.PortMapping) (nat.PortSet, nat.PortMap) {
	portSet := make(nat.PortSet)
	portMap := make(nat.PortMap)

	for _, p := range ports {
		protocol := strings.ToLower(p.Protocol)
		if protocol == "" {
			protocol = "tcp"
		}
		containerPort, err := nat.NewPort(protocol, strconv.FormatUint(uint64(p.ContainerPort), 10))
		if err != nil {
			log.Printf("⚠️ Invalid port mapping skipped: %v", err)
			continue
		}
		hostIP := p.HostIp
		if hostIP == "" {
			hostIP = "0.0.0.0"
		}
		binding := nat.PortBinding{
			HostIP:   hostIP,
			HostPort: strconv.FormatUint(uint64(p.HostPort), 10),
		}
		portSet[containerPort] = struct{}{}
		portMap[containerPort] = append(portMap[containerPort], binding)
	}
	return portSet, portMap
}

// prepareNetworkConfig prepares the networking configuration for the new container
func prepareNetworkConfig(inspect *container.InspectResponse, hostConfig *container.HostConfig) *network.NetworkingConfig {
	if len(inspect.NetworkSettings.Networks) > 0 && !hostConfig.NetworkMode.IsHost() && !hostConfig.NetworkMode.IsNone() {
		endpoints := make(map[string]*network.EndpointSettings)
		for name, settings := range inspect.NetworkSettings.Networks {
			endpoints[name] = &network.EndpointSettings{
				IPAMConfig: settings.IPAMConfig,
				Aliases:    settings.Aliases,
			}
		}
		return &network.NetworkingConfig{EndpointsConfig: endpoints}
	}
	return nil
}

// sendStatus is a helper to send status updates over the stream
func sendStatus(stream orchestrator.HostAgentService_ConnectAgentStreamClient, update *orchestrator.UpdateContainerCommand, stage orchestrator.UpdateStatus_Stage, logs string) {
	status := &orchestrator.UpdateStatus{
		ContainerUID: update.ContainerUID,
		MacAddress:   update.MacAddress,
		Image:        update.Image,
		Stage:        stage,
		Logs:         logs,
		Timestamp:    time.Now().String(),
	}
	if err := stream.Send(status); err != nil {
		log.Printf("Failed sending status update: %v", err)
	}
}

// cleanupOldImage removes the old Docker image if it's no longer in use
func cleanupOldImage(cli *dockerclient.Client, ctx context.Context, imageRef string) {
	time.Sleep(10 * time.Second) // Wait a bit before cleanup
	log.Printf("Cleaning up old image: %s", imageRef)

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Printf("Cleanup: could not list containers: %v", err)
		return
	}

	for _, c := range containers {
		if c.Image == imageRef {
			log.Printf("Old image %s still used by container %s; skipping removal", imageRef, c.ID)
			return
		}
	}

	if _, err := cli.ImageRemove(ctx, imageRef, image.RemoveOptions{PruneChildren: true}); err != nil {
		log.Printf("Could not remove old image %s: %v", imageRef, err)
	} else {
		log.Printf("Removed old image %s", imageRef)
	}
}
