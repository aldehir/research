package anthropic

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStream_LogsStreamStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	c := NewClient("test-key")
	c.BaseURL = server.URL
	c.Logger = logger

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	for range ch {
	}

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "stream starting"), "should log stream start")
}

func TestStream_LogsRequestDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c := NewClient("test-key")
	c.BaseURL = server.URL
	c.Logger = logger

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
			{Role: "user", Content: "Tell me more"},
		},
		SystemPrompt: "You are a helpful assistant.",
		Tools:        PDFTools(),
	})
	require.NoError(t, err)
	for range ch {
	}

	logOutput := buf.String()
	assert.Contains(t, logOutput, "system_prompt_length")
	assert.Contains(t, logOutput, "tool_count")
	assert.Contains(t, logOutput, "message_count")
}

func TestStream_LogsSSEEventTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"hi\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c := NewClient("test-key")
	c.BaseURL = server.URL
	c.Logger = logger

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	for range ch {
	}

	logOutput := buf.String()
	assert.Contains(t, logOutput, "content_block_start")
	assert.Contains(t, logOutput, "content_block_delta")
	assert.Contains(t, logOutput, "content_block_stop")
	assert.Contains(t, logOutput, "message_stop")
}

func TestStream_LogsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	c := NewClient("bad-key")
	c.BaseURL = server.URL
	c.Logger = logger

	_, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.Error(t, err)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "anthropic api error"), "should log API error")
}
