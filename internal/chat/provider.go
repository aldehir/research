package chat

import "context"

// Provider streams chat completions from an LLM.
// Implementations translate between domain types and their specific wire formats.
type Provider interface {
	Stream(ctx context.Context, req Request) (<-chan StreamEvent, error)
}
