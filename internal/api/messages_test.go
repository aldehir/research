package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aldehir/research/internal/chat"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	gopdf "github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStreamer struct {
	events []chat.StreamEvent
	err    error
}

func (m *mockStreamer) Stream(_ context.Context, _ chat.Request) (<-chan chat.StreamEvent, error) {
	if m.err != nil {
		return nil, m.err
	}
	ch := make(chan chat.StreamEvent)
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
	captured *chat.Request
}

func (c *captureStreamer) Stream(ctx context.Context, req chat.Request) (<-chan chat.StreamEvent, error) {
	c.captured = &req
	return c.mockStreamer.Stream(ctx, req)
}

// multiTurnStreamer simulates a tool call loop: first call returns tool_use,
// subsequent calls return the corresponding events from the sequence.
type multiTurnStreamer struct {
	calls    [][]chat.StreamEvent
	callIdx  int
	requests []chat.Request
}

func (m *multiTurnStreamer) Stream(_ context.Context, req chat.Request) (<-chan chat.StreamEvent, error) {
	m.requests = append(m.requests, req)
	idx := m.callIdx
	if idx >= len(m.calls) {
		idx = len(m.calls) - 1
	}
	m.callIdx++
	events := m.calls[idx]
	ch := make(chan chat.StreamEvent)
	go func() {
		defer close(ch)
		for _, ev := range events {
			ch <- ev
		}
	}()
	return ch, nil
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
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Error     string          `json:"error,omitempty"`
	Name      string          `json:"name,omitempty"`
	Args      json.RawMessage `json:"args,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
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
			events: []chat.StreamEvent{
				{Kind: chat.EventDelta, Text:"Hello"},
				{Kind: chat.EventDelta, Text:" world"},
				{Kind: chat.EventDone},
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
			events: []chat.StreamEvent{
				{Kind: chat.EventDelta, Text:"Reply"},
				{Kind: chat.EventDone},
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
			events: []chat.StreamEvent{
				{Kind: chat.EventDelta, Text:"Hello"},
				{Kind: chat.EventDelta, Text:" world"},
				{Kind: chat.EventDone},
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

	t.Run("stream error returns error SSE event", func(t *testing.T) {
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

	t.Run("appends viewer context to user message", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []chat.StreamEvent{
					{Kind: chat.EventDelta, Text:"Ok"},
					{Kind: chat.EventDone},
				},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"What is this?","current_page":5}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)

		// The last user message should contain viewer context
		lastMsg := mock.captured.Messages[len(mock.captured.Messages)-1]
		require.NotEmpty(t, lastMsg.Parts)
		assert.Contains(t, lastMsg.Parts[0].Text, "What is this?")
		assert.Contains(t, lastMsg.Parts[0].Text, "Current page: 5")
	})

	t.Run("no viewer context appended when no page info", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []chat.StreamEvent{
					{Kind: chat.EventDelta, Text:"Ok"},
					{Kind: chat.EventDone},
				},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)

		lastMsg := mock.captured.Messages[len(mock.captured.Messages)-1]
		require.NotEmpty(t, lastMsg.Parts)
		assert.Equal(t, "Hello", lastMsg.Parts[0].Text)
	})

	t.Run("populates document metadata from paper record", func(t *testing.T) {
		tdb := store.NewTestDB(t)

		// Create paper with metadata
		author := "Einstein"
		pageCount := 20
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Relativity",
			FilePath:  "/papers/test.pdf",
			FileSize:  12345,
			Author:    &author,
			PageCount: &pageCount,
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

		mock := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []chat.StreamEvent{
					{Kind: chat.EventDelta, Text:"Ok"},
					{Kind: chat.EventDone},
				},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"Explain this","current_page":5}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)
		assert.Contains(t, mock.captured.SystemPrompt, "Relativity")
		assert.Contains(t, mock.captured.SystemPrompt, "Einstein")
		assert.Contains(t, mock.captured.SystemPrompt, "20 pages")
	})

}

func TestSendMessage_ToolExecutionLoop(t *testing.T) {
	t.Run("executes search_pdf tool and returns final text", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		// Create a PDF with searchable text
		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "The attention mechanism is key to transformer models.")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Attention Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				// First call: model wants to search
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "search_pdf", Input: json.RawMessage(`{"query":"attention"}`)}},
					{Kind: chat.EventDone},
				},
				// Second call (after tool result): model gives final answer
				{
					{Kind: chat.EventDelta, Text:"I found attention on page 1"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Find where attention is discussed"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse SSE events
		events := parseSSEEvents(t, rec.Body.String())

		// Should have tool_call, delta, and done events
		var hasToolCall, hasDelta, hasDone bool
		for _, ev := range events {
			switch ev.Type {
			case "tool_call":
				hasToolCall = true
			case "delta":
				hasDelta = true
			case "done":
				hasDone = true
			}
		}
		assert.True(t, hasToolCall, "should emit tool_call event")
		assert.True(t, hasDelta, "should emit delta event with final text")
		assert.True(t, hasDone, "should emit done event")

		// Verify the second request included tool_result
		require.Len(t, mock.requests, 2)
		secondReq := mock.requests[1]
		lastMsg := secondReq.Messages[len(secondReq.Messages)-1]
		require.NotEmpty(t, lastMsg.Parts)
		assert.Equal(t, chat.PartToolResult, lastMsg.Parts[0].Kind)
	})

	t.Run("executes read_page tool", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Page one content here.")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_2", Name: "read_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"The page says..."},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"What's on page 1?"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify tool_result was sent back with page text
		require.Len(t, mock.requests, 2)
		secondReq := mock.requests[1]
		lastMsg := secondReq.Messages[len(secondReq.Messages)-1]
		require.NotEmpty(t, lastMsg.Parts)
		assert.Equal(t, chat.PartToolResult, lastMsg.Parts[0].Kind)
		require.NotNil(t, lastMsg.Parts[0].ToolResult)
		assert.Contains(t, lastMsg.Parts[0].ToolResult.Content, "Page one content")
	})

	t.Run("go_to_page emits client-side SSE event", func(t *testing.T) {
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_3", Name: "go_to_page", Input: json.RawMessage(`{"page":5}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Navigated to page 5"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Go to page 5"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		events := parseSSEEvents(t, rec.Body.String())
		var toolCallEvent *sseEvent
		for i := range events {
			if events[i].Type == "tool_call" {
				toolCallEvent = &events[i]
				break
			}
		}
		require.NotNil(t, toolCallEvent)
	})
}

func TestSendMessage_ToolResultSSE(t *testing.T) {
	t.Run("emits tool_result SSE events after tool execution", func(t *testing.T) {
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":3}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Done"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Go to page 3"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		events := parseSSEEvents(t, rec.Body.String())

		// Find tool_result event
		var toolResult *sseEvent
		for i := range events {
			if events[i].Type == "tool_result" {
				toolResult = &events[i]
				break
			}
		}
		require.NotNil(t, toolResult, "should emit tool_result SSE event")
		assert.Equal(t, "go_to_page", toolResult.Name)
		assert.NotEmpty(t, toolResult.Text, "tool_result should include content")
	})

	t.Run("tool_result includes preview for large results", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Page one content here.")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_2", Name: "read_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"The page says..."},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"What's on page 1?"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		events := parseSSEEvents(t, rec.Body.String())

		var toolResult *sseEvent
		for i := range events {
			if events[i].Type == "tool_result" {
				toolResult = &events[i]
				break
			}
		}
		require.NotNil(t, toolResult, "should emit tool_result SSE event")
		assert.Equal(t, "read_page", toolResult.Name)
		assert.NotEmpty(t, toolResult.Text, "tool_result should include content")
	})

	t.Run("emits tool_result for each tool in multi-tool response", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Some content.")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_2", Name: "read_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Here it is"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Read page 1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		events := parseSSEEvents(t, rec.Body.String())

		var toolResults []sseEvent
		for _, ev := range events {
			if ev.Type == "tool_result" {
				toolResults = append(toolResults, ev)
			}
		}
		require.Len(t, toolResults, 2, "should emit tool_result for each tool call")
		assert.Equal(t, "go_to_page", toolResults[0].Name)
		assert.Equal(t, "read_page", toolResults[1].Name)
	})
}

func TestSendMessage_LogsToolLoopIteration(t *testing.T) {
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

	mock := &multiTurnStreamer{
		calls: [][]chat.StreamEvent{
			{
				{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":3}`)}},
				{Kind: chat.EventDone},
			},
			{
				{Kind: chat.EventDelta, Text:"Done"},
				{Kind: chat.EventDone},
			},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mux := NewMux(tdb.DB, storage, mock, nil, logger, WithRetentionTTL(10*time.Millisecond))

	body := `{"content":"Go to page 3"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "tool_loop_iteration")
	assert.Contains(t, logOutput, "tool_call")
	assert.Contains(t, logOutput, "go_to_page")
}

func TestSendMessage_LogsToolExecutionResults(t *testing.T) {
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

	mock := &multiTurnStreamer{
		calls: [][]chat.StreamEvent{
			{
				{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":2}`)}},
				{Kind: chat.EventDone},
			},
			{
				{Kind: chat.EventDelta, Text:"Done"},
				{Kind: chat.EventDone},
			},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mux := NewMux(tdb.DB, storage, mock, nil, logger, WithRetentionTTL(10*time.Millisecond))

	body := `{"content":"Go to page 2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "tool_result")
	assert.Contains(t, logOutput, "result_length")
	assert.Contains(t, logOutput, "duration")
}

func TestSendMessage_LogsFinalResponseSummary(t *testing.T) {
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

	mock := &multiTurnStreamer{
		calls: [][]chat.StreamEvent{
			{
				{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":1}`)}},
				{Kind: chat.EventDone},
			},
			{
				{Kind: chat.EventDelta, Text:"Here is page 1"},
				{Kind: chat.EventDone},
			},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mux := NewMux(tdb.DB, storage, mock, nil, logger, WithRetentionTTL(10*time.Millisecond))

	body := `{"content":"Show page 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "response_complete")
	assert.Contains(t, logOutput, "response_length")
	assert.Contains(t, logOutput, "tool_iterations")
	assert.Contains(t, logOutput, "total_duration")
}

func TestSendMessage_SnapshotPage(t *testing.T) {
	t.Run("executes snapshot_page tool and returns image content", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		// Create a real PDF so pdftoppm can render it
		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Chart data here")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_snap", Name: "snapshot_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"I can see the chart"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Show me the chart"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		events := parseSSEEvents(t, rec.Body.String())

		// Should have tool_call and tool_result events
		var hasToolCall, hasToolResult, hasDone bool
		for _, ev := range events {
			switch ev.Type {
			case "tool_call":
				hasToolCall = true
				assert.Equal(t, "snapshot_page", ev.Name)
			case "tool_result":
				hasToolResult = true
				assert.Equal(t, "snapshot_page", ev.Name)
			case "done":
				hasDone = true
			}
		}
		assert.True(t, hasToolCall, "should emit tool_call event")
		assert.True(t, hasToolResult, "should emit tool_result event")
		assert.True(t, hasDone, "should emit done event")

		// Verify the tool result sent to provider has image content
		require.Len(t, mock.requests, 2)
		secondReq := mock.requests[1]
		lastMsg := secondReq.Messages[len(secondReq.Messages)-1]
		require.NotEmpty(t, lastMsg.Parts)
		toolResult := lastMsg.Parts[0]
		assert.Equal(t, chat.PartToolResult, toolResult.Kind)
		require.NotNil(t, toolResult.ToolResult)
		require.NotNil(t, toolResult.ToolResult.Image, "snapshot_page should return image content")
		assert.Equal(t, "image/png", toolResult.ToolResult.Image.MediaType)
		assert.NotEmpty(t, toolResult.ToolResult.Image.Data)
	})

	t.Run("snapshot_page SSE result includes content_type image", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Visual content")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_snap2", Name: "snapshot_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Done"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Show me page 1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse raw SSE to check content_type field
		events := parseSSEEvents(t, rec.Body.String())
		var toolResult *sseEvent
		for i := range events {
			if events[i].Type == "tool_result" {
				toolResult = &events[i]
				break
			}
		}
		require.NotNil(t, toolResult)
		assert.Equal(t, "snapshot_page", toolResult.Name)
	})
}

func TestSendMessage_ReadPageUsesIndex(t *testing.T) {
	tdb := store.NewTestDB(t)
	storage := pdf.NewStorage(t.TempDir())

	// Create paper WITHOUT a real PDF file — read_page must use the index
	p := store.Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  "/nonexistent.pdf",
		FileSize:  1000,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, store.CreatePaper(tdb.DB, p))

	// Pre-index page text
	require.NoError(t, store.UpsertPageText(tdb.DB, "paper-1", 1, "Indexed page content"))

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	mock := &multiTurnStreamer{
		calls: [][]chat.StreamEvent{
			{
				{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "read_page", Input: json.RawMessage(`{"page":1}`)}},
				{Kind: chat.EventDone},
			},
			{
				{Kind: chat.EventDelta, Text:"Got it"},
				{Kind: chat.EventDone},
			},
		},
	}

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

	body := `{"content":"Read page 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify tool result contains indexed content (not pdftotext)
	require.Len(t, mock.requests, 2)
	lastMsg := mock.requests[1].Messages[len(mock.requests[1].Messages)-1]
	require.NotEmpty(t, lastMsg.Parts)
	assert.Equal(t, chat.PartToolResult, lastMsg.Parts[0].Kind)
	require.NotNil(t, lastMsg.Parts[0].ToolResult)
	assert.Contains(t, lastMsg.Parts[0].ToolResult.Content, "Indexed page content")
}

func TestSendMessage_WithAttachments(t *testing.T) {
	t.Run("includes image and text parts in chat message", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []chat.StreamEvent{
					{Kind: chat.EventDelta, Text:"I see the figure"},
					{Kind: chat.EventDone},
				},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"What does this show?","current_page":3,"attachments":[{"image_data":"aWdub3Jl","text":"Figure 1: Results","page":3}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)

		// The user message should have parts (multimodal)
		lastMsg := mock.captured.Messages[len(mock.captured.Messages)-1]
		require.NotEmpty(t, lastMsg.Parts, "should use parts for multimodal message")

		// Should have a text part and an image part
		var hasText, hasImage bool
		for _, part := range lastMsg.Parts {
			if part.Kind == chat.PartText {
				hasText = true
				assert.Contains(t, part.Text, "What does this show?")
				assert.Contains(t, part.Text, "Figure 1: Results")
			}
			if part.Kind == chat.PartImage {
				hasImage = true
				require.NotNil(t, part.Image)
				assert.Equal(t, "image/png", part.Image.MediaType)
				assert.Equal(t, "aWdub3Jl", part.Image.Data)
			}
		}
		assert.True(t, hasText, "should have text part")
		assert.True(t, hasImage, "should have image part")
	})

	t.Run("message without attachments works as before", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		seedChatSession(t, tdb)

		mock := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []chat.StreamEvent{
					{Kind: chat.EventDelta, Text:"Ok"},
					{Kind: chat.EventDone},
				},
			},
		}
		mux := testMuxWithChat(t, tdb, mock)

		body := `{"content":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)

		// Should use a single text part
		lastMsg := mock.captured.Messages[len(mock.captured.Messages)-1]
		require.Len(t, lastMsg.Parts, 1)
		assert.Equal(t, chat.PartText, lastMsg.Parts[0].Kind)
		assert.Equal(t, "Hello", lastMsg.Parts[0].Text)
	})
}

func TestSendMessage_PersistsToolInteractions(t *testing.T) {
	t.Run("tool_use and tool_result messages are persisted to DB", func(t *testing.T) {
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

		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":3}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Done navigating"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Go to page 3"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Check persisted messages
		messages, err := store.ListMessages(tdb.DB, "chat-1")
		require.NoError(t, err)

		// Should have: user msg, assistant tool_use, user tool_result, assistant final text
		require.GreaterOrEqual(t, len(messages), 4, "should persist tool interaction messages")

		// First: user message
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "Go to page 3", messages[0].Content)

		// Second: assistant with tool_call content blocks
		assert.Equal(t, "assistant", messages[1].Role)
		require.NotNil(t, messages[1].ContentBlocks, "assistant tool_call message should have content_blocks")
		assert.Contains(t, *messages[1].ContentBlocks, "tool_call")
		assert.Contains(t, *messages[1].ContentBlocks, "toolu_1")

		// Third: user with tool_result content blocks
		assert.Equal(t, "user", messages[2].Role)
		require.NotNil(t, messages[2].ContentBlocks, "user tool_result message should have content_blocks")
		assert.Contains(t, *messages[2].ContentBlocks, "tool_result")

		// Fourth: final assistant text
		assert.Equal(t, "assistant", messages[3].Role)
		assert.Equal(t, "Done navigating", messages[3].Content)
	})

	t.Run("persisted tool messages are reconstructed for next turn", func(t *testing.T) {
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

		// --- Turn 1: tool call ---
		mock1 := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":3}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Navigated"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux1 := NewMux(tdb.DB, storage, mock1, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body1 := `{"content":"Go to page 3"}`
		req1 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		rec1 := httptest.NewRecorder()
		mux1.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// --- Turn 2: follow-up question ---
		mock2 := &captureStreamer{
			mockStreamer: mockStreamer{
				events: []chat.StreamEvent{
					{Kind: chat.EventDelta, Text:"Yes"},
					{Kind: chat.EventDone},
				},
			},
		}
		mux2 := NewMux(tdb.DB, storage, mock2, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body2 := `{"content":"What did you find?"}`
		req2 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		rec2 := httptest.NewRecorder()
		mux2.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)

		// Verify the second request's messages include tool interactions
		require.NotNil(t, mock2.captured)
		msgs := mock2.captured.Messages
		// Should have: user, assistant(tool_use), user(tool_result), assistant(text), user
		require.GreaterOrEqual(t, len(msgs), 5, "history should include tool interaction messages")

		// Find the assistant message with tool_call parts
		var hasToolCallMsg, hasToolResultMsg bool
		for _, m := range msgs {
			for _, p := range m.Parts {
				if p.Kind == chat.PartToolCall {
					hasToolCallMsg = true
				}
				if p.Kind == chat.PartToolResult {
					hasToolResultMsg = true
				}
			}
		}
		assert.True(t, hasToolCallMsg, "history should contain tool_call message")
		assert.True(t, hasToolResultMsg, "history should contain tool_result message")
	})

	t.Run("no duplicate final message when tools were used", func(t *testing.T) {
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

		// Tool loop with final text
		mock := &multiTurnStreamer{
			calls: [][]chat.StreamEvent{
				{
					{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "go_to_page", Input: json.RawMessage(`{"page":1}`)}},
					{Kind: chat.EventDone},
				},
				{
					{Kind: chat.EventDelta, Text:"Here is page 1"},
					{Kind: chat.EventDone},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

		body := `{"content":"Show page 1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		messages, err := store.ListMessages(tdb.DB, "chat-1")
		require.NoError(t, err)

		// Count assistant messages with "Here is page 1" — should be exactly 1
		var count int
		for _, m := range messages {
			if m.Role == "assistant" && m.Content == "Here is page 1" {
				count++
			}
		}
		assert.Equal(t, 1, count, "final assistant text should appear exactly once, not duplicated")
	})
}

func TestSendMessage_PersistsSnapshotPageImage(t *testing.T) {
	tdb := store.NewTestDB(t)
	storage := pdf.NewStorage(t.TempDir())

	pdfPath := storage.Path("paper-1")
	createTestPDFWithText(t, pdfPath, "Chart data here")

	pageCount := 1
	p := store.Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  pdfPath,
		FileSize:  1000,
		PageCount: &pageCount,
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

	mock := &multiTurnStreamer{
		calls: [][]chat.StreamEvent{
			{
				{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_snap", Name: "snapshot_page", Input: json.RawMessage(`{"page":1}`)}},
				{Kind: chat.EventDone},
			},
			{
				{Kind: chat.EventDelta, Text:"I see the chart"},
				{Kind: chat.EventDone},
			},
		},
	}

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

	body := `{"content":"Show chart"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	messages, err := store.ListMessages(tdb.DB, "chat-1")
	require.NoError(t, err)

	// Find the tool_result message
	var toolResultMsg *store.Message
	for i := range messages {
		if messages[i].Role == "user" && messages[i].ContentBlocks != nil {
			if strings.Contains(*messages[i].ContentBlocks, "tool_result") {
				toolResultMsg = &messages[i]
				break
			}
		}
	}
	require.NotNil(t, toolResultMsg, "should persist tool_result message")

	// Deserialize and verify it contains image content
	var parts []chat.Part
	require.NoError(t, json.Unmarshal([]byte(*toolResultMsg.ContentBlocks), &parts))
	require.Len(t, parts, 1)
	assert.Equal(t, chat.PartToolResult, parts[0].Kind)
	require.NotNil(t, parts[0].ToolResult, "snapshot_page should have a tool result")
	require.NotNil(t, parts[0].ToolResult.Image, "snapshot_page tool_result should contain image content")
	assert.Equal(t, "image/png", parts[0].ToolResult.Image.MediaType)
	assert.NotEmpty(t, parts[0].ToolResult.Image.Data)
}

func TestSendMessage_ReloadsPersistedAttachmentsInHistory(t *testing.T) {
	tdb := store.NewTestDB(t)
	dataDir := t.TempDir()
	storage := pdf.NewStorage(filepath.Join(dataDir, "papers"))

	seedTestPaper(t, tdb)

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	// --- Turn 1: send message with attachment ---
	mock1 := &mockStreamer{
		events: []chat.StreamEvent{
			{Kind: chat.EventDelta, Text:"I see a chart"},
			{Kind: chat.EventDone},
		},
	}
	mux1 := NewMux(tdb.DB, storage, mock1, nil, slog.Default(), WithDataDir(dataDir), WithRetentionTTL(10*time.Millisecond))

	body1 := `{"content":"What is this?","attachments":[{"image_data":"iVBORw0KGgo=","text":"Figure 1","page":3}]}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	mux1.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// --- Turn 2: follow-up question (simulates page reload — no in-memory state) ---
	mock2 := &captureStreamer{
		mockStreamer: mockStreamer{
			events: []chat.StreamEvent{
				{Kind: chat.EventDelta, Text:"Yes, the chart shows..."},
				{Kind: chat.EventDone},
			},
		},
	}
	mux2 := NewMux(tdb.DB, storage, mock2, nil, slog.Default(), WithDataDir(dataDir), WithRetentionTTL(10*time.Millisecond))

	body2 := `{"content":"Can you still see the image?"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux2.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Verify the second request included the image from the first turn
	require.NotNil(t, mock2.captured)
	msgs := mock2.captured.Messages

	// The first user message should have parts with an image
	firstMsg := msgs[0]
	assert.Equal(t, chat.RoleUser, firstMsg.Role)
	require.NotEmpty(t, firstMsg.Parts, "first user message should have parts with image from persisted attachment")

	var hasImage bool
	for _, p := range firstMsg.Parts {
		if p.Kind == chat.PartImage && p.Image != nil {
			hasImage = true
			assert.Equal(t, "image/png", p.Image.MediaType)
			assert.NotEmpty(t, p.Image.Data)
		}
	}
	assert.True(t, hasImage, "first user message should include the persisted attachment image")
}

func TestSendMessage_PersistsAttachments(t *testing.T) {
	tdb := store.NewTestDB(t)
	dataDir := t.TempDir()
	storage := pdf.NewStorage(filepath.Join(dataDir, "papers"))

	seedTestPaper(t, tdb)

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	mock := &mockStreamer{
		events: []chat.StreamEvent{
			{Kind: chat.EventDelta, Text:"I see the figure"},
			{Kind: chat.EventDone},
		},
	}
	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithDataDir(dataDir), WithRetentionTTL(10*time.Millisecond))

	// Send message with attachment (base64 PNG)
	body := `{"content":"What is this?","attachments":[{"image_data":"iVBORw0KGgo=","text":"Figure 1","page":3}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Check attachments were persisted
	atts, err := store.ListAttachmentsByChat(tdb.DB, "chat-1")
	require.NoError(t, err)
	require.Len(t, atts, 1)
	assert.Equal(t, "Figure 1", atts[0].Text)
	assert.Equal(t, 3, atts[0].Page)

	// Check image file exists on disk
	assert.FileExists(t, atts[0].FilePath)
}

func TestGetAttachmentImage(t *testing.T) {
	tdb := store.NewTestDB(t)
	dataDir := t.TempDir()
	storage := pdf.NewStorage(filepath.Join(dataDir, "papers"))

	seedTestPaper(t, tdb)

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	// Create a message and attachment manually
	msg := store.Message{ID: "msg-1", ChatSessionID: "chat-1", Role: "user", Content: "Hello", CreatedAt: "2026-03-28T10:01:00Z"}
	require.NoError(t, store.CreateMessage(tdb.DB, msg))

	// Write a fake PNG to disk
	attDir := filepath.Join(dataDir, "attachments")
	require.NoError(t, os.MkdirAll(attDir, 0o755))
	imgPath := filepath.Join(attDir, "att-1.png")
	pngData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	require.NoError(t, os.WriteFile(imgPath, pngData, 0o644))

	att := store.Attachment{
		ID:        "att-1",
		MessageID: "msg-1",
		FilePath:  imgPath,
		Text:      "Fig 1",
		Page:      1,
		CreatedAt: "2026-03-28T10:01:00Z",
	}
	require.NoError(t, store.CreateAttachment(tdb.DB, att))

	mux := NewMux(tdb.DB, storage, nil, nil, slog.Default(), WithDataDir(dataDir), WithRetentionTTL(10*time.Millisecond))

	req := httptest.NewRequest(http.MethodGet, "/api/attachments/att-1/image", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	assert.Equal(t, pngData, rec.Body.Bytes())
}

func TestGetChatSessionIncludesAttachments(t *testing.T) {
	tdb := store.NewTestDB(t)
	dataDir := t.TempDir()
	storage := pdf.NewStorage(filepath.Join(dataDir, "papers"))

	seedTestPaper(t, tdb)

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	msg := store.Message{ID: "msg-1", ChatSessionID: "chat-1", Role: "user", Content: "Hello", CreatedAt: "2026-03-28T10:01:00Z"}
	require.NoError(t, store.CreateMessage(tdb.DB, msg))

	att := store.Attachment{
		ID:        "att-1",
		MessageID: "msg-1",
		FilePath:  "/data/att-1.png",
		Text:      "Figure 1",
		Page:      3,
		CreatedAt: "2026-03-28T10:01:00Z",
	}
	require.NoError(t, store.CreateAttachment(tdb.DB, att))

	mux := NewMux(tdb.DB, storage, nil, nil, slog.Default(), WithDataDir(dataDir), WithRetentionTTL(10*time.Millisecond))

	req := httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats/chat-1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result struct {
		Messages []struct {
			ID          string `json:"id"`
			Attachments []struct {
				ID   string `json:"id"`
				Text string `json:"text"`
				Page int    `json:"page"`
			} `json:"attachments"`
		} `json:"messages"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	require.Len(t, result.Messages, 1)
	require.Len(t, result.Messages[0].Attachments, 1)
	assert.Equal(t, "att-1", result.Messages[0].Attachments[0].ID)
	assert.Equal(t, "Figure 1", result.Messages[0].Attachments[0].Text)
	assert.Equal(t, 3, result.Messages[0].Attachments[0].Page)
}

func TestSendMessage_SearchUsesIndex(t *testing.T) {
	tdb := store.NewTestDB(t)
	storage := pdf.NewStorage(t.TempDir())

	// Create paper WITHOUT a real PDF file — search must use the index
	p := store.Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  "/nonexistent.pdf",
		FileSize:  1000,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, store.CreatePaper(tdb.DB, p))

	// Pre-index text
	require.NoError(t, store.UpsertPageText(tdb.DB, "paper-1", 1, "Introduction to neural networks"))
	require.NoError(t, store.UpsertPageText(tdb.DB, "paper-1", 2, "Training deep neural models"))

	session := store.ChatSession{
		ID:        "chat-1",
		PaperID:   "paper-1",
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, store.CreateChatSession(tdb.DB, session))

	mock := &multiTurnStreamer{
		calls: [][]chat.StreamEvent{
			{
				{Kind: chat.EventToolCall, ToolCall: &chat.ToolCall{ID: "toolu_1", Name: "search_pdf", Input: json.RawMessage(`{"query":"neural"}`)}},
				{Kind: chat.EventDone},
			},
			{
				{Kind: chat.EventDelta, Text:"Found neural on pages 1 and 2"},
				{Kind: chat.EventDone},
			},
		},
	}

	mux := NewMux(tdb.DB, storage, mock, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))

	body := `{"content":"Search for neural"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify tool result contains FTS results
	require.Len(t, mock.requests, 2)
	lastMsg := mock.requests[1].Messages[len(mock.requests[1].Messages)-1]
	require.NotEmpty(t, lastMsg.Parts)
	assert.Equal(t, chat.PartToolResult, lastMsg.Parts[0].Kind)
	require.NotNil(t, lastMsg.Parts[0].ToolResult)
	assert.Contains(t, lastMsg.Parts[0].ToolResult.Content, "neural")
}

// createTestPDFWithText is a helper that creates a valid PDF with text.
func createTestPDFWithText(t *testing.T, path string, text string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	doc := gopdf.New("P", "mm", "Letter", "")
	doc.AddPage()
	doc.SetFont("Helvetica", "", 12)
	doc.Text(10, 20, text)
	require.NoError(t, doc.OutputFileAndClose(path))
}

func testMuxWithChat(t *testing.T, tdb *store.TestDB, provider chat.Provider) *http.ServeMux {
	t.Helper()
	return NewMux(tdb.DB, pdf.NewStorage(t.TempDir()), provider, nil, slog.Default(), WithRetentionTTL(10*time.Millisecond))
}
