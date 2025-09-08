package servicestatus

var (
	Orchestrator = true
	Database     = true
	Registry     = true
)

func GetServiceStatus() (bool, bool, bool) {
	return Orchestrator, Database, Registry
}

func SetOrchestratorStatus(status bool) {
	Orchestrator = status
}

func SetDatabaseStatus(status bool) {
	Database = status
}

func SetRegistryStatus(status bool) {
	Registry = status
}
