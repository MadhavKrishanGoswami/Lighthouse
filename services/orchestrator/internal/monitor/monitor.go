package monitor

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	agentserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/agent"
)

func CronMonitor(ctx context.Context, timeinmin int, grpcClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries, agentServer *agentserver.Server) {
	// Convert minutes to duration
	duration := time.Duration(timeinmin) * time.Minute
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Cron: checking for updates")
			toUpdateContainers, err := CheckForUpdates(ctx, grpcClient, queries)
			if err != nil {
				log.Printf("CheckForUpdates failed: %v", err)
				continue
			}
			if len(toUpdateContainers.ImagestoUpdate) > 0 {
				for _, image := range toUpdateContainers.ImagestoUpdate {
					// Get the host where this container is running
					host, err := queries.GetHostbyContainerUID(ctx, image.ContainerUid)
					if err != nil {
						log.Printf("Get host for container %s failed: %v", image.ContainerUid, err)
						continue
					}

					// Get container details from DB
					container, err := queries.GetContainerbyContainerUID(ctx, image.ContainerUid)
					if err != nil {
						log.Printf("Get container %s failed: %v", image.ContainerUid, err)
						continue
					}

					log.Printf("Sending update command host %s container %s", host.ID, image.ContainerUid)

					hostStream, ok := agentServer.Hosts[host.MacAddress]
					if !ok {
						log.Printf("No active stream for host %s", host.ID)
						continue
					}

					// Convert DB stored ports ([]byte JSON) into []*orchestrator.PortMapping
					var overridePorts []*orchestrator.PortMapping
					var portStrings []string
					if err := json.Unmarshal(container.Ports, &portStrings); err != nil {
						log.Printf("Could not unmarshal ports for container %s: %v", image.ContainerUid, err)
					} else {
						for _, p := range portStrings {
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
						log.Printf("Send update command host %s failed: %v", host.ID, err)
						continue
					}
					log.Printf("Update command sent host %s container %s", host.ID, image.ContainerUid)
				}
			} else {
				log.Println("No images to update")
			}
		case <-ctx.Done():
			log.Println("Cron monitor context canceled")
			return
		}
	}
}
