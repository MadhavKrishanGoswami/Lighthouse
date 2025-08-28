package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func UpdateContainer(cli *dockerclient.Client, ctx context.Context, update *orchestrator.UpdateContainerCommand, stream orchestrator.HostAgentService_ConnectAgentStreamClient) error {
	log.Printf("Pulling image: %s", update.Image)
	// Find the container by ID
	inspect, err := cli.ContainerInspect(ctx, update.ContainerUID)
	if err != nil {
		log.Printf("Error inspecting container: %v", err)
		log.Printf("Container with ID %s not found", update.ContainerUID)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         "Container not found",
			Timestamp:    time.Now().String(),
		})
		return err
	}
	if inspect.ID == "" {
		log.Printf("Container with ID %s not found", update.ContainerUID)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         "Container not found",
			Timestamp:    time.Now().String(),
		})
		return fmt.Errorf("container not found")
	}
	stream.Send(&orchestrator.UpdateStatus{
		DeploymentID: update.DeploymentID,
		ContainerUID: update.ContainerUID,
		MacAddress:   update.MacAddress,
		Stage:        orchestrator.UpdateStatus_PULLING,
		Logs:         "Pulling image",
		Timestamp:    time.Now().String(),
	})
	// Pull the new image
	out, err := cli.ImagePull(ctx, update.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image: %v", err)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         fmt.Sprintf("Failed to pull image: %v", err),
			Timestamp:    time.Now().String(),
		})
		return err
	}
	defer out.Close()
	stream.Send(&orchestrator.UpdateStatus{
		DeploymentID: update.DeploymentID,
		ContainerUID: update.ContainerUID,
		MacAddress:   update.MacAddress,
		Stage:        orchestrator.UpdateStatus_STARTING,
		Logs:         "Image pulled successfully",
		Timestamp:    time.Now().String(),
	})
	// Stop the existing container
	log.Printf("Stopping container: %s", update.ContainerUID)
	err = cli.ContainerStop(ctx, update.ContainerUID, container.StopOptions{})
	if err != nil {
		log.Printf("Error stopping container: %v", err)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         fmt.Sprintf("Failed to stop container: %v", err),
			Timestamp:    time.Now().String(),
		})
		return err
	}
	stream.Send(&orchestrator.UpdateStatus{
		DeploymentID: update.DeploymentID,
		ContainerUID: update.ContainerUID,
		MacAddress:   update.MacAddress,
		Stage:        orchestrator.UpdateStatus_RUNNING,
		Logs:         "Container stopped successfully",
		Timestamp:    time.Now().String(),
	})
	// Remove the existing container
	log.Printf("Removing container: %s", update.ContainerUID)
	err = cli.ContainerRemove(ctx, update.ContainerUID, container.RemoveOptions{Force: true})
	if err != nil {
		log.Printf("Error removing container: %v", err)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         fmt.Sprintf("Failed to remove container: %v", err),
			Timestamp:    time.Now().String(),
		})
		return err
	}
	stream.Send(&orchestrator.UpdateStatus{
		DeploymentID: update.DeploymentID,
		ContainerUID: update.ContainerUID,
		MacAddress:   update.MacAddress,
		Stage:        orchestrator.UpdateStatus_RUNNING,
		Logs:         "Container removed successfully",
		Timestamp:    time.Now().String(),
	})
	// Create and start a new container with the same configuration but new image
	log.Printf("Creating new container with image: %s", update.Image)
	// Recreate container with overrides

	envVars := inspect.Config.Env
	if len(update.OverrideEnvVars) > 0 {
		envVars = update.OverrideEnvVars
	}

	// Parse ports (format expected: "8080:80/tcp")
	portSet := nat.PortSet{}
	portMap := nat.PortMap{}
	for _, p := range update.OverridePorts {
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			log.Printf("Invalid port mapping: %s", p)
			continue
		}
		hostPort := parts[0]
		containerPort := parts[1]
		portProto := nat.Port(containerPort) // e.g. "80/tcp"
		portSet[portProto] = struct{}{}
		portMap[portProto] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}
	volumes := inspect.HostConfig.Binds
	if len(update.OverrideVolumes) > 0 {
		volumes = update.OverrideVolumes
	}

	networkMode := inspect.HostConfig.NetworkMode
	if update.OverrideNetwork != "" {
		networkMode = container.NetworkMode(update.OverrideNetwork)
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        update.Image,
		Cmd:          inspect.Config.Cmd,
		Env:          envVars,
		ExposedPorts: portSet,
	}, &container.HostConfig{
		Binds:        volumes,
		NetworkMode:  networkMode,
		PortBindings: portMap,
	}, nil, nil, "")
	if err != nil {
		log.Printf("Error creating container: %v", err)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         fmt.Sprintf("Failed to create container: %v", err),
			Timestamp:    time.Now().String(),
		})
		return err
	}
	log.Printf("Starting new container: %s", resp.ID)
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container: %v", err)
		stream.Send(&orchestrator.UpdateStatus{
			DeploymentID: update.DeploymentID,
			ContainerUID: update.ContainerUID,
			MacAddress:   update.MacAddress,
			Stage:        orchestrator.UpdateStatus_FAILED,
			Logs:         fmt.Sprintf("Failed to start container: %v", err),
			Timestamp:    time.Now().String(),
		})
		return err
	}
	stream.Send(&orchestrator.UpdateStatus{
		DeploymentID: update.DeploymentID,
		ContainerUID: resp.ID,
		MacAddress:   update.MacAddress,
		Stage:        orchestrator.UpdateStatus_COMPLETED,
		Logs:         "Container updated and started successfully",
		Timestamp:    time.Now().String(),
	})
	log.Printf("Container %s updated successfully", resp.ID)
	return nil
}
