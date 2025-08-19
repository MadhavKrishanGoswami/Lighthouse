// Package agent
// handling the registration of the agent with the orchestrator.
package agent

import (
	"context"
	"log"
	"net"
	"os"
	"strings"

	lhtypes "github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/types"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func RegisterAgent(cli *client.Client, ctx context.Context) {
	// List all containers on the host
	containersList, err := cli.ContainerList(ctx, container.ListOptions{
		All: true, // include stopped containers if you want
	})
	if err != nil {
		log.Printf("Failed to list containers: %v", err)
		return
	}

	var containers []lhtypes.ContainerInfo

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
		cInfo := lhtypes.ContainerInfo{
			ContainerID: c.ID,
			Name:        strings.TrimPrefix(inspect.Name, "/"),
			Image:       inspect.Config.Image,
			Ports:       ports,
			EnvVars:     envVars,
			Volumes:     volumes,
			Network:     string(inspect.HostConfig.NetworkMode),
		}

		containers = append(containers, cInfo)
	}

	// Get host information
	hostID, err := GetMACAddress()
	if err != nil {
		log.Printf("Failed to get MAC address: %v", err)
		return
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Failed to get hostname: %v", err)
		return
	}
	ip, err := GetHostIP()
	if err != nil {
		log.Printf("Failed to get host IP: %v", err)
		return
	}
	hostInfo := lhtypes.HostInfo{
		HostID:     hostID,
		Hostname:   hostname,
		IP:         ip,
		Containers: containers,
	}
	// Prrtty print the host info for now
	log.Printf("Host Info: %+v", hostInfo)
}

func GetMACAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.HardwareAddr != nil {
			return iface.HardwareAddr.String(), nil
		}
	}
	return "", nil
}

func GetHostIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", nil
}
