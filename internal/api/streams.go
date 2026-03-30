package api

import (
	"log/slog"
	"sync"
)

// RunningStreams tracks which chat IDs currently have an in-flight
// LLM stream. Used to prevent duplicate concurrent streams.
type RunningStreams struct {
	mu      sync.Mutex
	running map[string]bool
	logger  *slog.Logger
}

// NewRunningStreams creates a new tracker.
func NewRunningStreams(logger *slog.Logger) *RunningStreams {
	return &RunningStreams{
		running: make(map[string]bool),
		logger:  logger,
	}
}

// TryStart marks a chat as running. Returns false if already running.
func (r *RunningStreams) TryStart(chatID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running[chatID] {
		return false
	}
	r.running[chatID] = true
	r.logger.Info("stream_started", "chat_id", chatID)
	return true
}

// Done marks a chat as no longer running.
func (r *RunningStreams) Done(chatID string) {
	r.mu.Lock()
	delete(r.running, chatID)
	r.mu.Unlock()
	r.logger.Info("stream_done", "chat_id", chatID)
}
