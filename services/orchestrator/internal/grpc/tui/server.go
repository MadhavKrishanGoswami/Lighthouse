package tuiserver

import (
	"context"
	"fmt"
	"log"
	"sync"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	db "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/db/sqlc"
	monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/monitor"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TUIConnection struct {
	mu      sync.Mutex
	Streams map[string]tui.TUIService_SendDatastreamServer
}

// NewServer creates a new instance of the gRPC server.
func NewStreamManager() *TUIConnection {
	return &TUIConnection{
		Streams: make(map[string]tui.TUIService_SendDatastreamServer),
	}
}

// Add adds a new stream to the manager and returns a unique ID for it.
func (sm *TUIConnection) Add(stream tui.TUIService_SendDatastreamServer) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	id := uuid.New().String()
	sm.Streams[id] = stream
	log.Printf("[TUI Service] Added new stream with ID: %s. Total streams: %d", id, len(sm.Streams))
	return id
}

// Remove removes a stream from the manager using its ID.
func (sm *TUIConnection) Remove(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.Streams[id]; exists {
		delete(sm.Streams, id)
		log.Printf("[TUI Service] Removed stream with ID: %s. Total streams: %d", id, len(sm.Streams))
	} else {
		log.Printf("[TUI Service] Attempted to remove non-existent stream with ID: %s", id)
	}
}

// Broadcast sends a message to all active streams.
func (sm *TUIConnection) Broadcast(message *tui.DataStreamSend) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for id, stream := range sm.Streams {
		if err := stream.Send(message); err != nil {
			log.Printf("[TUI Service] Error sending message to stream ID %s: %v. Removing stream.", id, err)
			delete(sm.Streams, id)
		} else {
			log.Printf("[TUI Service] Sent message to stream ID %s", id)
		}
	}
}

type Server struct {
	tui.UnimplementedTUIServiceServer
	DB *db.Queries
}

func NewServer(queries *db.Queries) *Server {
	return &Server{
		DB: queries,
	}
}

// SetWatch is a unary RPC that allows a client to set a watch on a container.
func (s *Server) SetWatch(ctx context.Context, req *tui.SetWatchlistRequest) (*tui.SetWatchlistResponse, error) {
	log.Printf("[TUI Service] Received SetWatch request for container '%s' on host '%s'. Watch status: %v", req.GetContainerName(), req.GetHostMac(), req.GetWatch())
	// convert bool to db compatible type
	pgtypebool := pgtype.Bool{Bool: req.GetWatch(), Valid: true}

	prams := db.SetWatchStatusParams{
		Watch:      pgtypebool,
		Name:       req.GetContainerName(),
		MacAddress: req.GetHostMac(),
	}
	if err := s.DB.SetWatchStatus(ctx, prams); err != nil {
		log.Printf("[TUI Service] Error setting watch status: %v", err)
		return &tui.SetWatchlistResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to set watch status for %s: %v", req.GetContainerName(), err),
		}, err
	}

	return &tui.SetWatchlistResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully set watch status for %s to %v", req.GetContainerName(), req.GetWatch()),
	}, nil
}

// SetCronTime is a unary RPC that allows a client to set a cron time.
func (s *Server) SetCronTime(ctx context.Context, req *tui.SetCronTimeRequest) (*tui.SetCronTimeResponse, error) {
	log.Printf("[TUI Service] Received SetCronTime request. New time (hours): %d", req.GetCronTime())
	// Update the orchestrator cron schedule immediately
	monitor.SetCronTimeInHours(int(req.GetCronTime()))
	return &tui.SetCronTimeResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully set cron time to %d hour(s)", req.GetCronTime()),
	}, nil
}
