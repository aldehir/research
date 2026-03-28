package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Model      string
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

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    "https://api.anthropic.com",
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		Model:      "claude-sonnet-4-20250514",
	}
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

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic api: status %d", resp.StatusCode)
	}

	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		c.readSSE(ctx, resp, ch)
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
