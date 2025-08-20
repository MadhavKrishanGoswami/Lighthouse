// Package config provides functionality to load application configuration
package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// GRPCServer configures the HTTP server.
type GRPCServer struct {
	Addr string `yaml:"address"`
}

// Config holds all configuration for the application.
type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	DataBaseURL string `yaml:"DataBaseURL" env-required:"true"`
	GRPCServer  `yaml:"gRPCServer"`
}

// MustLoad loads the configuration from environment variables and panics if it fails.
func MustLoad() *Config {
	// Define the config path lag
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to the configuration file (e.g., config/local.yaml)")
	flag.Parse()
	// If the flag is not set, try to get the path from an environment variable.
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	// If still no path, exit.
	if configPath == "" {
		log.Fatal("config path is not set: use -config flag or CONFIG_PATH env variable")
	}

	// Check if the file exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	// Read the configuration file into the struct.
	// CRITICAL: We must pass a pointer to cfg using '&'.
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config file: %s", err)
	}

	return &cfg
}
