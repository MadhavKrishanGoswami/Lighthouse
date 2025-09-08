package logstream

import (
	"sync"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	tuiserver "github.com/MadhavKrishanGoswami/Lighthouse/services/orchestrator/internal/grpc/tui"
)

// simple indirection so other packages do not need direct server import beyond this package
var (
	mu     sync.RWMutex
	server *tuiserver.Server
)

// SetServer wires the TUI server so logs can be pushed.
func SetServer(s *tuiserver.Server) {
	mu.Lock()
	server = s
	mu.Unlock()
}

// Push emits a formatted line to the TUI log stream if server present.
func Push(format string, args ...any) {
	mu.RLock()
	s := server
	mu.RUnlock()
	if s == nil {
		return
	}
	s.PushLog(format, args...)
}

// Convenience to expose the proto type if needed externally (unused now)
var _ = tui.LogLine{}
