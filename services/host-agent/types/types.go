package types

type ContainerInfo struct {
	containerID string
	name        string
	image       string
	ports       string
	envVars     string
	volumes     string
	network     string
}
type HostInfo struct {
	hostID   string
	hostname string
	ip       string
	ContainerInfo
}
type HeartbeatRequest struct {
	hostID      string
	timestamp   string
	cpuUsage    float64
	memoryUsage float64
}
type UpdateContainerRequest struct {
	deploymentID    string
	image           string
	overrideEnvVars string
	overridePorts   string
}
