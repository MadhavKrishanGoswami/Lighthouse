package tuiserver

import (
	"context"
	"fmt"
	"log"

	"sync"
	"sync/atomic"
	"time"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/monitor"
	servicestatus "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/serviceStatus"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Server implements the TUI gRPC service and manages active streams.
type Server struct {
	TUILogsBufferSize int
	tui.UnimplementedTUIServiceServer
	DB      *db.Queries
	mu      sync.Mutex
	streams map[string]tui.TUIService_SendDatastreamServer
}

// NewServer creates a new TUI server instance.
func NewServer(queries *db.Queries) *Server {
	return &Server{
		TUILogsBufferSize: 50,
		DB:                queries,
		streams:           make(map[string]tui.TUIService_SendDatastreamServer),
	}
}

// AddStream adds a stream to the server.
func (s *Server) AddStream(stream tui.TUIService_SendDatastreamServer) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New().String()
	s.streams[id] = stream
	log.Printf("[TUI Service] Stream added: %s. Total streams: %d", id, len(s.streams))
	return id
}

// RemoveStream removes a stream from the server.
func (s *Server) RemoveStream(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.streams[id]; ok {
		delete(s.streams, id)
		log.Printf("[TUI Service] Stream removed: %s. Total streams: %d", id, len(s.streams))
	}
}

// Broadcast sends a message to all connected streams.
func (s *Server) Broadcast(message *tui.DataStreamSend) {
	s.mu.Lock()
	streams := make(map[string]tui.TUIService_SendDatastreamServer, len(s.streams))
	for id, stream := range s.streams {
		streams[id] = stream
	}
	s.mu.Unlock()

	for id, stream := range streams {
		if err := stream.Send(message); err != nil {
			log.Printf("[TUI Service] Failed to send to stream %s: %v. Removing stream.", id, err)
			s.RemoveStream(id)
		} else {
			log.Printf("[TUI Service] Message sent to stream %s", id)
		}
	}
}

// GracefulShutdown closes all active streams.
func (s *Server) GracefulShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, stream := range s.streams {
		if err := stream.Send(&tui.DataStreamSend{
			Logs: "Server is shutting down. Closing connection.",
		}); err != nil {
			log.Printf("[TUI Service] Error notifying stream %s of shutdown: %v", id, err)
		}
		s.RemoveStream(id)
	}
	log.Println("[TUI Service] All streams closed.")
}

// SendDatastream handles the bidirectional streaming RPC.
func (s *Server) SendDatastream(stream tui.TUIService_SendDatastreamServer) error {
	var heartbeatCount uint64
	// Simplified logging: keep only the last log line (most recent event) instead of a slice.
	var lastLog atomic.Value // stores string
	setLog := func(line string) {
		if line == "" {
			return
		}
		lastLog.Store(fmt.Sprintf("%s | %s", time.Now().Format(time.RFC3339), line))
	}
	getLog := func() string {
		v := lastLog.Load()
		if v == nil {
			return ""
		}
		return v.(string)
	}
	_ = getLog
	setLog("client stream connected") // initial log
	clientID := s.AddStream(stream)
	defer s.RemoveStream(clientID)
	log.Printf("[TUI Service] Client connected: %s", clientID)

	snapshotReq := make(chan string, 8) // reason channel

	// Goroutine to read acks / heartbeats from client
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				setLog(fmt.Sprintf("recv loop ended: %v", err)) // record final state
				log.Printf("[TUI Service] recv loop ended for %s: %v", clientID, err)
				return
			}
			if ack := in.GetAck(); ack != "" {
				atomic.AddUint64(&heartbeatCount, 1)
				setLog(fmt.Sprintf("heartbeat %d (ack=%s)", atomic.LoadUint64(&heartbeatCount), ack))
				select { // non-blocking push
				case snapshotReq <- "heartbeat":
				default:
				}
			}
		}
	}()

	sendSnapshot := func(reason string) error {
		// Build host list with embedded containers
		hostRows, err := s.DB.GetAllHosts(context.Background())
		if err != nil {
			return fmt.Errorf("fetch hosts: %w", err)
		}
		// For each host fetch containers (N+1). Acceptable for small scale; optimize later with join.
		containerRows := make(map[string][]db.Container)
		for _, h := range hostRows {
			rows, err := s.DB.GetAllContainersonHost(context.Background(), h.MacAddress)
			if err != nil {
				log.Printf("[TUI Service] fetch containers for host %s: %v", h.MacAddress, err)
				continue
			}
			containerRows[h.MacAddress] = rows
		}
		// group containers by host mac
		contByHost := make(map[string][]*tui.ContainerInfo)
		for _, h := range hostRows {
			rows := containerRows[h.MacAddress]
			for _, c := range rows {
				ci := &tui.ContainerInfo{Name: c.Name, Image: c.Image, Status: 0 /* no status col yet */, Watch: c.Watch.Bool}
				contByHost[h.MacAddress] = append(contByHost[h.MacAddress], ci)
			}
		}
		hostInfos := make([]*tui.HostInfo, 0, len(hostRows))
		for _, h := range hostRows {
			lastHB := ""
			if h.LastHeartbeat.Valid {
				lastHB = h.LastHeartbeat.Time.Format(time.RFC3339)
			}
			hostInfos = append(hostInfos, &tui.HostInfo{
				MacAddress:    h.MacAddress,
				Hostname:      h.Hostname,
				IpAddress:     h.IpAddress,
				LastHeartbeat: lastHB,
				Containers:    contByHost[h.MacAddress],
			})
		}
		orchestratorUp, dbUp, registryUp := servicestatus.GetServiceStatus()
		servicesStatus := []*tui.ServicesStatus{
			{ServicesStatus: tui.ServicesStatus_ORCHESTRATOR, Status: orchestratorUp},
			{ServicesStatus: tui.ServicesStatus_Database, Status: dbUp},
			{ServicesStatus: tui.ServicesStatus_REGISTRY_Monitor, Status: registryUp},
		}
		msg := &tui.DataStreamSend{
			HostList:       &tui.HostList{Hosts: hostInfos},
			Logs:           fmt.Sprintf("%s\n%s", getLog(), fmt.Sprintf("snapshot reason=%s hosts=%d", reason, len(hostInfos))),
			CronTime:       int32(monitor.GetCronTimeInHours()),
			ServicesStatus: servicesStatus,
		}
		setLog(fmt.Sprintf("snapshot sent reason=%s hosts=%d", reason, len(hostInfos)))
		return stream.Send(msg)
	}

	if err := sendSnapshot("initial"); err != nil {
		log.Printf("[TUI Service] initial snapshot error: %v", err)
	}
	// periodic updates + heartbeat-triggered snapshots
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := sendSnapshot("ticker"); err != nil {
				setLog(fmt.Sprintf("snapshot error: %v", err))
				log.Printf("[TUI Service] snapshot send error: %v", err)
			}
		case reason := <-snapshotReq:
			if err := sendSnapshot(reason); err != nil {
				setLog(fmt.Sprintf("snapshot error: %v", err))
				log.Printf("[TUI Service] snapshot send error: %v", err)
			}
		}
	}
}

func (s *Server) SetWatch(ctx context.Context, req *tui.SetWatchlistRequest) (*tui.SetWatchlistResponse, error) {
	log.Printf("[TUI Service] SetWatch request: container=%s, host=%s, watch=%v",
		req.GetContainerName(), req.GetHostMac(), req.GetWatch())

	prams := db.SetWatchStatusParams{
		Name:       req.GetContainerName(),
		MacAddress: req.GetHostMac(),
		Watch:      boolToPgtype(req.GetWatch()),
	}

	if err := s.DB.SetWatchStatus(ctx, prams); err != nil {
		log.Printf("[TUI Service] Error setting watch: %v", err)
		return &tui.SetWatchlistResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to set watch: %v", err),
		}, err
	}

	return &tui.SetWatchlistResponse{
		Success: true,
		Message: fmt.Sprintf("Watch set for container %s to %v", req.GetContainerName(), req.GetWatch()),
	}, nil
}

// SetCronTime updates the cron schedule.
func (s *Server) SetCronTime(ctx context.Context, req *tui.SetCronTimeRequest) (*tui.SetCronTimeResponse, error) {
	log.Printf("[TUI Service] SetCronTime request: %d hours", req.GetCronTime())
	monitor.SetCronTimeInHours(int(req.GetCronTime()))
	return &tui.SetCronTimeResponse{
		Success: true,
		Message: fmt.Sprintf("Cron time set to %d hours", req.GetCronTime()),
	}, nil
}

// Helper to convert bool to pgtype.Bool
func boolToPgtype(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}
