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
	dockerclient "github.com/moby/moby/client"
)

func UpdateContainerStream(cli *dockerclient.Client, ctx context.Context, gRPCClient orchestrator.HostAgentServiceClient) error {
	// Get agent ID (MAC address)
	agentID, err := GetMACAddress()
	if err != nil {
		log.Printf("Failed to get MAC address: %v", err)
	}
	stream, err := gRPCClient.ConnectAgentStream(ctx)
	if err != nil {
		return err
	}
	log.Println("Connected to orchestrator for update stream")

	// Step 1: send dummy registration update
	err = stream.Send(&orchestrator.UpdateStatus{
		DeploymentID: "init-deployment",
		ContainerUID: "init-container",
		MacAddress:   agentID,
		Stage:        orchestrator.UpdateStatus_COMPLETED,
		Logs:         "Agent connected",
		Timestamp:    time.Now().String(),
	})
	if err != nil {
		return err
	}

	// Step 2: Goroutine to receive commands from orchestrator
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down stream listener")
				return
			default:
				cmd, err := stream.Recv()
				if err == io.EOF {
					log.Println("Stream closed by server")
					return
				}
				if err != nil {
					log.Printf("Error receiving command: %v", err)
					continue
				}

				log.Printf("Received command: %+v", cmd)

				// Simulate container update work
				time.Sleep(3 * time.Second)

				// Step 3: Send back status after executing command
				err = stream.Send(&orchestrator.UpdateStatus{
					DeploymentID: cmd.DeploymentID,
					ContainerUID: cmd.ContainerUID,
					MacAddress:   agentID,
					Stage:        orchestrator.UpdateStatus_COMPLETED,
					Logs:         "Container updated successfully",
					Timestamp:    time.Now().String(),
				})
				if err != nil {
					log.Printf("Failed to send update: %v", err)
				}
			}
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Received shutdown signal, cleaning up...")

	// Close the send side of the stream
	if err := stream.CloseSend(); err != nil {
		log.Printf("Error closing stream: %v", err)
	}

	return nil
}
