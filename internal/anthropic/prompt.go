package anthropic

import (
	"fmt"
	"strings"
)

const basePrompt = "You are a helpful research paper reading assistant. " +
	"Help the user understand academic papers, explain concepts, " +
	"and answer questions about the content."

// PromptContext holds all context for building the system prompt.
type PromptContext struct {
	DocumentTitle  string
	DocumentAuthor string
	DocumentDate   string
	CurrentPage    int
	TotalPages     int
	SelectedText   string
	SurroundingText string
}

func BuildSystemPrompt(selectedText, surroundingText string) string {
	return BuildSystemPromptFromContext(PromptContext{
		SelectedText:    selectedText,
		SurroundingText: surroundingText,
	})
}

func BuildSystemPromptFromContext(ctx PromptContext) string {
	var b strings.Builder
	b.WriteString(basePrompt)

	if ctx.DocumentTitle != "" || ctx.DocumentAuthor != "" || ctx.DocumentDate != "" {
		b.WriteString("\n\nDocument:")
		if ctx.DocumentTitle != "" {
			b.WriteString(" ")
			b.WriteString(ctx.DocumentTitle)
		}
		if ctx.DocumentAuthor != "" {
			b.WriteString(" by ")
			b.WriteString(ctx.DocumentAuthor)
		}
		if ctx.DocumentDate != "" {
			b.WriteString(" (")
			b.WriteString(ctx.DocumentDate)
			b.WriteString(")")
		}
	}

	if ctx.TotalPages > 0 {
		b.WriteString(fmt.Sprintf("\nThe user is currently viewing page %d of %d.", ctx.CurrentPage, ctx.TotalPages))
	}

	if ctx.SelectedText != "" {
		b.WriteString("\n\nThe user has selected the following text from the paper:\n> ")
		b.WriteString(ctx.SelectedText)
	}

	if ctx.SurroundingText != "" {
		b.WriteString("\n\nSurrounding context from the paper:\n")
		b.WriteString(ctx.SurroundingText)
	}

	return b.String()
}
