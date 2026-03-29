package chat

import "encoding/json"

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
		{
			Name:        "snapshot_page",
			Description: "Render a PDF page as an image for visual inspection. Use this to see charts, figures, diagrams, tables, or any visual content that text extraction might miss.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"page": {
						"type": "integer",
						"description": "The 1-based page number to render as an image"
					}
				},
				"required": ["page"]
			}`),
		},
	}
}
