package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt_BaseOnly(t *testing.T) {
	prompt := BuildSystemPrompt("", "")
	assert.Contains(t, prompt, "reading companion")
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
	assert.Contains(t, prompt, "reading companion")
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

func TestBuildSystemPrompt_ExplanatoryTone(t *testing.T) {
	prompt := BuildSystemPromptFromContext(PromptContext{})
	assert.Contains(t, prompt, "explanatory")
	assert.Contains(t, prompt, "page")
}

func TestBuildSystemPrompt_UncertaintyGuidance(t *testing.T) {
	prompt := BuildSystemPromptFromContext(PromptContext{})
	assert.Contains(t, prompt, "uncertain")
}

func TestBuildSystemPrompt_StructuredExplanations(t *testing.T) {
	prompt := BuildSystemPromptFromContext(PromptContext{})
	assert.Contains(t, prompt, "definition")
}

func TestBuildSystemPrompt_DocumentMetadataStillAppended(t *testing.T) {
	ctx := PromptContext{
		DocumentTitle:  "Deep Residual Learning",
		DocumentAuthor: "He et al.",
		DocumentDate:   "2015",
		TotalPages:     9,
	}
	prompt := BuildSystemPromptFromContext(ctx)
	assert.Contains(t, prompt, "Deep Residual Learning")
	assert.Contains(t, prompt, "He et al.")
	assert.Contains(t, prompt, "2015")
	assert.Contains(t, prompt, "9 pages")
}
