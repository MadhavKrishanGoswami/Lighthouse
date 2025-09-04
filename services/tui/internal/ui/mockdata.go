package ui

import "time"

func fetchMockHosts() []Host {
	hosts := []Host{
		{
			Name:          "Madhav-MacBookPro",
			IP:            "192.168.1.10",
			LastHeartbeat: time.Now(),
			MACAddress:    "00:1A:2B:3C:4D:5E",
		},
		{
			Name:          "RaspberryPi-Server",
			IP:            "192.168.1.11",
			LastHeartbeat: time.Now().Add(-2 * time.Minute),
			MACAddress:    "00:1A:2B:3C:4D:5F",
		},
		{
			Name:          "Ubuntu-VM01",
			IP:            "192.168.1.12",
			LastHeartbeat: time.Now().Add(-7 * time.Minute),
			MACAddress:    "00:1A:2B:3C:4D:60",
		},
		{
			Name:          "Windows-Laptop",
			IP:            "192.168.1.13",
			LastHeartbeat: time.Now().Add(-15 * time.Minute),
			MACAddress:    "00:1A:2B:3C:4D:61",
		},
		{
			Name:          "Docker-Host01",
			IP:            "192.168.1.14",
			LastHeartbeat: time.Now().Add(-1 * time.Minute),
			MACAddress:    "00:1A:2B:3C:4D:62",
		},
	}
	return hosts
}

// fetchMockContainers simulates realistic container data for a given host.
func fetchMockContainers(hostName string) []Container {
	return []Container{
		{Name: hostName + "-nginx", Image: "nginx:1.25", Status: "Running", IsWatching: true, IsUpdating: false},
		{Name: hostName + "-redis", Image: "redis:7-alpine", Status: "Exited", IsWatching: false, IsUpdating: true},
		{Name: hostName + "-postgres", Image: "postgres:15", Status: "Running", IsWatching: true, IsUpdating: true},
		{Name: hostName + "-rabbitmq", Image: "rabbitmq:3-management", Status: "Running", IsWatching: true, IsUpdating: false},
		{Name: hostName + "-prometheus", Image: "prom/prometheus:latest", Status: "Running", IsWatching: true, IsUpdating: false},
		{Name: hostName + "-grafana", Image: "grafana/grafana:latest", Status: "Running", IsWatching: true, IsUpdating: true},
		{Name: hostName + "-custom-app", Image: "myapp:v1.0", Status: "Exited", IsWatching: false, IsUpdating: false},
	}
}

// fetchMockServices simulates realistic service status data.
func fetchMockServices() Service {
	return Service{
		OrchestratorStatus:    true,
		RegistryMonitorStatus: true,
		DatabaseStatus:        true,
		TotalHosts:            5,
	}
}

// fetchMockLogs provides realistic system logs.
func fetchMockLogs() []string {
	return []string{
		"[green]2025-09-05 08:00:12[white] - System initialized successfully.",
		"[yellow]2025-09-05 08:02:45[white] - Host Madhav-MacBookPro connected.",
		"[red]2025-09-05 08:05:30[white] - Container redis exited unexpectedly on RaspberryPi-Server.",
		"[green]2025-09-05 08:10:10[white] - Cron job executed successfully on Ubuntu-VM01.",
		"[yellow]2025-09-05 08:15:00[white] - Host Windows-Laptop heartbeat delayed by 15 minutes.",
		"[green]2025-09-05 08:20:33[white] - Docker-Host01 deployed new container myapp:v1.0.",
		"[red]2025-09-05 08:25:12[white] - Error fetching metrics from Prometheus on Docker-Host01.",
		"[yellow]2025-09-05 08:30:50[white] - Grafana dashboard updated.",
	}
}
