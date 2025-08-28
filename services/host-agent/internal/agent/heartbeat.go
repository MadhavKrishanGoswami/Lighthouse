// Package agent provides functionalities for the heartbeat agent.
package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	host_agent "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

func Heartbeat(cli *dockerclient.Client, ctx context.Context, gRPCClient host_agent.HostAgentServiceClient) error {
	// List all containers on the host
	containersList, err := cli.ContainerList(ctx, container.ListOptions{
		All: true, // include stopped containers if you want
	})
	if err != nil {
		log.Printf("Failed to list containers: %v", err)
		return err
	}
	var containers []*host_agent.ContainerInfo

	// Iterate through each container and inspect it for full details
	// Iterate through each container and inspect it for full details
	for _, c := range containersList {
		inspect, err := cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			log.Printf("Failed to inspect container %s: %v", c.ID, err)
			continue
		}

		// Ports
		var ports []string
		if inspect.NetworkSettings != nil {
			for _, bindings := range inspect.NetworkSettings.Ports {
				for _, b := range bindings {
					ports = append(ports, b.HostPort)
				}
			}
		}

		// Env vars
		var envVars []string
		if inspect.Config != nil {
			envVars = append(envVars, inspect.Config.Env...)
		}

		// Volumes (mount sources)
		var volumes []string
		for _, m := range inspect.Mounts {
			volumes = append(volumes, m.Source)
		}

		// Build our container info
		cInfo := host_agent.ContainerInfo{
			ContainerID: c.ID,
			Name:        strings.TrimPrefix(inspect.Name, "/"),
			Image:       inspect.Config.Image,
			Ports:       ports,
			EnvVars:     envVars,
			Volumes:     volumes,
			Network:     string(inspect.HostConfig.NetworkMode),
		}

		containers = append(containers, &cInfo)
	}
	macAddress, err := getLocalMacAddress()
	if err != nil {
		log.Printf("Failed to get MAC address: %v", err)
	}

	HeartbeatRequest := &host_agent.HeartbeatRequest{
		MacAddress: macAddress,
		Containers: containers,
	}
	_, err = gRPCClient.Heartbeat(context.Background(), HeartbeatRequest)
	if err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
		return err
	}
	return nil
}

func getLocalMacAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		mac := iface.HardwareAddr.String()
		if mac != "" {
			return mac, nil
		}
	}
	return "", fmt.Errorf("no valid MAC address found")
}
