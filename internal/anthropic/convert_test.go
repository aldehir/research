package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/aldehir/research/internal/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAnthropicMessage_PlainText(t *testing.T) {
	m := chat.Message{
		Role:  chat.RoleUser,
		Parts: []chat.Part{{Kind: chat.PartText, Text: "Hello"}},
	}
	got := toAnthropicMessage(m)
	assert.Equal(t, "user", got.Role)
	assert.Equal(t, "Hello", got.Content)
	assert.Empty(t, got.ContentBlocks)
}

func TestToAnthropicMessage_Multimodal(t *testing.T) {
	m := chat.Message{
		Role: chat.RoleUser,
		Parts: []chat.Part{
			{Kind: chat.PartText, Text: "What is this?"},
			{Kind: chat.PartImage, Image: &chat.Image{MediaType: "image/png", Data: "abc123"}},
		},
	}
	got := toAnthropicMessage(m)
	assert.Equal(t, "user", got.Role)
	assert.Empty(t, got.Content)
	require.Len(t, got.ContentBlocks, 2)
	assert.Equal(t, "text", got.ContentBlocks[0].Type)
	assert.Equal(t, "What is this?", got.ContentBlocks[0].Text)
	assert.Equal(t, "image", got.ContentBlocks[1].Type)
	require.NotNil(t, got.ContentBlocks[1].Source)
	assert.Equal(t, "image/png", got.ContentBlocks[1].Source.MediaType)
}

func TestToAnthropicMessage_ToolCalls(t *testing.T) {
	m := chat.Message{
		Role: chat.RoleAssistant,
		Parts: []chat.Part{
			{Kind: chat.PartText, Text: "Let me check"},
			{Kind: chat.PartToolCall, ToolCall: &chat.ToolCall{
				ID: "toolu_1", Name: "read_page", Input: json.RawMessage(`{"page":3}`),
			}},
		},
	}
	got := toAnthropicMessage(m)
	assert.Equal(t, "assistant", got.Role)
	require.Len(t, got.ContentBlocks, 2)
	assert.Equal(t, "text", got.ContentBlocks[0].Type)
	assert.Equal(t, "tool_use", got.ContentBlocks[1].Type)
	assert.Equal(t, "toolu_1", got.ContentBlocks[1].ID)
	assert.Equal(t, "read_page", got.ContentBlocks[1].Name)
	assert.JSONEq(t, `{"page":3}`, string(got.ContentBlocks[1].Input))
}

func TestToAnthropicMessage_ToolResult_Text(t *testing.T) {
	m := chat.Message{
		Role: chat.RoleUser,
		Parts: []chat.Part{
			{Kind: chat.PartToolResult, ToolResult: &chat.ToolResult{
				ToolCallID: "toolu_1", Content: "Page text here",
			}},
		},
	}
	got := toAnthropicMessage(m)
	assert.Equal(t, "user", got.Role)
	require.Len(t, got.ContentBlocks, 1)
	assert.Equal(t, "tool_result", got.ContentBlocks[0].Type)
	assert.Equal(t, "toolu_1", got.ContentBlocks[0].ToolUseID)
	assert.Equal(t, "Page text here", got.ContentBlocks[0].Content)
	assert.Empty(t, got.ContentBlocks[0].ContentParts)
}

func TestToAnthropicMessage_ToolResult_Image(t *testing.T) {
	m := chat.Message{
		Role: chat.RoleUser,
		Parts: []chat.Part{
			{Kind: chat.PartToolResult, ToolResult: &chat.ToolResult{
				ToolCallID: "toolu_snap",
				Image:      &chat.Image{MediaType: "image/png", Data: "iVBORw0KGgo="},
			}},
		},
	}
	got := toAnthropicMessage(m)
	require.Len(t, got.ContentBlocks, 1)
	cb := got.ContentBlocks[0]
	assert.Equal(t, "tool_result", cb.Type)
	assert.Equal(t, "toolu_snap", cb.ToolUseID)
	require.Len(t, cb.ContentParts, 1)
	assert.Equal(t, "image", cb.ContentParts[0].Type)
	require.NotNil(t, cb.ContentParts[0].Source)
	assert.Equal(t, "image/png", cb.ContentParts[0].Source.MediaType)
}

func TestFromAnthropicStreamEvent_Delta(t *testing.T) {
	ev := StreamEvent{Type: "content_block_delta", Text: "Hello"}
	got := fromAnthropicStreamEvent(ev)
	assert.Equal(t, chat.EventDelta, got.Kind)
	assert.Equal(t, "Hello", got.Text)
}

func TestFromAnthropicStreamEvent_ToolUse(t *testing.T) {
	ev := StreamEvent{
		Type:      "tool_use",
		ToolUseID: "toolu_1",
		ToolName:  "read_page",
		ToolInput: `{"page":5}`,
	}
	got := fromAnthropicStreamEvent(ev)
	assert.Equal(t, chat.EventToolCall, got.Kind)
	require.NotNil(t, got.ToolCall)
	assert.Equal(t, "toolu_1", got.ToolCall.ID)
	assert.Equal(t, "read_page", got.ToolCall.Name)
	assert.JSONEq(t, `{"page":5}`, string(got.ToolCall.Input))
}

func TestFromAnthropicStreamEvent_MessageStop(t *testing.T) {
	ev := StreamEvent{Type: "message_stop"}
	got := fromAnthropicStreamEvent(ev)
	assert.Equal(t, chat.EventDone, got.Kind)
}

func TestToAnthropicRequest_ConvertsAll(t *testing.T) {
	req := chat.Request{
		SystemPrompt: "Be helpful",
		Messages: []chat.Message{
			{Role: chat.RoleUser, Parts: []chat.Part{{Kind: chat.PartText, Text: "Hi"}}},
		},
		Tools: chat.PDFTools(),
	}
	got := toAnthropicRequest(req)
	assert.Equal(t, "Be helpful", got.SystemPrompt)
	require.Len(t, got.Messages, 1)
	assert.Equal(t, "user", got.Messages[0].Role)
	assert.Equal(t, "Hi", got.Messages[0].Content)
	assert.Len(t, got.Tools, 4)
	assert.Equal(t, "search_pdf", got.Tools[0].Name)
}
