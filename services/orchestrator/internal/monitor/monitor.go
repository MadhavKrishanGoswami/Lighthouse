package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

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
			toUpdateContainers, err := ChecckForUpdates(ctx, grpcClient, queries)
			if err != nil {
				fmt.Printf("Error checking for updates: %v\n", err)
				continue
			}
			if len(toUpdateContainers.ImagestoUpdate) > 0 {
				// Itrate over all the images to update find the host with the container uid and send the update command to that host
				for _, image := range toUpdateContainers.ImagestoUpdate {
					hostid, err := queries.GetMacAddressByContainerUID(ctx, image.ContainerUid)
					if err != nil {
						fmt.Printf("Error getting host ID for container UID %s: %v\n", image.ContainerUid, err)
						continue
					}
					log.Printf("Host ID for container UID %s: %s\n", image.ContainerUid, hostid)
					// Find the host with the hostid in the agentServer.Hosts map
					// host, ok := agentServer.Hosts[hostid]
					// if !ok {
					// 	fmt.Printf("Host with ID %s not found in agent server\n", hostid)
					// 	continue
					// }
					// // Save the deployment in the database
					// if err != nil {
					// 	fmt.Printf("Error sending update command to host %s for container UID %s: %v\n", hostid, image.ContainerUid, err)
					// 	continue
					// }

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
