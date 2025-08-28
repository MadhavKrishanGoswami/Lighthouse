// Package agentserver provides a gRPC server implementation for handling requests.
package agentserver

import (
	"context"
	"log"
	"sync"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Server struct {
	orchestrator.UnimplementedHostAgentServiceServer
	db *db.Queries // Database queries instance
	mu sync.Mutex

	Hosts map[string]orchestrator.HostAgentService_ConnectAgentStreamServer
}

func NewServer(queries *db.Queries) *Server {
	return &Server{
		db:    queries, // Initialize the database queries instance
		Hosts: make(map[string]orchestrator.HostAgentService_ConnectAgentStreamServer),
	}
}

func (s *Server) RegisterHost(ctx context.Context, req *orchestrator.RegisterHostRequest) (*orchestrator.RegisterHostResponse, error) {
	if req == nil || req.Host == nil {
		return &orchestrator.RegisterHostResponse{Success: false, Message: "invalid request: host is nil"}, nil
	}

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
			Digest:       container.Digest,
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
	log.Printf("Host registered successfully with IP: %s", host.IpAddress)
	return &orchestrator.RegisterHostResponse{
		Success: true,
		Message: "Host registered successfully with IP: " + host.IpAddress,
	}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *orchestrator.HeartbeatRequest) (*orchestrator.HeartbeatResponse, error) {
	log.Printf("Received heartbeat from host: %s", req.MacAddress)

	host, err := s.db.GetHostByMacAddress(ctx, req.MacAddress)
	if err != nil {
		log.Printf("Failed to get host by MAC address: %v", err)
		return &orchestrator.HeartbeatResponse{
			Success: false,
			Message: "Failed to get host by MAC address: " + err.Error(),
		}, nil
	}

	// Update last heartbeat timestamp
	_, err = s.db.UpdateHostLastHeartbeat(ctx, host.ID)
	if err != nil {
		log.Printf("Failed to update host last heartbeat: %v", err)
		// Non-critical error, we can continue processing containers
	}

	// --- NEW: Step 1 - Collect all active container UIDs from the request ---
	// We'll use this list later to determine which containers are stale.
	activeContainerUIDs := make([]string, 0, len(req.Containers))
	for _, container := range req.Containers {
		activeContainerUIDs = append(activeContainerUIDs, container.ContainerID)
	}

	// Upsert all containers from the current heartbeat
	for _, container := range req.Containers {
		containerParams := db.InsertContainerParams{
			ContainerUid: container.ContainerID,
			HostID:       host.ID,
			Name:         container.Name,
			Image:        container.Image,
			Digest:       container.Digest,
			Ports:        container.Ports,
			EnvVars:      container.EnvVars,
			Volumes:      container.Volumes,
			Network:      pgtype.Text{String: container.Network, Valid: true},
		}
		_, err := s.db.InsertContainer(ctx, containerParams)
		if err != nil {
			// Log the error but continue trying to process other containers
			log.Printf("Failed to upsert container %s: %v", container.Name, err)
		}
	}

	// --- NEW: Step 2 - Delete stale containers for this host ---
	// If there were no active containers, the slice will be empty, and this will
	// correctly delete all containers associated with the host.
	deleteParams := db.DeleteStaleContainersForHostParams{
		HostID:  host.ID,
		Column2: activeContainerUIDs,
	}
	// Skip deletion if no active containers reported (avoid deleting everything when agent temporarily reports none)
	if len(activeContainerUIDs) > 0 {
		err = s.db.DeleteStaleContainersForHost(ctx, deleteParams)
		if err != nil {
			log.Printf("Failed to delete stale containers: %v", err)
			return &orchestrator.HeartbeatResponse{Success: false, Message: "Heartbeat processed, but failed to clean up stale containers: " + err.Error()}, nil
		}
	}
	if err != nil {
		log.Printf("Failed to delete stale containers: %v", err)
		// This is a significant issue, but the main heartbeat has been processed.
		// You might want to handle this error more robustly.
		return &orchestrator.HeartbeatResponse{
			Success: false,
			Message: "Heartbeat processed, but failed to clean up stale containers: " + err.Error(),
		}, nil
	}

	log.Printf("Successfully synced %d containers for host %s", len(activeContainerUIDs), req.MacAddress)
	return &orchestrator.HeartbeatResponse{
		Success: true,
		Message: "Heartbeat received and containers synced successfully",
	}, nil
}

func (s *Server) ConnectAgentStream(stream orchestrator.HostAgentService_ConnectAgentStreamServer) error {
	firstMsg, err := stream.Recv()
	if err != nil {
		log.Printf("Failed to receive initial message from stream: %v", err)
		return err
	}
	agentID := firstMsg.MacAddress
	log.Printf("Agent connected with ID: %s", agentID)
	// Store the stream in the map with mutex protection
	s.mu.Lock()
	s.Hosts[agentID] = stream
	s.mu.Unlock()

	// Keep reading for messages from agent (like UpdateStatus) and update DB accordingly
	for {
		select {
		case <-stream.Context().Done():
			log.Printf("stream context done for agent %s: %v", agentID, stream.Context().Err())
			s.mu.Lock()
			delete(s.Hosts, agentID)
			s.mu.Unlock()
			return stream.Context().Err()
		default:
		}
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("stream closed from agent %s: %v", agentID, err)
			s.mu.Lock()
			delete(s.Hosts, agentID)
			s.mu.Unlock()
			return err
		}
		log.Printf("Received message from agent %s: %+v", agentID, msg)
		// Here you can process UpdateStatus messages and update the database
		// For example, update the status of the deployment/container in the db
	}
}
