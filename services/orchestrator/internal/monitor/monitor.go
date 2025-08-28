package monitor

import (
	"context"
	"fmt"
	"log"
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
				// Itrate over all the images to update find the host with he container uid and send the update command to that host
				for _, image := range toUpdateContainers.ImagestoUpdate {
					host, err := queries.GetHostbyContainerUID(ctx, image.ContainerUid)
					if err != nil {
						fmt.Printf("Error getting host ID for container UID %s: %v\n", image.ContainerUid, err)
						continue
					}
					container, err := queries.GetContainerbyContainerUID(ctx, image.ContainerUid)
					if err != nil {
						fmt.Printf("Error getting container by UID %s: %v\n", image.ContainerUid, err)
						continue
					}
					log.Printf("Sending update command to host %s for container UID %s\n", host.ID, image.ContainerUid)
					// Save data to deployment table
					deplayment, err := queries.InsertDeployment(ctx, db.InsertDeploymentParams{
						ContainerID: container.ID,
						HostID:      host.ID,
						TargetImage: image.NewTag,
						Status:      db.DeploymentStatusPending,
					})
					if err != nil {
						fmt.Printf("Error inserting deployment record for container UID %s: %v\n", image.ContainerUid, err)
						continue
					}
					hostStream, ok := agentServer.Hosts[host.MacAddress]
					if !ok {
						fmt.Printf("No active stream found for host %s\n", host.ID)
						continue
					}
					cmd := orchestrator.UpdateContainerCommand{
						DeploymentID:    deplayment.ID.String(),
						ContainerUID:    image.ContainerUid,
						Image:           image.NewTag,
						OverrideEnvVars: container.EnvVars,
						OverridePorts:   container.Ports,
						OverrideVolumes: container.Volumes,
						OverrideNetwork: container.Network.String,
						MacAddress:      host.MacAddress, // target host MAC
					}
					if err := hostStream.Send(&cmd); err != nil {
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
