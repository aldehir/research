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

func TestStreamRegistry_StartAndGet(t *testing.T) {
	reg := NewStreamRegistry(slog.Default())

	stream, ctx := reg.Start("chat-1")
	require.NotNil(t, stream)
	require.NotNil(t, ctx)
	assert.Equal(t, "chat-1", stream.ChatID)
	assert.Equal(t, StreamRunning, stream.Status())

	got := reg.Get("chat-1")
	assert.Equal(t, stream, got)

	assert.Nil(t, reg.Get("nonexistent"))
}

func TestStreamRegistry_AppendAndEventsSince(t *testing.T) {
	reg := NewStreamRegistry(slog.Default())
	stream, _ := reg.Start("chat-1")

	reg.Append(stream, SSEEvent{Data: `{"type":"delta","text":"Hello"}`})
	reg.Append(stream, SSEEvent{Data: `{"type":"delta","text":" world"}`})
	reg.Append(stream, SSEEvent{Data: `{"type":"done"}`})

	// Read all events from the beginning
	events, done := stream.EventsSince(0)
	assert.Len(t, events, 3)
	assert.False(t, done)

	// Read events from offset 2
	events, done = stream.EventsSince(2)
	assert.Len(t, events, 1)
	assert.Equal(t, `{"type":"done"}`, events[0].Data)
	assert.False(t, done)

	// Mark stream as done and verify
	stream.SetStatus(StreamDone)
	events, done = stream.EventsSince(3)
	assert.Len(t, events, 0)
	assert.True(t, done, "should report done when stream is complete and no more events")
}

func TestStreamRegistry_Notify(t *testing.T) {
	reg := NewStreamRegistry(slog.Default())
	stream, _ := reg.Start("chat-1")

	// Get the notify channel before appending
	ch := stream.Notify()

	// Append an event — should close the notify channel
	reg.Append(stream, SSEEvent{Data: `{"type":"delta","text":"Hi"}`})

	// The channel should be readable (closed)
	select {
	case <-ch:
		// expected
	default:
		t.Fatal("notify channel should have been closed after Append")
	}
}

func TestStreamRegistry_Remove(t *testing.T) {
	reg := NewStreamRegistry(slog.Default())
	stream, _ := reg.Start("chat-1")
	require.NotNil(t, reg.Get("chat-1"))

	reg.Remove("chat-1")
	assert.Nil(t, reg.Get("chat-1"))

	// Verify context was cancelled
	assert.Equal(t, StreamRunning, stream.Status()) // status unchanged by Remove
}

// slowStreamer sends events with a controllable gate so the test can
// cancel the HTTP context mid-stream. Unlike mockStreamer, it respects
// context cancellation (simulating real provider behavior).
type slowStreamer struct {
	events []chat.StreamEvent
	gate   chan struct{} // close to release events
}

func (s *slowStreamer) Stream(ctx context.Context, _ chat.Request) (<-chan chat.StreamEvent, error) {
	ch := make(chan chat.StreamEvent)
	go func() {
		defer close(ch)
		select {
		case <-s.gate: // wait for test to release
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
	// Limit to 1 connection so the background goroutine shares the same
	// in-memory SQLite instance as the test assertions.
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

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

	// Create a cancellable request to simulate client disconnect
	ctx, cancel := context.WithCancel(context.Background())
	body := `{"content":"Hi"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Serve in a goroutine
	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec, req)
		close(done)
	}()

	// Give the handler time to start, then cancel the client connection
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Now release the streamer — events should still be processed in background
	close(gate)

	// Wait for the handler to return (it should return quickly after cancel)
	<-done

	// Wait for the background goroutine to finish persisting
	// Poll the DB for the assistant message
	var found bool
	for i := 0; i < 50; i++ {
		msgs, _ := store.ListMessages(tdb.DB, "chat-1")
		for _, m := range msgs {
			if m.Role == "assistant" && m.Content == "Hello world" {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.True(t, found, "assistant message should be persisted even after client disconnect")
}

func TestReconnectStream_NotFound(t *testing.T) {
	tdb := store.NewTestDB(t)
	storage := pdf.NewStorage(t.TempDir())
	mux := NewMux(tdb.DB, storage, nil, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

	req := httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats/nonexistent/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestReconnectStream_ReplaysAllEvents(t *testing.T) {
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

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(5*time.Second))

	// Start the initial request
	body := `{"content":"Hi"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	ctx1, cancel1 := context.WithCancel(context.Background())
	req1 = req1.WithContext(ctx1)
	rec1 := httptest.NewRecorder()

	done1 := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec1, req1)
		close(done1)
	}()

	// Wait for the stream to register, then disconnect first client
	time.Sleep(50 * time.Millisecond)
	cancel1()
	<-done1

	// Now release the gate — background goroutine processes events
	close(gate)

	// Wait for background to complete
	time.Sleep(100 * time.Millisecond)

	// Reconnect — should replay all events
	req2 := httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats/chat-1/stream", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "text/event-stream", rec2.Header().Get("Content-Type"))

	events := parseSSEEvents(t, rec2.Body.String())
	require.GreaterOrEqual(t, len(events), 3)

	var deltas []string
	var hasDone bool
	for _, ev := range events {
		if ev.Type == "delta" {
			deltas = append(deltas, ev.Text)
		}
		if ev.Type == "done" {
			hasDone = true
		}
	}
	assert.Equal(t, []string{"Hello", " world"}, deltas)
	assert.True(t, hasDone, "reconnect should replay done event")
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

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

	// First request — starts a stream
	body1 := `{"content":"First message"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec1, req1)
		close(done)
	}()

	// Wait for the first request to register the stream
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
