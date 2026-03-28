package anthropic

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt_BaseOnly(t *testing.T) {
	prompt := BuildSystemPrompt("", "")
	assert.Contains(t, prompt, "research paper")
}

func TestBuildSystemPrompt_WithSelectedText(t *testing.T) {
	prompt := BuildSystemPrompt("E=mc^2", "")
	assert.Contains(t, prompt, "research paper")
	assert.Contains(t, prompt, "E=mc^2")
}

func TestBuildSystemPrompt_WithSurroundingText(t *testing.T) {
	prompt := BuildSystemPrompt("", "The theory of relativity...")
	assert.Contains(t, prompt, "research paper")
	assert.Contains(t, prompt, "The theory of relativity...")
}

func TestBuildSystemPrompt_WithBoth(t *testing.T) {
	prompt := BuildSystemPrompt("E=mc^2", "The theory of relativity...")
	assert.Contains(t, prompt, "E=mc^2")
	assert.Contains(t, prompt, "The theory of relativity...")

	// Selected text should appear before surrounding text
	selectedIdx := strings.Index(prompt, "E=mc^2")
	surroundingIdx := strings.Index(prompt, "The theory of relativity...")
	assert.Less(t, selectedIdx, surroundingIdx)
}

func TestBuildSystemPrompt_WithDocumentContext(t *testing.T) {
	ctx := PromptContext{
		DocumentTitle:  "Attention Is All You Need",
		DocumentAuthor: "Vaswani et al.",
		CurrentPage:    5,
		TotalPages:     12,
	}
	prompt := BuildSystemPromptFromContext(ctx)
	assert.Contains(t, prompt, "Attention Is All You Need")
	assert.Contains(t, prompt, "Vaswani et al.")
	assert.Contains(t, prompt, "5")
	assert.Contains(t, prompt, "12")
}

func TestBuildSystemPrompt_ContextWithSelectedText(t *testing.T) {
	ctx := PromptContext{
		DocumentTitle: "Test Paper",
		SelectedText:  "key finding",
		CurrentPage:   3,
		TotalPages:    10,
	}
	prompt := BuildSystemPromptFromContext(ctx)
	assert.Contains(t, prompt, "Test Paper")
	assert.Contains(t, prompt, "key finding")
}

func TestBuildSystemPrompt_ContextMinimalFields(t *testing.T) {
	ctx := PromptContext{}
	prompt := BuildSystemPromptFromContext(ctx)
	assert.Contains(t, prompt, "research paper")
	// Should not contain metadata section when no metadata
	assert.NotContains(t, prompt, "Document:")
}
