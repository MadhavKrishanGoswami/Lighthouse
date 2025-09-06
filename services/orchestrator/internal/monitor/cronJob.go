package monitor

import (
	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	agentserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/agent"
)

var CronTimeInHours = 1 // Default to 1 hour

func SetCronTimeInHours(hours int) {
	if hours > 0 {
		CronTimeInHours = hours
	}
}

func StartCronJob(registryMonitorClient registry_monitor.RegistryMonitorServiceClient, queries *db.Queries, agentServer *agentserver.Server) {
	CronMonitor(CronTimeInHours*60, registryMonitorClient, queries, agentServer)
}
