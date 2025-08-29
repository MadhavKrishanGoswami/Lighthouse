// Package agent
// this updates the orcastater with any updates to the host or containers
package agent

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	dockerclient "github.com/docker/docker/client"
)

func UpdateContainerStream(cli *dockerclient.Client, ctx context.Context, gRPCClient orchestrator.HostAgentServiceClient) error {
	// Get agent ID (MAC address)
	agentID, err := GetMACAddress()
	if err != nil {
		log.Printf("MAC address lookup failed: %v", err)
	}
	stream, err := gRPCClient.ConnectAgentStream(ctx)
	if err != nil {
		return err
	}
	log.Println("Connected to orchestrator stream")

	// Step 1: send dummy registration update
	err = stream.Send(&orchestrator.UpdateStatus{
		ContainerUID: "init-container",
		MacAddress:   agentID,
		Stage:        orchestrator.UpdateStatus_COMPLETED,
		Logs:         "Agent connected",
		Timestamp:    time.Now().String(),
		Image:        "init-image",
	})
	if err != nil {
		return err
	}

	// Step 2: Goroutine to receive commands from orchestrator
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stream listener stopping")
				return
			default:
				cmd, err := stream.Recv()
				if err == io.EOF {
					log.Println("Stream closed by server")
					return
				}
				if err != nil {
					log.Printf("Command receive error: %v", err)
					continue
				}

				log.Printf("Received command: %+v", cmd)
				// Process the command to update container
				err = UpdateContainer(cli, ctx, cmd, stream)
				if err != nil {
					log.Printf("Update command failed: %v", err)
				}
			}
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutdown signal received, cleaning up...")

	// Close the send side of the stream
	if err := stream.CloseSend(); err != nil {
		log.Printf("Error closing stream: %v", err)
	}

	return nil
}
