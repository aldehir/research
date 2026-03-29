package anthropic

import (
	"context"

	"github.com/aldehir/research/internal/chat"
)

// Adapter wraps an Anthropic Client and implements chat.Provider.
type Adapter struct {
	Client *Client
}

// NewAdapter creates a new Adapter that satisfies chat.Provider.
func NewAdapter(client *Client) *Adapter {
	return &Adapter{Client: client}
}

// Stream converts domain types to Anthropic wire format, streams the response,
// and converts events back to domain types.
func (a *Adapter) Stream(ctx context.Context, req chat.Request) (<-chan chat.StreamEvent, error) {
	anthReq := toAnthropicRequest(req)
	rawCh, err := a.Client.Stream(ctx, anthReq)
	if err != nil {
		return nil, err
	}
	ch := make(chan chat.StreamEvent)
	go func() {
		defer close(ch)
		for ev := range rawCh {
			ch <- fromAnthropicStreamEvent(ev)
		}
	}()
	return ch, nil
}
