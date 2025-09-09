package main

import (
	"context"
	"log"
	"sync"
	"time"

	hostagent "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/config"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/agent"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/host-agent/internal/client"
	dockerclient "github.com/docker/docker/client"
	"github.com/kardianos/service"
	"google.golang.org/grpc"
)

var logger service.Logger

// --- Program Struct ---
// This struct holds all the state for our running agent.
type program struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc

	// Add other necessary fields
	config     *config.Config
	dockerCli  *dockerclient.Client
	grpcClient hostagent.HostAgentServiceClient
	grpcConn   *grpc.ClientConn
}

// --- Service Interface Methods ---

// Start is called by the service manager when the service is started.
func (p *program) Start(s service.Service) error {
	// The Start method must not block.
	// We will start our main logic in a goroutine.
	go p.run()
	logger.Info("Lighthouse Host-Agent Service started.")
	return nil
}

// Stop is called by the service manager when the service is stopped.
func (p *program) Stop(s service.Service) error {
	logger.Info("Lighthouse Host-Agent Service stopping...")

	// Signal all goroutines to stop by canceling the context.
	if p.cancel != nil {
		p.cancel()
	}

	// Wait for all goroutines to finish.
	p.wg.Wait()

	// Clean up resources.
	if p.grpcConn != nil {
		if err := p.grpcConn.Close(); err != nil {
			logger.Errorf("Failed to close gRPC client connection: %v", err)
		}
	}
	if p.dockerCli != nil {
		if err := p.dockerCli.Close(); err != nil {
			logger.Errorf("Failed to close Docker client connection: %v", err)
		}
	}

	logger.Info("Lighthouse Host-Agent Service stopped.")
	return nil
}

// --- Main Application Logic ---

// run contains the core logic of the agent.
func (p *program) run() {
	// ---  Configuration Loading ---
	p.config = config.MustLoad()
	logger.Info("Configuration loaded successfully.")

	// --- 1. Start gRPC client ---
	var err error
	p.grpcClient, p.grpcConn, err = client.StartClient(*p.config)
	if err != nil {
		logger.Errorf("gRPC client init failed: %v", err)
		return // Exit if we can't connect
	}
	logger.Info("Connected to orchestrator gRPC endpoint.")

	// --- 2. Docker client setup ---
	p.dockerCli, err = dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		logger.Errorf("Docker client creation failed: %v", err)
		return // Exit if Docker isn't available
	}

	// --- 3. Context for all goroutines ---
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	// --- 4. Register agent with orchestrator ---
	regCtx, regCancel := context.WithTimeout(ctx, 10*time.Second)
	defer regCancel()
	if err := agent.RegisterAgent(p.dockerCli, regCtx, p.grpcClient); err != nil {
		logger.Errorf("Agent registration failed: %v", err)
		return // Exit if we can't register
	}
	logger.Info("Agent registered with orchestrator.")

	// --- 6. Heartbeat goroutine ---
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := agent.Heartbeat(p.dockerCli, ctx, p.grpcClient); err != nil {
					logger.Warningf("Heartbeat send failed: %v", err)
				}
			case <-ctx.Done():
				logger.Info("Heartbeat loop stopping (context canceled).")
				return
			}
		}
	}()

	// --- 7. UpdateContainer gRPC stream goroutine ---
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := agent.UpdateContainerStream(p.dockerCli, ctx, p.grpcClient); err != nil && ctx.Err() == nil {
			logger.Errorf("UpdateContainer stream error: %v", err)
		}
		logger.Info("UpdateContainer stream stopped.")
	}()

	logger.Info("Agent is now running in the background.")
}

// --- Main Function ---
func main() {
	// Configure the service.
	svcConfig := &service.Config{
		Name:        "LighthouseHostAgent",
		DisplayName: "Lighthouse Host Agent",
		Description: "Monitors the host system for the Lighthouse platform.",
		// Set dependencies if needed, e.g., "After=network.target" on Linux.
	}

	prg := &program{}

	// Create the service.
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Setup the logger.
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Run the service. This will block until the service is stopped.
	// It also handles the command-line arguments like install, uninstall, start, stop.
	if err := s.Run(); err != nil {
		logger.Error(err)
	}
}
