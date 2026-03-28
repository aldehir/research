package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt_BaseOnly(t *testing.T) {
	prompt := BuildSystemPrompt("", "")
	assert.Contains(t, prompt, "research paper")
	assert.Contains(t, prompt, "viewer context")
}

func TestBuildSystemPrompt_WithDocumentContext(t *testing.T) {
	ctx := PromptContext{
		DocumentTitle:  "Attention Is All You Need",
		DocumentAuthor: "Vaswani et al.",
		TotalPages:     12,
	}
	prompt := BuildSystemPromptFromContext(ctx)
	assert.Contains(t, prompt, "Attention Is All You Need")
	assert.Contains(t, prompt, "Vaswani et al.")
	assert.Contains(t, prompt, "12 pages")
}

func TestBuildSystemPrompt_ContextMinimalFields(t *testing.T) {
	ctx := PromptContext{}
	prompt := BuildSystemPromptFromContext(ctx)
	assert.Contains(t, prompt, "research paper")
	assert.NotContains(t, prompt, "Document:")
}

func TestBuildSystemPrompt_DoesNotContainCurrentPage(t *testing.T) {
	ctx := PromptContext{
		DocumentTitle: "Test",
		TotalPages:    10,
	}
	prompt := BuildSystemPromptFromContext(ctx)
	// Current page should NOT be in the system prompt — model uses get_viewer_context
	assert.NotContains(t, prompt, "currently viewing")
}
