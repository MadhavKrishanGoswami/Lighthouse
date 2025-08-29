package monitor

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	agentserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/agent"
)

func CronMonitor(timeinmin int, grpcClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries, agentServer *agentserver.Server) {
	// Convert minutes to duration
	duration := time.Duration(timeinmin) * time.Minute
	ctx := context.Background()
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Ticker ticked! Checking for updates...")
			toUpdateContainers, err := CheckForUpdates(ctx, grpcClient, queries)
			if err != nil {
				fmt.Printf("Error checking for updates: %v\n", err)
				continue
			}
			if len(toUpdateContainers.ImagestoUpdate) > 0 {
				for _, image := range toUpdateContainers.ImagestoUpdate {
					// Get the host where this container is running
					host, err := queries.GetHostbyContainerUID(ctx, image.ContainerUid)
					if err != nil {
						fmt.Printf("Error getting host ID for container UID %s: %v\n", image.ContainerUid, err)
						continue
					}

					// Get container details from DB
					container, err := queries.GetContainerbyContainerUID(ctx, image.ContainerUid)
					if err != nil {
						fmt.Printf("Error getting container by UID %s: %v\n", image.ContainerUid, err)
						continue
					}

					log.Printf("Sending update command to host %s for container UID %s\n", host.ID, image.ContainerUid)

					hostStream, ok := agentServer.Hosts[host.MacAddress]
					if !ok {
						fmt.Printf("No active stream found for host %s\n", host.ID)
						continue
					}

					// Convert DB stored ports ([]string) into []*orchestrator.PortMapping
					var overridePorts []*orchestrator.PortMapping
					for _, p := range container.Ports {
						// expected format from DB: "hostIP:hostPort->containerPort/protocol"
						parts := strings.Split(p, "->")
						if len(parts) != 2 {
							continue
						}
						// Left side = hostIP:hostPort
						hostParts := strings.Split(parts[0], ":")
						if len(hostParts) != 2 {
							continue
						}
						hostIP := hostParts[0]
						hostPort, _ := strconv.Atoi(hostParts[1])

						// Right side = containerPort/protocol
						containerParts := strings.Split(parts[1], "/")
						if len(containerParts) != 2 {
							continue
						}
						containerPort, _ := strconv.Atoi(containerParts[0])
						protocol := containerParts[1]

						overridePorts = append(overridePorts, &orchestrator.PortMapping{
							HostIp:        hostIP,
							HostPort:      uint32(hostPort),
							ContainerPort: uint32(containerPort),
							Protocol:      protocol,
						})
					}

					// Build update command
					cmd := orchestrator.UpdateContainerCommand{
						ContainerUID:    image.ContainerUid,
						Image:           image.NewTag,
						OverrideEnvVars: container.EnvVars,
						OverridePorts:   overridePorts,
						OverrideVolumes: container.Volumes,
						OverrideNetwork: container.Network.String,
						MacAddress:      host.MacAddress, // target host MAC
					}

					// Send update command
					if err := hostStream.Stream.Send(&cmd); err != nil {
						fmt.Printf("Error sending update command to host %s: %v\n", host.ID, err)
						continue
					}
					log.Printf("Update command sent to host %s for container UID %s\n", host.ID, image.ContainerUid)
				}
			} else {
				fmt.Println("No images to update.")
			}
		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping the ticker.")
			return
		}
	}
}
