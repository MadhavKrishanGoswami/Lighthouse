// Package server provides a gRPC server implementation for handling requests.
package agentserver

import (
	"context"
	"log"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type server struct {
	orchestrator.UnimplementedHostAgentServiceServer
	db *db.Queries // Database queries instance
}

func NewServer(queries *db.Queries) *server {
	return &server{
		db: queries, // Initialize the database queries instance
	}
}

func (s *server) RegisterHost(ctx context.Context, req *orchestrator.RegisterHostRequest) (*orchestrator.RegisterHostResponse, error) {
	// Implement the logic to register a host here
	log.Printf("Received request to register host: %s", req.Host.Hostname+" with IP: "+req.Host.IpAddress)

	params := db.InsertHostParams{
		MacAddress: req.Host.MacAddress,
		Hostname:   req.Host.Hostname,
		IpAddress:  req.Host.IpAddress,
	}
	host, err := s.db.InsertHost(ctx, params)
	if err != nil {
		log.Printf("Failed to register host: %v", err)
		return &orchestrator.RegisterHostResponse{
			Success: false,
			Message: "Failed to register host in the database with error: " + err.Error(),
		}, nil
	}
	for _, container := range req.Host.Containers {
		log.Printf("Container: %v", container)
		containerParams := db.InsertContainerParams{
			ContainerUid: container.ContainerID,
			HostID:       host.ID,
			Name:         container.Name,
			Image:        container.Image,
			Ports:        container.Ports,
			EnvVars:      container.EnvVars,
			Volumes:      container.Volumes,
			Network:      pgtype.Text{String: container.Network, Valid: true},
		}
		_, err := s.db.InsertContainer(ctx, containerParams)
		if err != nil {
			log.Printf("Failed to register container: %v", err)
		}
	}
	log.Printf("Host registered successfully with ID: %s", host.IpAddress)
	return &orchestrator.RegisterHostResponse{
		Success: true,
		Message: "Host registered successfully with ID: " + string(host.IpAddress),
	}, nil
}

func (s *server) Heartbeat(ctx context.Context, req *orchestrator.HeartbeatRequest) (*orchestrator.HeartbeatResponse, error) {
	// Implement the logic to handle heartbeat messages here
	log.Printf("Received heartbeat from host: %s", req.MacAddress)
	// Log each container info
	host, err := s.db.GertHostByMacAddress(ctx, req.MacAddress)
	if err != nil {
		log.Printf("Failed to get host by IP: %v", err)
		return &orchestrator.HeartbeatResponse{
			Success: false,
			Message: "Failed to get host by IP with error: " + err.Error(),
		}, nil
	}
	// inserat last heartbeat timestamp
	_, err = s.db.UpdateHostLastHeartbeat(ctx, host.ID)
	if err != nil {
		log.Printf("Failed to update host last heartbeat: %v", err)
		return &orchestrator.HeartbeatResponse{
			Success: false,
			Message: "Failed to update host last heartbeat with error: " + err.Error(),
		}, nil
	}
	// Upsert containers
	for _, container := range req.Containers {
		containerParams := db.InsertContainerParams{
			ContainerUid: container.ContainerID,
			HostID:       host.ID,
			Name:         container.Name,
			Image:        container.Image,
			Ports:        container.Ports,
			EnvVars:      container.EnvVars,
			Volumes:      container.Volumes,
			Network:      pgtype.Text{String: container.Network, Valid: true},
		}
		_, err := s.db.InsertContainer(ctx, containerParams)
		if err != nil {
			log.Printf("Failed to register container: %v", err)
		}
	}
	return &orchestrator.HeartbeatResponse{
		Success: true,
		Message: "Heartbeat received successfully",
	}, nil
}
