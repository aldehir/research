package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("test-key")
	assert.Equal(t, "test-key", c.APIKey)
	assert.Equal(t, "https://api.anthropic.com", c.BaseURL)
	assert.Equal(t, "claude-sonnet-4-20250514", c.Model)
	assert.NotNil(t, c.HTTPClient)
}

func TestStream_ReceivesTextDeltas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)

		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\"}\n\n")
		flusher.Flush()

		fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0}\n\n")
		flusher.Flush()

		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n")
		flusher.Flush()

		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\n")
		flusher.Flush()

		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)

	var texts []string
	for ev := range ch {
		if ev.Text != "" {
			texts = append(texts, ev.Text)
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, texts)
}

func TestStream_ChannelClosesAfterMessageStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)

		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"done\"}}\n\n")
		flusher.Flush()

		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)

	// Drain channel — it should close
	var count int
	for range ch {
		count++
	}
	assert.Greater(t, count, 0)
}

func TestStream_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"type":"authentication_error","message":"invalid api key"}}`)
	}))
	defer server.Close()

	c := NewClient("bad-key")
	c.BaseURL = server.URL

	_, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestStream_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)

		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"start\"}}\n\n")
		flusher.Flush()

		// Block until client disconnects
		<-r.Context().Done()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())

	ch, err := c.Stream(ctx, Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)

	// Read the first event
	ev := <-ch
	assert.Equal(t, "start", ev.Text)

	// Cancel and verify channel closes
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			// Might get one more event, but channel should close soon
			_, ok = <-ch
			assert.False(t, ok, "channel should close after context cancellation")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close after context cancellation")
	}
}
