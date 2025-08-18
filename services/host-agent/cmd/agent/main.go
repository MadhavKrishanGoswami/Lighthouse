package main

import "github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/agent"

func main() {
	// This is the main entry point for the application.
	// The actual functionality is implemented in the agent package.
	// You can start the agent, monitor containers, or register with an orchestrator here.

	agent.StartAgent()
}
