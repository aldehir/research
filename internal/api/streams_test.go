package api

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aldehir/research/internal/chat"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunningStreams_TryStartAndDone(t *testing.T) {
	rs := NewRunningStreams(slog.Default())

	assert.True(t, rs.TryStart("chat-1"), "first start should succeed")
	assert.False(t, rs.TryStart("chat-1"), "duplicate start should fail")
	assert.True(t, rs.TryStart("chat-2"), "different chat should succeed")

	rs.Done("chat-1")
	assert.True(t, rs.TryStart("chat-1"), "start after done should succeed")
}

// slowStreamer sends events with a controllable gate so the test can
// cancel the HTTP context mid-stream. It respects context cancellation
// (simulating real provider behavior).
type slowStreamer struct {
	events []chat.StreamEvent
	gate   chan struct{} // close to release events
}

func (s *slowStreamer) Stream(ctx context.Context, _ chat.Request) (<-chan chat.StreamEvent, error) {
	ch := make(chan chat.StreamEvent)
	go func() {
		defer close(ch)
		select {
		case <-s.gate:
		case <-ctx.Done():
			return
		}
		for _, ev := range s.events {
			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func TestSendMessage_BackgroundCompletion(t *testing.T) {
	tdb := store.NewTestDB(t)
	tdb.DB.SetMaxOpenConns(1)
	storage := pdf.NewStorage(t.TempDir())

	p := store.Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  "/test.pdf",
		FileSize:  1000,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, store.CreatePaper(tdb.DB, p))

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	gate := make(chan struct{})
	mock := &slowStreamer{
		events: []chat.StreamEvent{
			{Kind: chat.EventDelta, Text: "Hello"},
			{Kind: chat.EventDelta, Text: " world"},
			{Kind: chat.EventDone},
		},
		gate: gate,
	}

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default())

	// Create a cancellable request to simulate client disconnect
	ctx, cancel := context.WithCancel(context.Background())
	body := `{"content":"Hi"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Release the gate immediately — the handler uses context.Background()
	// for the provider, so the stream proceeds even after cancel.
	close(gate)

	// Serve in a goroutine, cancel client mid-flight
	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec, req)
		close(done)
	}()

	// Cancel the HTTP context — handler goroutine continues because the
	// provider stream uses context.Background().
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for handler to finish (it runs to completion)
	<-done

	// The assistant message should be persisted
	msgs, err := store.ListMessages(tdb.DB, "chat-1")
	require.NoError(t, err)

	var found bool
	for _, m := range msgs {
		if m.Role == "assistant" && m.Content == "Hello world" {
			found = true
			break
		}
	}
	assert.True(t, found, "assistant message should be persisted even after client disconnect")
}

func TestSendMessage_ConflictOnDuplicateStream(t *testing.T) {
	tdb := store.NewTestDB(t)
	storage := pdf.NewStorage(t.TempDir())

	p := store.Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  "/test.pdf",
		FileSize:  1000,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, store.CreatePaper(tdb.DB, p))

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	gate := make(chan struct{})
	mock := &slowStreamer{
		events: []chat.StreamEvent{
			{Kind: chat.EventDelta, Text: "Hello"},
			{Kind: chat.EventDone},
		},
		gate: gate,
	}

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default())

	// First request — starts a stream (gate stays closed so it blocks)
	body1 := `{"content":"First message"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec1, req1)
		close(done)
	}()

	// Wait for the first request to register as running
	time.Sleep(50 * time.Millisecond)

	// Second request to the same chat — should get 409
	body2 := `{"content":"Second message"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusConflict, rec2.Code)

	// Clean up: release the gate so first request finishes
	close(gate)
	<-done
}
