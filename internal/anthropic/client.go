package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Model      string
	Logger     *slog.Logger
}

// ImageSource holds the base64-encoded data for an image content part.
type ImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // e.g. "image/png"
	Data      string `json:"data"`       // base64-encoded image data
}

// ContentPart is a structured part within a tool_result content array.
type ContentPart struct {
	Type   string       `json:"type"`             // "text" or "image"
	Text   string       `json:"text,omitempty"`   // for type "text"
	Source *ImageSource `json:"source,omitempty"` // for type "image"
}

// ContentBlock represents a structured content block in a message.
// For tool_result blocks, use Content for plain text results or
// ContentParts for structured results (e.g. images).
// For image blocks in user messages, use Source.
type ContentBlock struct {
	Type         string          `json:"-"`
	Text         string          `json:"-"`
	ID           string          `json:"-"`
	Name         string          `json:"-"`
	Input        json.RawMessage `json:"-"`
	ToolUseID    string          `json:"-"`
	Content      string          `json:"-"`
	ContentParts []ContentPart   `json:"-"`
	Source       *ImageSource    `json:"-"`
}

// MarshalJSON serializes a ContentBlock.
// For tool_result blocks with ContentParts, the content field is an array.
// For image blocks, includes the source field.
func (cb ContentBlock) MarshalJSON() ([]byte, error) {
	if cb.Type == "tool_result" && len(cb.ContentParts) > 0 {
		return json.Marshal(struct {
			Type      string        `json:"type"`
			ToolUseID string        `json:"tool_use_id,omitempty"`
			Content   []ContentPart `json:"content"`
		}{cb.Type, cb.ToolUseID, cb.ContentParts})
	}

	if cb.Type == "image" && cb.Source != nil {
		return json.Marshal(struct {
			Type   string       `json:"type"`
			Source *ImageSource `json:"source"`
		}{cb.Type, cb.Source})
	}

	// Default: use struct tags for all other block types
	type plain struct {
		Type      string          `json:"type"`
		Text      string          `json:"text,omitempty"`
		ID        string          `json:"id,omitempty"`
		Name      string          `json:"name,omitempty"`
		Input     json.RawMessage `json:"input,omitempty"`
		ToolUseID string          `json:"tool_use_id,omitempty"`
		Content   string          `json:"content,omitempty"`
	}
	return json.Marshal(plain{
		Type:      cb.Type,
		Text:      cb.Text,
		ID:        cb.ID,
		Name:      cb.Name,
		Input:     cb.Input,
		ToolUseID: cb.ToolUseID,
		Content:   cb.Content,
	})
}

// UnmarshalJSON deserializes a ContentBlock from JSON.
func (cb *ContentBlock) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type      string          `json:"type"`
		Text      string          `json:"text,omitempty"`
		ID        string          `json:"id,omitempty"`
		Name      string          `json:"name,omitempty"`
		Input     json.RawMessage `json:"input,omitempty"`
		ToolUseID string          `json:"tool_use_id,omitempty"`
		Content   json.RawMessage `json:"content,omitempty"`
		Source    *ImageSource    `json:"source,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	cb.Type = raw.Type
	cb.Text = raw.Text
	cb.ID = raw.ID
	cb.Name = raw.Name
	cb.Input = raw.Input
	cb.ToolUseID = raw.ToolUseID
	cb.Source = raw.Source

	if raw.Content != nil {
		// Try string first
		var s string
		if err := json.Unmarshal(raw.Content, &s); err == nil {
			cb.Content = s
		} else {
			// Try content parts array
			var parts []ContentPart
			if err := json.Unmarshal(raw.Content, &parts); err == nil {
				cb.ContentParts = parts
			}
		}
	}
	return nil
}

// Message represents a chat message. Use Content for simple text messages
// or ContentBlocks for structured content (tool_use, tool_result).
type Message struct {
	Role          string         `json:"-"`
	Content       string         `json:"-"`
	ContentBlocks []ContentBlock `json:"-"`
}

type Request struct {
	Messages       []Message
	SystemPrompt   string
	DocumentTitle  string
	DocumentAuthor string
	DocumentDate   string
	TotalPages     int
	Tools          []Tool
}

type StreamEvent struct {
	Type      string
	Text      string
	Error     string
	ToolUseID string
	ToolName  string
	ToolInput string
}

// MarshalJSON serializes a Message. Uses content blocks if present, otherwise plain text.
func (m Message) MarshalJSON() ([]byte, error) {
	if len(m.ContentBlocks) > 0 {
		return json.Marshal(struct {
			Role    string         `json:"role"`
			Content []ContentBlock `json:"content"`
		}{m.Role, m.ContentBlocks})
	}
	return json.Marshal(struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{m.Role, m.Content})
}

// UnmarshalJSON deserializes a Message from JSON.
func (m *Message) UnmarshalJSON(data []byte) error {
	var raw struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Role = raw.Role
	// Try string first
	var s string
	if err := json.Unmarshal(raw.Content, &s); err == nil {
		m.Content = s
		return nil
	}
	// Otherwise try content blocks
	return json.Unmarshal(raw.Content, &m.ContentBlocks)
}

type Option func(*Client)

func WithModel(model string) Option {
	return func(c *Client) {
		if model != "" {
			c.Model = model
		}
	}
}

func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    "https://api.anthropic.com",
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		Model:      "claude-sonnet-4-20250514",
		Logger:     slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) logger() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

type apiRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

type sseData struct {
	Type         string          `json:"type"`
	Index        int             `json:"index"`
	Delta        sseDelta        `json:"delta"`
	ContentBlock sseContentBlock `json:"content_block"`
}

type sseDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	PartialJSON string `json:"partial_json"`
}

type sseContentBlock struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) Stream(ctx context.Context, req Request) (<-chan StreamEvent, error) {
	log := c.logger()

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = BuildSystemPromptFromContext(PromptContext{
			DocumentTitle:  req.DocumentTitle,
			DocumentAuthor: req.DocumentAuthor,
			DocumentDate:   req.DocumentDate,
			TotalPages:     req.TotalPages,
		})
	}

	body := apiRequest{
		Model:     c.Model,
		MaxTokens: 4096,
		Stream:    true,
		System:    systemPrompt,
		Messages:  req.Messages,
		Tools:     req.Tools,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	log.Info("stream starting", "model", c.Model, "messages", len(req.Messages))
	log.Debug("request details",
		"system_prompt_length", len(systemPrompt),
		"tool_count", len(req.Tools),
		"message_count", len(req.Messages),
	)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		log.Error("stream request failed", "error", err)
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Error("anthropic api error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("anthropic api: status %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		c.readSSE(ctx, log, resp, ch)
		log.Debug("stream completed")
	}()

	return ch, nil
}

func (c *Client) readSSE(ctx context.Context, log *slog.Logger, resp *http.Response, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(resp.Body)

	// Track in-progress tool_use blocks by index
	type toolBlock struct {
		id    string
		name  string
		input strings.Builder
	}
	toolBlocks := make(map[int]*toolBlock)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var parsed sseData
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			continue
		}

		log.Debug("sse event", "type", parsed.Type)

		switch parsed.Type {
		case "content_block_start":
			if parsed.ContentBlock.Type == "tool_use" {
				toolBlocks[parsed.Index] = &toolBlock{
					id:   parsed.ContentBlock.ID,
					name: parsed.ContentBlock.Name,
				}
			}

		case "content_block_delta":
			if parsed.Delta.Type == "text_delta" {
				ch <- StreamEvent{Type: "content_block_delta", Text: parsed.Delta.Text}
			} else if parsed.Delta.Type == "input_json_delta" {
				if tb, ok := toolBlocks[parsed.Index]; ok {
					tb.input.WriteString(parsed.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			if tb, ok := toolBlocks[parsed.Index]; ok {
				ch <- StreamEvent{
					Type:      "tool_use",
					ToolUseID: tb.id,
					ToolName:  tb.name,
					ToolInput: tb.input.String(),
				}
				delete(toolBlocks, parsed.Index)
			}

		case "message_stop":
			ch <- StreamEvent{Type: "message_stop"}
			return
		}
	}
}
