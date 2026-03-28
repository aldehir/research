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
