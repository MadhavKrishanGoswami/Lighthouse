package monitor

import (
	"context"
	"log"
	"sync"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	agentserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/agent"
)

var (
	CronTimeInHours = 1 // Default to 1 hour

	cronMu     sync.Mutex
	cronCancel context.CancelFunc
	cronArgs   struct {
		registryMonitorClient registry_monitor.RegistryMonitorServiceClient
		queries               *db.Queries
		agentServer           *agentserver.Server
	}
)

// SetRuntimeDeps allows wiring dependencies once so restart can reuse them.
func SetRuntimeDeps(registryMonitorClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries, agentServer *agentserver.Server) {
	cronMu.Lock()
	defer cronMu.Unlock()
	cronArgs.registryMonitorClient = registryMonitorClient
	cronArgs.queries = queries
	cronArgs.agentServer = agentServer
}

func SetCronTimeInHours(hours int) {
	cronMu.Lock()
	defer cronMu.Unlock()
	if hours <= 0 {
		log.Printf("[Cron] Ignoring invalid cron hours: %d", hours)
		return
	}
	if hours == CronTimeInHours {
		log.Printf("[Cron] Cron time unchanged (%dh), not restarting", hours)
		return
	}
	CronTimeInHours = hours
	log.Printf("[Cron] Updating cron time to %d hour(s). Restarting...", hours)
	// stop existing
	if cronCancel != nil {
		cronCancel()
		cronCancel = nil
	}
	// start new if deps are set
	if cronArgs.registryMonitorClient != nil && cronArgs.queries != nil && cronArgs.agentServer != nil {
		ctx, cancel := context.WithCancel(context.Background())
		cronCancel = cancel
		go CronMonitor(ctx, CronTimeInHours*60, cronArgs.registryMonitorClient, cronArgs.queries, cronArgs.agentServer)
	} else {
		log.Printf("[Cron] Dependencies not set; will start on next StartCronJob")
	}
}

func StartCronJob(registryMonitorClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries, agentServer *agentserver.Server) {
	cronMu.Lock()
	defer cronMu.Unlock()
	// store deps for restarts
	cronArgs.registryMonitorClient = registryMonitorClient
	cronArgs.queries = queries
	cronArgs.agentServer = agentServer
	// stop existing
	if cronCancel != nil {
		log.Printf("[Cron] Stopping existing cron before starting new one")
		cronCancel()
		cronCancel = nil
	}
	// start new
	ctx, cancel := context.WithCancel(context.Background())
	cronCancel = cancel
	log.Printf("[Cron] Starting cron with interval %dh", CronTimeInHours)
	go CronMonitor(ctx, CronTimeInHours*60, registryMonitorClient, queries, agentServer)
}
