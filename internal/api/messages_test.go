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

	"github.com/aldehir/research/internal/anthropic"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	gopdf "github.com/go-pdf/fpdf"
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

// multiTurnStreamer simulates a tool call loop: first call returns tool_use,
// subsequent calls return the corresponding events from the sequence.
type multiTurnStreamer struct {
	calls    [][]anthropic.StreamEvent
	callIdx  int
	requests []anthropic.Request
}

func (m *multiTurnStreamer) Stream(_ context.Context, req anthropic.Request) (<-chan anthropic.StreamEvent, error) {
	m.requests = append(m.requests, req)
	idx := m.callIdx
	if idx >= len(m.calls) {
		idx = len(m.calls) - 1
	}
	m.callIdx++
	events := m.calls[idx]
	ch := make(chan anthropic.StreamEvent)
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

	t.Run("appends viewer context to user message", func(t *testing.T) {
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

		body := `{"content":"What is this?","current_page":5,"selected_text":"some formula"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)

		// The last user message should contain viewer context
		lastMsg := mock.captured.Messages[len(mock.captured.Messages)-1]
		assert.Contains(t, lastMsg.Content, "What is this?")
		assert.Contains(t, lastMsg.Content, "Current page: 5")
		assert.Contains(t, lastMsg.Content, "Selected text: some formula")
		// Should NOT contain surrounding text
		assert.NotContains(t, lastMsg.Content, "surrounding")
	})

	t.Run("no viewer context appended when no page info", func(t *testing.T) {
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

		body := `{"content":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, mock.captured)

		lastMsg := mock.captured.Messages[len(mock.captured.Messages)-1]
		assert.Equal(t, "Hello", lastMsg.Content)
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
				events: []anthropic.StreamEvent{
					{Type: "content_block_delta", Text: "Ok"},
					{Type: "message_stop"},
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
		assert.Equal(t, "Relativity", mock.captured.DocumentTitle)
		assert.Equal(t, "Einstein", mock.captured.DocumentAuthor)
		assert.Equal(t, 20, mock.captured.TotalPages)
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
			calls: [][]anthropic.StreamEvent{
				// First call: model wants to search
				{
					{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "search_pdf", ToolInput: `{"query":"attention"}`},
					{Type: "message_stop"},
				},
				// Second call (after tool result): model gives final answer
				{
					{Type: "content_block_delta", Text: "I found attention on page 1"},
					{Type: "message_stop"},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, slog.Default())

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
		require.NotEmpty(t, lastMsg.ContentBlocks)
		assert.Equal(t, "tool_result", lastMsg.ContentBlocks[0].Type)
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
			calls: [][]anthropic.StreamEvent{
				{
					{Type: "tool_use", ToolUseID: "toolu_2", ToolName: "read_page", ToolInput: `{"page":1}`},
					{Type: "message_stop"},
				},
				{
					{Type: "content_block_delta", Text: "The page says..."},
					{Type: "message_stop"},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, slog.Default())

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
		require.NotEmpty(t, lastMsg.ContentBlocks)
		assert.Equal(t, "tool_result", lastMsg.ContentBlocks[0].Type)
		assert.Contains(t, lastMsg.ContentBlocks[0].Content, "Page one content")
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
			calls: [][]anthropic.StreamEvent{
				{
					{Type: "tool_use", ToolUseID: "toolu_3", ToolName: "go_to_page", ToolInput: `{"page":5}`},
					{Type: "message_stop"},
				},
				{
					{Type: "content_block_delta", Text: "Navigated to page 5"},
					{Type: "message_stop"},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, slog.Default())

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
			calls: [][]anthropic.StreamEvent{
				{
					{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "go_to_page", ToolInput: `{"page":3}`},
					{Type: "message_stop"},
				},
				{
					{Type: "content_block_delta", Text: "Done"},
					{Type: "message_stop"},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, slog.Default())

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
			calls: [][]anthropic.StreamEvent{
				{
					{Type: "tool_use", ToolUseID: "toolu_2", ToolName: "read_page", ToolInput: `{"page":1}`},
					{Type: "message_stop"},
				},
				{
					{Type: "content_block_delta", Text: "The page says..."},
					{Type: "message_stop"},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, slog.Default())

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
			calls: [][]anthropic.StreamEvent{
				{
					{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "go_to_page", ToolInput: `{"page":1}`},
					{Type: "tool_use", ToolUseID: "toolu_2", ToolName: "read_page", ToolInput: `{"page":1}`},
					{Type: "message_stop"},
				},
				{
					{Type: "content_block_delta", Text: "Here it is"},
					{Type: "message_stop"},
				},
			},
		}

		mux := NewMux(tdb.DB, storage, mock, slog.Default())

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
		calls: [][]anthropic.StreamEvent{
			{
				{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "go_to_page", ToolInput: `{"page":3}`},
				{Type: "message_stop"},
			},
			{
				{Type: "content_block_delta", Text: "Done"},
				{Type: "message_stop"},
			},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mux := NewMux(tdb.DB, storage, mock, logger)

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
		calls: [][]anthropic.StreamEvent{
			{
				{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "go_to_page", ToolInput: `{"page":2}`},
				{Type: "message_stop"},
			},
			{
				{Type: "content_block_delta", Text: "Done"},
				{Type: "message_stop"},
			},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mux := NewMux(tdb.DB, storage, mock, logger)

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
		calls: [][]anthropic.StreamEvent{
			{
				{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "go_to_page", ToolInput: `{"page":1}`},
				{Type: "message_stop"},
			},
			{
				{Type: "content_block_delta", Text: "Here is page 1"},
				{Type: "message_stop"},
			},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mux := NewMux(tdb.DB, storage, mock, logger)

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
		calls: [][]anthropic.StreamEvent{
			{
				{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "read_page", ToolInput: `{"page":1}`},
				{Type: "message_stop"},
			},
			{
				{Type: "content_block_delta", Text: "Got it"},
				{Type: "message_stop"},
			},
		},
	}

	mux := NewMux(tdb.DB, storage, mock, slog.Default())

	body := `{"content":"Read page 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify tool result contains indexed content (not pdftotext)
	require.Len(t, mock.requests, 2)
	lastMsg := mock.requests[1].Messages[len(mock.requests[1].Messages)-1]
	require.NotEmpty(t, lastMsg.ContentBlocks)
	assert.Contains(t, lastMsg.ContentBlocks[0].Content, "Indexed page content")
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
		calls: [][]anthropic.StreamEvent{
			{
				{Type: "tool_use", ToolUseID: "toolu_1", ToolName: "search_pdf", ToolInput: `{"query":"neural"}`},
				{Type: "message_stop"},
			},
			{
				{Type: "content_block_delta", Text: "Found neural on pages 1 and 2"},
				{Type: "message_stop"},
			},
		},
	}

	mux := NewMux(tdb.DB, storage, mock, slog.Default())

	body := `{"content":"Search for neural"}`
	req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats/chat-1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify tool result contains FTS results
	require.Len(t, mock.requests, 2)
	lastMsg := mock.requests[1].Messages[len(mock.requests[1].Messages)-1]
	require.NotEmpty(t, lastMsg.ContentBlocks)
	assert.Contains(t, lastMsg.ContentBlocks[0].Content, "neural")
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

func testMuxWithChat(t *testing.T, tdb *store.TestDB, chat ChatStreamer) *http.ServeMux {
	t.Helper()
	return NewMux(tdb.DB, pdf.NewStorage(t.TempDir()), chat, slog.Default())
}
