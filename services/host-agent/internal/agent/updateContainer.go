package agent

import (
	"context"
	"log"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	dockerclient "github.com/docker/docker/client"
)

func UpdateContainer(cli *dockerclient.Client, ctx context.Context, update *orchestrator.UpdateContainerCommand) error {
	log.Printf("Updating container: %+v", update.ContainerUID)
	// Pull the new image with latest tag  and send status to  the orchestrator via gRPC stream

	return nil
}
