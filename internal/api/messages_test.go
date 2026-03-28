package api

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aldehir/research/internal/anthropic"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStreamer struct {
	events []anthropic.StreamEvent
	err    error
}

func (m *mockStreamer) Stream(_ context.Context, _ anthropic.Request) (<-chan anthropic.StreamEvent, error) {
	if m.err != nil {
		return nil, m.err
	}
	ch := make(chan anthropic.StreamEvent)
	go func() {
		defer close(ch)
		for _, ev := range m.events {
			ch <- ev
		}
	}()
	return ch, nil
}

// captureStreamer records the request passed to Stream.
type captureStreamer struct {
	mockStreamer
	captured *anthropic.Request
}

func (c *captureStreamer) Stream(ctx context.Context, req anthropic.Request) (<-chan anthropic.StreamEvent, error) {
	c.captured = &req
	return c.mockStreamer.Stream(ctx, req)
}

func seedChatSession(t *testing.T, tdb *store.TestDB) store.ChatSession {
	t.Helper()
	seedTestPaper(t, tdb)
	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))
	return session
}

type sseEvent struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Error string `json:"error,omitempty"`
}

func parseSSEEvents(t *testing.T, body string) []sseEvent {
	t.Helper()
	var events []sseEvent
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var ev sseEvent
		require.NoError(t, json.Unmarshal([]byte(data), &ev))
		events = append(events, ev)
	}
	return events
}

func TestSendMessage(t *testing.T) {
	t.Run("streams SSE response with delta and done events", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &mockStreamer{
			events: []anthropic.StreamEvent{
				{Type: "content_block_delta", Text: "Hello"},
				{Type: "content_block_delta", Text: " world"},
				{Type: "message_stop"},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"What does this mean?"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))

		events := parseSSEEvents(t, rec.Body.String())
		require.Len(t, events, 3)
		assert.Equal(t, "delta", events[0].Type)
		assert.Equal(t, "Hello", events[0].Text)
		assert.Equal(t, "delta", events[1].Type)
		assert.Equal(t, " world", events[1].Text)
		assert.Equal(t, "done", events[2].Type)
	})

	t.Run("stores user message in DB before streaming", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &mockStreamer{
			events: []anthropic.StreamEvent{
				{Type: "content_block_delta", Text: "Reply"},
				{Type: "message_stop"},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"User question"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		messages, err := store.ListMessages(tdb.DB, "chat-1")
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(messages), 1)
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "User question", messages[0].Content)
	})

	t.Run("stores assistant message in DB after streaming", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &mockStreamer{
			events: []anthropic.StreamEvent{
				{Type: "content_block_delta", Text: "Hello"},
				{Type: "content_block_delta", Text: " world"},
				{Type: "message_stop"},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"Hi"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		messages, err := store.ListMessages(tdb.DB, "chat-1")
		require.NoError(t, err)
		require.Len(t, messages, 2)
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "assistant", messages[1].Role)
		assert.Equal(t, "Hello world", messages[1].Content)
	})

	t.Run("chat session not found returns 404", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedTestPaper(t, tdb)

		mock := &mockStreamer{}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/missing/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("missing content returns 400", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &mockStreamer{}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("anthropic stream error returns error SSE event", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &mockStreamer{
			err: assert.AnError,
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

		events := parseSSEEvents(t, rec.Body.String())
		require.Len(t, events, 1)
		assert.Equal(t, "error", events[0].Type)
		assert.NotEmpty(t, events[0].Error)
	})

	t.Run("chat streamer nil returns 503", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mux := testMuxWithChat(t, tdb, nil)

		body := `{"content":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("passes selected_text and surrounding_text to anthropic", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []anthropic.StreamEvent{
					{Type: "content_block_delta", Text: "Ok"},
					{Type: "message_stop"},
				},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"What is this?","selected_text":"some text","surrounding_text":"context around"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)
		assert.Equal(t, "some text", mock.captured.SelectedText)
		assert.Equal(t, "context around", mock.captured.SurroundingText)
	})

	t.Run("stores selected_text on user message", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &mockStreamer{
			events: []anthropic.StreamEvent{
				{Type: "content_block_delta", Text: "Ok"},
				{Type: "message_stop"},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"What is this?","selected_text":"some text","surrounding_text":"context"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		messages, err := store.ListMessages(tdb.DB, "chat-1")
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(messages), 1)
		require.NotNil(t, messages[0].SelectedText)
		assert.Equal(t, "some text", *messages[0].SelectedText)
		require.NotNil(t, messages[0].SurroundingText)
		assert.Equal(t, "context", *messages[0].SurroundingText)
	})
}

func testMuxWithChat(t *testing.T, tdb *store.TestDB, chat ChatStreamer) *http.ServeMux {
	t.Helper()
	storage := pdf.NewStorage(t.TempDir())
	return NewMux(tdb.DB, storage, chat)
}
