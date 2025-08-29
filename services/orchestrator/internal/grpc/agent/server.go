package agentserver

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/host-agents"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// AgentConnection holds the stream and a dedicated channel for safely sending commands.
type AgentConnection struct {
	Stream      orchestrator.HostAgentService_ConnectAgentStreamServer
	CommandChan chan *orchestrator.UpdateContainerCommand
	done        chan struct{}
}

// Server is the main gRPC server structure.
type Server struct {
	orchestrator.UnimplementedHostAgentServiceServer
	DB    *db.Queries
	Mu    sync.RWMutex
	Hosts map[string]*AgentConnection
}

// NewServer creates a new instance of the gRPC server.
func NewServer(queries *db.Queries) *Server {
	return &Server{
		DB:    queries,
		Hosts: make(map[string]*AgentConnection),
	}
}

// SendCommand sends a command to a connected agent safely.
func (s *Server) SendCommand(agentID string, cmd *orchestrator.UpdateContainerCommand) error {
	s.Mu.RLock()
	conn, ok := s.Hosts[agentID]
	s.Mu.RUnlock()
	if !ok {
		return fmt.Errorf("agent %s not connected or disconnected", agentID)
	}

	select {
	case conn.CommandChan <- cmd:
		log.Printf("Queued update command for agent %s image %s", agentID, cmd.Image)
		return nil
	case <-conn.done:
		return fmt.Errorf("agent %s disconnected, cannot send command", agentID)
	}
}

// RegisterHost handles initial registration of a host and its containers.
func (s *Server) RegisterHost(ctx context.Context, req *orchestrator.RegisterHostRequest) (*orchestrator.RegisterHostResponse, error) {
	if req == nil || req.Host == nil {
		return &orchestrator.RegisterHostResponse{Success: false, Message: "invalid request: host is nil"}, nil
	}

	log.Printf("Host registration: %s (%s)", req.Host.Hostname, req.Host.IpAddress)

	params := db.InsertHostParams{
		MacAddress: req.Host.MacAddress,
		Hostname:   req.Host.Hostname,
		IpAddress:  req.Host.IpAddress,
	}
	host, err := s.DB.InsertHost(ctx, params)
	if err != nil {
		log.Printf("Register host failed: %v", err)
		return &orchestrator.RegisterHostResponse{Success: false, Message: err.Error()}, nil
	}

	for _, container := range req.Host.Containers {
		containerParams := db.InsertContainerParams{
			ContainerUid: container.ContainerID,
			HostID:       host.ID,
			Name:         container.Name,
			Image:        container.Image,
			Ports:        convertPortsToDBFormat(container.Ports),
			EnvVars:      container.EnvVars,
			Volumes:      container.Volumes,
			Network:      pgtype.Text{String: container.Network, Valid: true},
		}
		if _, err := s.DB.InsertContainer(ctx, containerParams); err != nil {
			log.Printf("Register container %s failed: %v", container.Name, err)
		}
	}

	return &orchestrator.RegisterHostResponse{
		Success: true,
		Message: "Host registered successfully",
	}, nil
}

// Heartbeat processes periodic updates from agents.
func (s *Server) Heartbeat(ctx context.Context, req *orchestrator.HeartbeatRequest) (*orchestrator.HeartbeatResponse, error) {
	log.Printf("Heartbeat from host %s", req.MacAddress)
	host, err := s.DB.GetHostByMacAddress(ctx, req.MacAddress)
	if err != nil {
		return &orchestrator.HeartbeatResponse{Success: false, Message: err.Error()}, nil
	}

	if _, err := s.DB.UpdateHostLastHeartbeat(ctx, host.ID); err != nil {
		log.Printf("Update heartbeat failed: %v", err)
	}

	activeContainerUIDs := make([]string, 0, len(req.Containers))
	for _, c := range req.Containers {
		activeContainerUIDs = append(activeContainerUIDs, c.ContainerID)
		containerParams := db.InsertContainerParams{
			ContainerUid: c.ContainerID,
			HostID:       host.ID,
			Name:         c.Name,
			Image:        c.Image,
			Ports:        convertPortsToDBFormat(c.Ports),
			EnvVars:      c.EnvVars,
			Volumes:      c.Volumes,
			Network:      pgtype.Text{String: c.Network, Valid: true},
		}
		if _, err := s.DB.InsertContainer(ctx, containerParams); err != nil {
			log.Printf("Upsert container %s failed: %v", c.Name, err)
		}
	}

	// Delete stale containers
	if len(activeContainerUIDs) > 0 {
		params := db.DeleteStaleContainersForHostParams{
			HostID:  host.ID,
			Column2: activeContainerUIDs,
		}
		if err := s.DB.DeleteStaleContainersForHost(ctx, params); err != nil {
			log.Printf("Delete stale containers failed: %v", err)
			return &orchestrator.HeartbeatResponse{Success: false, Message: err.Error()}, nil
		}
	}

	log.Printf("Synced %d containers host %s", len(activeContainerUIDs), req.MacAddress)
	return &orchestrator.HeartbeatResponse{Success: true, Message: "Heartbeat processed successfully"}, nil
}

// ConnectAgentStream handles bidirectional agent streams.
func (s *Server) ConnectAgentStream(stream orchestrator.HostAgentService_ConnectAgentStreamServer) error {
	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}
	agentID := firstMsg.GetMacAddress()
	if agentID == "" {
		return fmt.Errorf("empty agent ID")
	}

	log.Printf("Agent stream connected %s", agentID)
	conn := &AgentConnection{
		Stream:      stream,
		CommandChan: make(chan *orchestrator.UpdateContainerCommand, 10),
		done:        make(chan struct{}),
	}

	s.Mu.Lock()
	s.Hosts[agentID] = conn
	s.Mu.Unlock()

	defer func() {
		s.Mu.Lock()
		delete(s.Hosts, agentID)
		s.Mu.Unlock()
		close(conn.done)
		log.Printf("Agent stream disconnected %s", agentID)
	}()

	// Write loop
	go func() {
		for {
			select {
			case cmd := <-conn.CommandChan:
				if err := stream.Send(cmd); err != nil {
					log.Printf("Send command to agent %s failed: %v", agentID, err)
					return
				}
			case <-conn.done:
				return
			}
		}
	}()

	// Read loop
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Agent stream closed by %s", agentID)
			return nil
		}
		if err != nil {
			return fmt.Errorf("agent stream recv failed: %w", err)
		}

		status := msg.GetStage()
		host, _ := s.DB.GetHostByMacAddress(context.Background(), agentID)

		_, err = s.DB.InsertUpdateStatus(context.Background(), db.InsertUpdateStatusParams{
			HostID: host.ID,
			Stage:  grpcenumtodbstatus(status),
			Logs:   pgtype.Text{String: strings.TrimSpace(msg.GetLogs()), Valid: true},
			Image:  msg.Image,
		})
		if err != nil {
			log.Printf("Insert update status failed: %v", err)
		}
	}
}

// convertPortsToDBFormat converts []*PortMapping to DB-storable []string.
func convertPortsToDBFormat(ports []*orchestrator.PortMapping) []string {
	var out []string
	for _, p := range ports {
		out = append(out, fmt.Sprintf("%s:%d->%d/%s", p.HostIp, p.HostPort, p.ContainerPort, p.Protocol))
	}
	return out
}

// grpcenumtodbstatus maps gRPC status to DB enum
func grpcenumtodbstatus(status orchestrator.UpdateStatus_Stage) db.UpdateStage {
	switch status {
	case orchestrator.UpdateStatus_STARTING:
		return db.UpdateStageStarting
	case orchestrator.UpdateStatus_PULLING:
		return db.UpdateStagePulling
	case orchestrator.UpdateStatus_RUNNING:
		return db.UpdateStageRunning
	case orchestrator.UpdateStatus_ROLLBACK:
		return db.UpdateStageRollback
	case orchestrator.UpdateStatus_COMPLETED:
		return db.UpdateStageCompleted
	case orchestrator.UpdateStatus_FAILED, orchestrator.UpdateStatus_UNKNOWN:
		return db.UpdateStageFailed
	default:
		return db.UpdateStageFailed
	}
}
