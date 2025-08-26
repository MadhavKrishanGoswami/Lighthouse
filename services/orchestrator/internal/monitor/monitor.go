package monitor

import (
	"context"
	"fmt"
	"time"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
)

func CronMonitor(timeinmin int, grpcClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries) {
	// Convert minutes to duration
	duration := time.Duration(timeinmin) * time.Minute
	ctx := context.Background()
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Ticker ticked! Checking for updates...")
			ChecckForUpdates(ctx, grpcClient, queries)
		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping the ticker.")
			return
		}
	}
}
