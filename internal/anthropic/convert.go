package anthropic

import (
	"encoding/json"

	"github.com/aldehir/research/internal/chat"
)

// toAnthropicRequest converts a domain chat.Request to an Anthropic API Request.
func toAnthropicRequest(req chat.Request) Request {
	msgs := make([]Message, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = toAnthropicMessage(m)
	}

	tools := make([]Tool, len(req.Tools))
	for i, t := range req.Tools {
		tools[i] = Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}

	return Request{
		SystemPrompt: req.SystemPrompt,
		Messages:     msgs,
		Tools:        tools,
	}
}

// toAnthropicMessage converts a domain chat.Message to an Anthropic API Message.
// Maps domain Part types to Anthropic content blocks:
//   - PartText → ContentBlock{Type: "text"}
//   - PartImage → ContentBlock{Type: "image"}
//   - PartToolCall → ContentBlock{Type: "tool_use"}
//   - PartToolResult → ContentBlock{Type: "tool_result"}
func toAnthropicMessage(m chat.Message) Message {
	// Simple case: single text part → plain string content
	if len(m.Parts) == 1 && m.Parts[0].Kind == chat.PartText {
		return Message{
			Role:    string(m.Role),
			Content: m.Parts[0].Text,
		}
	}

	blocks := make([]ContentBlock, 0, len(m.Parts))
	for _, p := range m.Parts {
		blocks = append(blocks, toAnthropicContentBlock(p))
	}
	return Message{
		Role:          string(m.Role),
		ContentBlocks: blocks,
	}
}

func toAnthropicContentBlock(p chat.Part) ContentBlock {
	switch p.Kind {
	case chat.PartText:
		return ContentBlock{Type: "text", Text: p.Text}
	case chat.PartImage:
		return ContentBlock{
			Type: "image",
			Source: &ImageSource{
				Type:      "base64",
				MediaType: p.Image.MediaType,
				Data:      p.Image.Data,
			},
		}
	case chat.PartToolCall:
		return ContentBlock{
			Type:  "tool_use",
			ID:    p.ToolCall.ID,
			Name:  p.ToolCall.Name,
			Input: json.RawMessage(p.ToolCall.Input),
		}
	case chat.PartToolResult:
		cb := ContentBlock{
			Type:      "tool_result",
			ToolUseID: p.ToolResult.ToolCallID,
		}
		if p.ToolResult.Image != nil {
			cb.ContentParts = []ContentPart{
				{
					Type: "image",
					Source: &ImageSource{
						Type:      "base64",
						MediaType: p.ToolResult.Image.MediaType,
						Data:      p.ToolResult.Image.Data,
					},
				},
			}
		} else {
			cb.Content = p.ToolResult.Content
		}
		return cb
	default:
		return ContentBlock{Type: "text", Text: p.Text}
	}
}

// fromAnthropicStreamEvent converts an Anthropic StreamEvent to a domain chat.StreamEvent.
func fromAnthropicStreamEvent(ev StreamEvent) chat.StreamEvent {
	switch ev.Type {
	case "content_block_delta":
		return chat.StreamEvent{Kind: chat.EventDelta, Text: ev.Text}
	case "tool_use":
		return chat.StreamEvent{
			Kind: chat.EventToolCall,
			ToolCall: &chat.ToolCall{
				ID:    ev.ToolUseID,
				Name:  ev.ToolName,
				Input: json.RawMessage(ev.ToolInput),
			},
		}
	case "message_stop":
		return chat.StreamEvent{Kind: chat.EventDone}
	default:
		return chat.StreamEvent{Kind: chat.EventDone}
	}
}
