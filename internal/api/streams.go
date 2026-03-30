package api

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// StreamStatus tracks the lifecycle of a background stream.
type StreamStatus string

const (
	StreamRunning StreamStatus = "running"
	StreamDone    StreamStatus = "done"
	StreamError   StreamStatus = "error"
)

// SSEEvent is a buffered SSE event (the JSON-encoded data line).
type SSEEvent struct {
	Data string
}

// ActiveStream tracks one background LLM response.
type ActiveStream struct {
	ChatID string

	mu     sync.Mutex
	status StreamStatus
	events []SSEEvent
	notify chan struct{}
	cancel context.CancelFunc
}

// Status returns the current stream status.
func (s *ActiveStream) Status() StreamStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// StreamRegistry maps chatID -> *ActiveStream.
type StreamRegistry struct {
	mu           sync.Mutex
	streams      map[string]*ActiveStream
	logger       *slog.Logger
	RetentionTTL time.Duration
}

// NewStreamRegistry creates a new stream registry.
func NewStreamRegistry(logger *slog.Logger) *StreamRegistry {
	return &StreamRegistry{
		streams:      make(map[string]*ActiveStream),
		logger:       logger,
		RetentionTTL: 60 * time.Second,
	}
}

// Start creates a new ActiveStream for the given chat ID and returns a
// background context that will be cancelled when the stream is removed.
func (r *StreamRegistry) Start(chatID string) (*ActiveStream, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := &ActiveStream{
		ChatID: chatID,
		status: StreamRunning,
		notify: make(chan struct{}),
		cancel: cancel,
	}
	r.mu.Lock()
	r.streams[chatID] = stream
	r.mu.Unlock()
	r.logger.Info("stream_started", "chat_id", chatID)
	return stream, ctx
}

// Get returns the active stream for the given chat ID, or nil.
func (r *StreamRegistry) Get(chatID string) *ActiveStream {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.streams[chatID]
}

// Remove removes a stream from the registry and cancels its context.
func (r *StreamRegistry) Remove(chatID string) {
	r.mu.Lock()
	stream, ok := r.streams[chatID]
	delete(r.streams, chatID)
	r.mu.Unlock()
	if ok {
		stream.cancel()
		r.logger.Info("stream_removed", "chat_id", chatID)
	}
}

// Append adds an SSE event to the stream buffer and notifies waiters.
func (r *StreamRegistry) Append(stream *ActiveStream, event SSEEvent) {
	stream.mu.Lock()
	stream.events = append(stream.events, event)
	ch := stream.notify
	stream.notify = make(chan struct{})
	stream.mu.Unlock()
	close(ch)
}

// SetStatus updates the stream status.
func (s *ActiveStream) SetStatus(status StreamStatus) {
	s.mu.Lock()
	s.status = status
	ch := s.notify
	s.notify = make(chan struct{})
	s.mu.Unlock()
	close(ch)
}

// EventsSince returns events starting from offset. The second return value
// is true when the stream is complete and there are no more events to read.
func (s *ActiveStream) EventsSince(offset int) ([]SSEEvent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if offset >= len(s.events) {
		done := s.status == StreamDone || s.status == StreamError
		return nil, done
	}
	return s.events[offset:], false
}

// Notify returns a channel that is closed when new events are available.
func (s *ActiveStream) Notify() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.notify
}
