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

func TestBuildSystemPrompt_WithOutline(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{
		DocumentTitle: "Test Doc",
		TotalPages:    10,
		Outline:       "- Chapter 1 (p. 1)\n  - Section 1.1 (p. 3)\n- Chapter 2 (p. 7)",
	})
	assert.Contains(t, prompt, "Document outline")
	assert.Contains(t, prompt, "read_page or search_pdf")
	assert.Contains(t, prompt, "- Chapter 1 (p. 1)")
	assert.Contains(t, prompt, "  - Section 1.1 (p. 3)")
}

func TestBuildSystemPrompt_OmitsOutlineWhenEmpty(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{
		DocumentTitle: "Test Doc",
	})
	assert.NotContains(t, prompt, "Document outline")
}

func TestBuildSystemPrompt_WithCustomInstructions(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{
		DocumentTitle:      "Test Doc",
		CustomInstructions: "Always respond in French. Focus on mathematical proofs.",
	})
	assert.Contains(t, prompt, "Custom instructions")
	assert.Contains(t, prompt, "Always respond in French. Focus on mathematical proofs.")
}

func TestBuildSystemPrompt_OmitsCustomInstructionsWhenEmpty(t *testing.T) {
	prompt := BuildSystemPrompt(PromptContext{})
	assert.NotContains(t, prompt, "Custom instructions")
}
