package chat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPart_JSONRoundTrip_Text(t *testing.T) {
	p := Part{Kind: PartText, Text: "Hello world"}
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Part
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, PartText, got.Kind)
	assert.Equal(t, "Hello world", got.Text)
	assert.Nil(t, got.Image)
	assert.Nil(t, got.ToolCall)
	assert.Nil(t, got.ToolResult)
}

func TestPart_JSONRoundTrip_Image(t *testing.T) {
	p := Part{Kind: PartImage, Image: &Image{MediaType: "image/png", Data: "iVBORw0KGgo="}}
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Part
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, PartImage, got.Kind)
	require.NotNil(t, got.Image)
	assert.Equal(t, "image/png", got.Image.MediaType)
	assert.Equal(t, "iVBORw0KGgo=", got.Image.Data)
}

func TestPart_JSONRoundTrip_ToolCall(t *testing.T) {
	p := Part{Kind: PartToolCall, ToolCall: &ToolCall{
		ID:    "call_1",
		Name:  "read_page",
		Input: json.RawMessage(`{"page":3}`),
	}}
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Part
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, PartToolCall, got.Kind)
	require.NotNil(t, got.ToolCall)
	assert.Equal(t, "call_1", got.ToolCall.ID)
	assert.Equal(t, "read_page", got.ToolCall.Name)
	assert.JSONEq(t, `{"page":3}`, string(got.ToolCall.Input))
}

func TestPart_JSONRoundTrip_ToolResult_Text(t *testing.T) {
	p := Part{Kind: PartToolResult, ToolResult: &ToolResult{
		ToolCallID: "call_1",
		Content:    "Page text here",
	}}
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Part
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, PartToolResult, got.Kind)
	require.NotNil(t, got.ToolResult)
	assert.Equal(t, "call_1", got.ToolResult.ToolCallID)
	assert.Equal(t, "Page text here", got.ToolResult.Content)
	assert.Nil(t, got.ToolResult.Image)
}

func TestPart_JSONRoundTrip_ToolResult_Image(t *testing.T) {
	p := Part{Kind: PartToolResult, ToolResult: &ToolResult{
		ToolCallID: "call_snap",
		Image:      &Image{MediaType: "image/png", Data: "iVBORw0KGgo="},
	}}
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Part
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, PartToolResult, got.Kind)
	require.NotNil(t, got.ToolResult)
	assert.Equal(t, "call_snap", got.ToolResult.ToolCallID)
	require.NotNil(t, got.ToolResult.Image)
	assert.Equal(t, "image/png", got.ToolResult.Image.MediaType)
}

func TestMessage_JSONRoundTrip(t *testing.T) {
	msg := Message{
		Role: RoleAssistant,
		Parts: []Part{
			{Kind: PartText, Text: "Let me check"},
			{Kind: PartToolCall, ToolCall: &ToolCall{ID: "c1", Name: "read_page", Input: json.RawMessage(`{"page":1}`)}},
		},
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var got Message
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, RoleAssistant, got.Role)
	require.Len(t, got.Parts, 2)
	assert.Equal(t, PartText, got.Parts[0].Kind)
	assert.Equal(t, "Let me check", got.Parts[0].Text)
	assert.Equal(t, PartToolCall, got.Parts[1].Kind)
	assert.Equal(t, "read_page", got.Parts[1].ToolCall.Name)
}
