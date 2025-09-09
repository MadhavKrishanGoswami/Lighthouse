package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	OrchestratorAddr string
}

func MustLoad() *Config {
	var orchestratorAddr string
	flag.StringVar(&orchestratorAddr, "o", "", "address of the orchestrator (e.g., localhost:50051)")
	flag.Parse()
	// If the flag is not set, try to get the path from an environment variable.
	if orchestratorAddr == "" {
		orchestratorAddr = os.Getenv("ORCHESTRATOR_ADDR")
	}
	// If still no path, exit.
	if orchestratorAddr == "" {
		log.Fatal("orchestrator address is not set: use -o flag or ORCHESTRATOR_ADDR env variable")
	}

	var cfg Config
	cfg.OrchestratorAddr = orchestratorAddr

	return &cfg
}
