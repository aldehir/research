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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Messages        []Message
	SystemPrompt    string
	SelectedText    string
	SurroundingText string
}

type StreamEvent struct {
	Type  string
	Text  string
	Error string
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
}

type sseData struct {
	Type  string   `json:"type"`
	Delta sseDelta `json:"delta"`
}

type sseDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c *Client) Stream(ctx context.Context, req Request) (<-chan StreamEvent, error) {
	log := c.logger()

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = BuildSystemPrompt(req.SelectedText, req.SurroundingText)
	}

	body := apiRequest{
		Model:     c.Model,
		MaxTokens: 4096,
		Stream:    true,
		System:    systemPrompt,
		Messages:  req.Messages,
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
		c.readSSE(ctx, resp, ch)
		log.Debug("stream completed")
	}()

	return ch, nil
}

func (c *Client) readSSE(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(resp.Body)

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

		ev := StreamEvent{Type: parsed.Type}

		switch parsed.Type {
		case "content_block_delta":
			if parsed.Delta.Type == "text_delta" {
				ev.Text = parsed.Delta.Text
			}
		case "message_stop":
			ch <- ev
			return
		}

		ch <- ev
	}
}
