package chat

import "encoding/json"

// Role represents a message participant.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// PartKind discriminates the content type within a message part.
type PartKind string

const (
	PartText       PartKind = "text"
	PartThinking   PartKind = "thinking"
	PartImage      PartKind = "image"
	PartToolCall   PartKind = "tool_call"
	PartToolResult PartKind = "tool_result"
)

// Part is a discriminated union for content within a message.
// Exactly one payload field is populated based on Kind.
type Part struct {
	Kind       PartKind    `json:"kind"`
	Text       string      `json:"text,omitempty"`
	Image      *Image      `json:"image,omitempty"`
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// Image holds base64-encoded image data.
type Image struct {
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ToolCall represents an assistant requesting a tool invocation.
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult is the response to a ToolCall.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content,omitempty"`
	Image      *Image `json:"image,omitempty"`
}

// Message is a single turn in a conversation.
type Message struct {
	Role  Role   `json:"role"`
	Parts []Part `json:"parts"`
}

// Tool defines a tool available to the model.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// Request is the input to a Provider's Stream method.
type Request struct {
	SystemPrompt string
	Messages     []Message
	Tools        []Tool
	MaxTokens    int
}

// EventKind discriminates stream events.
type EventKind string

const (
	EventDelta    EventKind = "delta"
	EventThinking EventKind = "thinking"
	EventToolCall EventKind = "tool_call"
	EventDone     EventKind = "done"
	EventError    EventKind = "error"
)

// StreamEvent is emitted during streaming.
type StreamEvent struct {
	Kind     EventKind
	Text     string
	ToolCall *ToolCall
	Error    string
}
