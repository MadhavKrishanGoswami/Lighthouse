package config

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	OrchestratorAddr string `mapstructure:"orchestrator_addr"`
}

// MustLoad reads configuration using a priority system: flags > env > file > defaults.
func MustLoad() *Config {
	// --- Highest Priority: Command-line flags for development override ---
	var orchestratorAddrFlag string
	// We define the flag here.
	flag.StringVar(&orchestratorAddrFlag, "o", "", "address of the orchestrator (e.g., localhost:50051) [dev override]")
	flag.Parse()

	// --- Initialize Viper ---
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")

	// --- Define Configuration Search Paths ---
	// 1. For Windows: C:\ProgramData\LighthouseHostAgent
	if os.Getenv("ProgramData") != "" {
		viper.AddConfigPath(filepath.Join(os.Getenv("ProgramData"), "LighthouseHostAgent"))
	}
	// 2. For Linux: /etc/lighthouse-host-agent/
	viper.AddConfigPath("/etc/lighthouse-host-agent/")
	// 3. For development: look in the current directory
	viper.AddConfigPath(".")

	// --- Set Defaults (Lowest Priority) ---
	viper.SetDefault("orchestrator_addr", "localhost:50051")

	// --- Bind to Environment Variables ---
	// This allows overriding config file values with env vars
	viper.SetEnvPrefix("LIGHTHOUSE") // will look for LIGHTHOUSE_ORCHESTRATOR_ADDR
	viper.BindEnv("orchestrator_addr", "ORCHESTRATOR_ADDR")

	// --- Read Configuration from file ---
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error, will use defaults/env/flags.
			log.Println("Config file not found, using other configuration sources.")
		} else {
			// Config file was found but another error was produced
			log.Fatalf("Fatal error reading config file: %s", err)
		}
	}

	// --- Apply the flag override if it was provided ---
	// Because this comes after reading the config file, it will take precedence.
	if orchestratorAddrFlag != "" {
		viper.Set("orchestrator_addr", orchestratorAddrFlag)
		log.Printf("Using developer flag override for orchestrator address: %s", orchestratorAddrFlag)
	}

	// --- Unmarshal into Struct ---
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct, %v", err)
	}

	return &cfg
}
