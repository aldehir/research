package anthropic

import "encoding/json"

// Tool defines an Anthropic tool_use tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// PDFTools returns the tool definitions for PDF interaction.
func PDFTools() []Tool {
	return []Tool{
		{
			Name:        "search_pdf",
			Description: "Search the PDF document for a text query. Returns matching page numbers and text snippets.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The text to search for in the document"
					}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        "read_page",
			Description: "Read the full text content of a specific page from the PDF document.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"page": {
						"type": "integer",
						"description": "The 1-based page number to read"
					}
				},
				"required": ["page"]
			}`),
		},
		{
			Name:        "go_to_page",
			Description: "Navigate the user's PDF viewer to a specific page. Use this when referring the user to a particular page.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"page": {
						"type": "integer",
						"description": "The 1-based page number to navigate to"
					}
				},
				"required": ["page"]
			}`),
		},
	}
}
