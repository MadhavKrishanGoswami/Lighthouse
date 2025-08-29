// Package agent
// handling the registration of the agent with the orchestrator.
package agent

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	host_agent "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

func RegisterAgent(cli *dockerclient.Client, ctx context.Context, gRPCClient host_agent.HostAgentServiceClient) error {
	// List all containers on the host
	containersList, err := cli.ContainerList(ctx, container.ListOptions{
		All: true, // include stopped containers if you want
	})
	if err != nil {
		log.Printf("Failed listing containers: %v", err)
		return err
	}

	var containers []*host_agent.ContainerInfo

	// Iterate through each container and inspect it for full details
	for _, c := range containersList {
		inspect, err := cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			log.Printf("Inspect failed for container %s: %v", c.ID, err)
			continue
		}

		// Ports -> structured PortMapping
		var ports []*host_agent.PortMapping
		if inspect.NetworkSettings != nil {
			for containerPortProto, bindings := range inspect.NetworkSettings.Ports {
				// containerPortProto looks like "80/tcp"
				parts := strings.Split(string(containerPortProto), "/")
				if len(parts) != 2 {
					continue
				}
				containerPort, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				protocol := parts[1]

				// If no bindings, it's exposed internally only
				if len(bindings) == 0 {
					ports = append(ports, &host_agent.PortMapping{
						HostIp:        "",
						HostPort:      0,
						ContainerPort: uint32(containerPort),
						Protocol:      protocol,
					})
					continue
				}

				for _, b := range bindings {
					hostIP := b.HostIP
					if hostIP == "" {
						hostIP = "0.0.0.0"
					}
					hostPort, _ := strconv.Atoi(b.HostPort)

					ports = append(ports, &host_agent.PortMapping{
						HostIp:        hostIP,
						HostPort:      uint32(hostPort),
						ContainerPort: uint32(containerPort),
						Protocol:      protocol,
					})
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

	// Get host information
	hostID, err := GetMACAddress()
	if err != nil {
		log.Printf("MAC address lookup failed: %v", err)
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Hostname lookup failed: %v", err)
		return err
	}
	ip, err := GetHostIP()
	if err != nil {
		log.Printf("Host IP lookup failed: %v", err)
		return err
	}
	hostInfo := &host_agent.HostInfo{
		MacAddress: hostID,
		Hostname:   hostname,
		IpAddress:  ip,
		Containers: containers,
	}

	// Pretty print the host info for now
	log.Printf("Host info collected: %+v", hostInfo)

	// Register the host with the orchestrator
	res, err := gRPCClient.RegisterHost(ctx, &host_agent.RegisterHostRequest{
		Host: hostInfo,
	})
	if err != nil {
		log.Printf("Host registration RPC failed: %v", err)
		return err
	}
	if res.Success {
		log.Printf("Host registered: %s", res.Message)
	} else {
		log.Printf("Host registration rejected: %s", res.Message)
	}
	return nil
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
