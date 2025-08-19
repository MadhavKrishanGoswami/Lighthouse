// Package types defines the data structures used in the application.
package types

type ContainerInfo struct {
	ContainerID string   `json:"containerID"` // Docker container ID
	Name        string   `json:"name"`        // Container name
	Image       string   `json:"image"`       // Image name with tag
	Ports       []string `json:"ports"`       // Exposed ports
	EnvVars     []string `json:"envVars"`     // Environment variables
	Volumes     []string `json:"volumes"`     // Mounted volumes
	Network     string   `json:"network"`     // Network name
}
type HostInfo struct {
	HostID     string          `json:"hostID"`     // Unique identifier for the host
	Hostname   string          `json:"hostname"`   // Hostname of the machine
	IP         string          `json:"ip"`         // IP address of the HostInfo
	Containers []ContainerInfo `json:"containers"` // List of containers running on the HostInfo
}
