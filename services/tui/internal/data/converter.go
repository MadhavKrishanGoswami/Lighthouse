package data

import (
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/ui"
)

// ConvertHostInfo converts protobuf HostInfo to UI Host struct
func ConvertHostInfo(hostInfo *orchestrator.HostInfo) ui.Host {
	lastHeartbeat, _ := time.Parse(time.RFC3339, hostInfo.GetLastHeartbeat())

	return ui.Host{
		Name:          hostInfo.GetHostname(),
		IP:            hostInfo.GetIpAddress(),
		MACAddress:    hostInfo.GetMacAddress(),
		LastHeartbeat: lastHeartbeat,
	}
}

// ConvertContainerInfo converts protobuf ContainerInfo to UI Container struct
func ConvertContainerInfo(containerInfo *orchestrator.ContainerInfo) ui.Container {
	status := convertContainerStatus(containerInfo.GetStatus())

	return ui.Container{
		Name:       containerInfo.GetName(),
		Image:      containerInfo.GetImage(),
		Status:     status,
		IsWatching: containerInfo.GetWatch(),
		IsUpdating: false, // You may need to add this field to protobuf
	}
}

// convertContainerStatus maps protobuf enum to string
func convertContainerStatus(status orchestrator.ContainerInfo_Status) string {
	switch status {
	case orchestrator.ContainerInfo_RUNNING:
		return "Running"
	case orchestrator.ContainerInfo_STOPPED:
		return "Stopped"
	case orchestrator.ContainerInfo_PAUSED:
		return "Paused"
	case orchestrator.ContainerInfo_RESTARTING:
		return "Restarting"
	case orchestrator.ContainerInfo_EXITED:
		return "Exited"
	case orchestrator.ContainerInfo_DEAD:
		return "Dead"
	default:
		return "Unknown"
	}
}

// ConvertServicesStatus converts protobuf servicesStatus to UI Service struct
func ConvertServicesStatus(servicesStatus []*orchestrator.ServicesStatus) ui.Service {
	service := ui.Service{}

	for _, svc := range servicesStatus {
		switch svc.GetServices() {
		case orchestrator.ServicesStatus_ORCHESTRATOR:
			service.OrchestratorStatus = svc.GetStatus()
		case orchestrator.ServicesStatus_REGISTRY_Monitor:
			service.RegistryMonitorStatus = svc.GetStatus()
		case orchestrator.ServicesStatus_Database:
			service.DatabaseStatus = svc.GetStatus()
		}
	}

	return service
}
