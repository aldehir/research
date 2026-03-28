package anthropic

import "strings"

const basePrompt = "You are a helpful research paper reading assistant. " +
	"Help the user understand academic papers, explain concepts, " +
	"and answer questions about the content."

func BuildSystemPrompt(selectedText, surroundingText string) string {
	var b strings.Builder
	b.WriteString(basePrompt)

	if selectedText != "" {
		b.WriteString("\n\nThe user has selected the following text from the paper:\n> ")
		b.WriteString(selectedText)
	}

	if surroundingText != "" {
		b.WriteString("\n\nSurrounding context from the paper:\n")
		b.WriteString(surroundingText)
	}

	return b.String()
}
