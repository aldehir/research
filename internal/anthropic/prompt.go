package anthropic

import (
	"fmt"
	"strings"
)

const basePrompt = "You are a helpful research paper reading assistant. " +
	"Help the user understand academic papers, explain concepts, " +
	"and answer questions about the content. " +
	"Each user message includes their current viewer context (page number and visible text)."

// PromptContext holds all context for building the system prompt.
type PromptContext struct {
	DocumentTitle  string
	DocumentAuthor string
	DocumentDate   string
	TotalPages     int
}

func BuildSystemPrompt(_, _ string) string {
	return BuildSystemPromptFromContext(PromptContext{})
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
		b.WriteString(fmt.Sprintf("\nThe document has %d pages.", ctx.TotalPages))
	}

	return b.String()
}
