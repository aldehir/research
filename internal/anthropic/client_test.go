package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func TestNewClient_WithModel(t *testing.T) {
	c := NewClient("test-key", WithModel("claude-haiku-4-5-20251001"))
	assert.Equal(t, "claude-haiku-4-5-20251001", c.Model)
}

func TestNewClient_WithModel_Empty_UsesDefault(t *testing.T) {
	c := NewClient("test-key", WithModel(""))
	assert.Equal(t, "claude-sonnet-4-20250514", c.Model)
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

func TestStream_APIError_IncludesResponseBody(t *testing.T) {
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
	assert.Contains(t, err.Error(), "invalid api key")
}

func TestStream_SendsToolDefinitions(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"hi\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "search for attention"}},
		Tools:    PDFTools(),
	})
	require.NoError(t, err)
	for range ch {
	}

	// Verify tools were sent in the request
	tools, ok := receivedBody["tools"].([]any)
	require.True(t, ok, "tools field should be present in API request")
	assert.GreaterOrEqual(t, len(tools), 3, "should have at least 3 tool definitions")

	// Verify first tool has expected structure
	tool := tools[0].(map[string]any)
	assert.NotEmpty(t, tool["name"])
	assert.NotEmpty(t, tool["description"])
	assert.NotNil(t, tool["input_schema"])
}

func TestStream_ParsesToolUseEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)

		// content_block_start with tool_use
		fmt.Fprint(w, `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_abc123","name":"search_pdf","input":{}}}

`)
		flusher.Flush()

		// input_json_delta
		fmt.Fprint(w, `event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"query\":"}}

`)
		flusher.Flush()

		fmt.Fprint(w, `event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"attention\"}"}}

`)
		flusher.Flush()

		// content_block_stop
		fmt.Fprint(w, `event: content_block_stop
data: {"type":"content_block_stop","index":0}

`)
		flusher.Flush()

		// message_stop
		fmt.Fprint(w, `event: message_stop
data: {"type":"message_stop"}

`)
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "search for attention"}},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}

	// Should have a tool_use event
	var toolEvent *StreamEvent
	for i := range events {
		if events[i].Type == "tool_use" {
			toolEvent = &events[i]
			break
		}
	}
	require.NotNil(t, toolEvent, "should emit a tool_use event")
	assert.Equal(t, "toolu_abc123", toolEvent.ToolUseID)
	assert.Equal(t, "search_pdf", toolEvent.ToolName)
	assert.JSONEq(t, `{"query":"attention"}`, toolEvent.ToolInput)
}

func TestStream_ToolResultMessages(t *testing.T) {
	// Verify that messages with tool_result content blocks are sent correctly
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"found it\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{
			{Role: "user", Content: "search for attention"},
			{Role: "assistant", ContentBlocks: []ContentBlock{
				{Type: "tool_use", ID: "toolu_abc", Name: "search_pdf", Input: json.RawMessage(`{"query":"attention"}`)},
			}},
			{Role: "user", ContentBlocks: []ContentBlock{
				{Type: "tool_result", ToolUseID: "toolu_abc", Content: "Found on page 3: attention mechanism"},
			}},
		},
	})
	require.NoError(t, err)
	for range ch {
	}

	// Verify the messages were serialized with content blocks
	msgs := receivedBody["messages"].([]any)
	require.Len(t, msgs, 3)

	// Assistant message should have tool_use content block
	assistantMsg := msgs[1].(map[string]any)
	assert.Equal(t, "assistant", assistantMsg["role"])
	blocks := assistantMsg["content"].([]any)
	toolUseBlock := blocks[0].(map[string]any)
	assert.Equal(t, "tool_use", toolUseBlock["type"])
	assert.Equal(t, "toolu_abc", toolUseBlock["id"])

	// User message should have tool_result content block
	userMsg := msgs[2].(map[string]any)
	assert.Equal(t, "user", userMsg["role"])
	resultBlocks := userMsg["content"].([]any)
	resultBlock := resultBlocks[0].(map[string]any)
	assert.Equal(t, "tool_result", resultBlock["type"])
	assert.Equal(t, "toolu_abc", resultBlock["tool_use_id"])
}

func TestContentBlock_ImageToolResultMarshaling(t *testing.T) {
	// Verify that tool_result blocks with image content serialize as
	// structured content arrays, not plain strings.
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"ok\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{
			{Role: "user", Content: "show me the chart"},
			{Role: "assistant", ContentBlocks: []ContentBlock{
				{Type: "tool_use", ID: "toolu_img", Name: "snapshot_page", Input: json.RawMessage(`{"page":1}`)},
			}},
			{Role: "user", ContentBlocks: []ContentBlock{
				{Type: "tool_result", ToolUseID: "toolu_img", ContentParts: []ContentPart{
					{Type: "image", Source: &ImageSource{
						Type:      "base64",
						MediaType: "image/png",
						Data:      "iVBORw0KGgo=",
					}},
				}},
			}},
		},
	})
	require.NoError(t, err)
	for range ch {
	}

	msgs := receivedBody["messages"].([]any)
	require.Len(t, msgs, 3)

	// Tool result message should have structured content array
	userMsg := msgs[2].(map[string]any)
	resultBlocks := userMsg["content"].([]any)
	resultBlock := resultBlocks[0].(map[string]any)
	assert.Equal(t, "tool_result", resultBlock["type"])
	assert.Equal(t, "toolu_img", resultBlock["tool_use_id"])

	// The content field should be an array with an image block
	contentArr, ok := resultBlock["content"].([]any)
	require.True(t, ok, "content should be an array for image tool results")
	require.Len(t, contentArr, 1)

	imgBlock := contentArr[0].(map[string]any)
	assert.Equal(t, "image", imgBlock["type"])
	source := imgBlock["source"].(map[string]any)
	assert.Equal(t, "base64", source["type"])
	assert.Equal(t, "image/png", source["media_type"])
	assert.Equal(t, "iVBORw0KGgo=", source["data"])
}

func TestContentBlock_TextToolResultMarshaling(t *testing.T) {
	// Existing text tool_result blocks should continue to serialize as strings
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"ok\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.BaseURL = server.URL

	ch, err := c.Stream(context.Background(), Request{
		Messages: []Message{
			{Role: "user", Content: "read page 1"},
			{Role: "assistant", ContentBlocks: []ContentBlock{
				{Type: "tool_use", ID: "toolu_txt", Name: "read_page", Input: json.RawMessage(`{"page":1}`)},
			}},
			{Role: "user", ContentBlocks: []ContentBlock{
				{Type: "tool_result", ToolUseID: "toolu_txt", Content: "Page text here"},
			}},
		},
	})
	require.NoError(t, err)
	for range ch {
	}

	msgs := receivedBody["messages"].([]any)
	require.Len(t, msgs, 3)

	userMsg := msgs[2].(map[string]any)
	resultBlocks := userMsg["content"].([]any)
	resultBlock := resultBlocks[0].(map[string]any)
	assert.Equal(t, "tool_result", resultBlock["type"])
	// The content field should be a string for text results
	content, ok := resultBlock["content"].(string)
	require.True(t, ok, "content should be a string for text tool results")
	assert.Equal(t, "Page text here", content)
}

func TestPDFTools_HasExpectedTools(t *testing.T) {
	tools := PDFTools()
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	assert.Contains(t, names, "search_pdf")
	assert.Contains(t, names, "read_page")
	assert.Contains(t, names, "go_to_page")
	assert.Contains(t, names, "snapshot_page")
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
