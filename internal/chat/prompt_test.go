package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt_BaseOnly(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{})
	assert.Contains(t, prompt, "reading companion")
	assert.Contains(t, prompt, "viewer context")
	assert.NotContains(t, prompt, "Document:")
}

func TestBuildSystemPrompt_WithDocumentContext(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{
		DocumentTitle:  "Attention Is All You Need",
		DocumentAuthor: "Vaswani et al.",
		TotalPages:     12,
	})
	assert.Contains(t, prompt, "Attention Is All You Need")
	assert.Contains(t, prompt, "Vaswani et al.")
	assert.Contains(t, prompt, "12 pages")
}

func TestBuildSystemPrompt_FullMetadata(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{
		DocumentTitle:  "Deep Residual Learning",
		DocumentAuthor: "He et al.",
		DocumentDate:   "2015",
		TotalPages:     9,
	})
	assert.Contains(t, prompt, "Deep Residual Learning")
	assert.Contains(t, prompt, "He et al.")
	assert.Contains(t, prompt, "2015")
	assert.Contains(t, prompt, "9 pages")
}

func TestBuildSystemPrompt_MentionsLua(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{})
	assert.Contains(t, prompt, "Lua")
	assert.Contains(t, prompt, "```lua")
}

func TestBuildSystemPrompt_VerticalFormatting(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{})
	assert.Contains(t, prompt, "narrow")
	assert.Contains(t, prompt, "wide table")
}
