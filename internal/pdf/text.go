package pdf

import (
	"bytes"
	"fmt"
	"strings"

	gopdf "github.com/ledongthuc/pdf"
)

// SearchResult represents a text search match within a PDF.
type SearchResult struct {
	Page    int    `json:"page"`
	Snippet string `json:"snippet"`
}

// PageCount returns the number of pages in a PDF file.
func PageCount(path string) (int, error) {
	f, r, err := gopdf.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()
	return r.NumPage(), nil
}

// ExtractPageText extracts the text content of a specific page (1-based).
func ExtractPageText(path string, pageNum int) (string, error) {
	f, r, err := gopdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	if pageNum < 1 || pageNum > r.NumPage() {
		return "", fmt.Errorf("page %d out of range (1-%d)", pageNum, r.NumPage())
	}

	page := r.Page(pageNum)
	if page.V.IsNull() {
		return "", fmt.Errorf("page %d not found", pageNum)
	}

	var buf bytes.Buffer
	texts := page.Content().Text
	for _, t := range texts {
		buf.WriteString(t.S)
	}
	return buf.String(), nil
}

// SearchText searches all pages of a PDF for the given query string.
// Returns matching pages with text snippets.
func SearchText(path string, query string) ([]SearchResult, error) {
	f, r, err := gopdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	queryLower := strings.ToLower(query)
	var results []SearchResult

	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		var buf bytes.Buffer
		for _, t := range page.Content().Text {
			buf.WriteString(t.S)
		}
		pageText := buf.String()

		if idx := strings.Index(strings.ToLower(pageText), queryLower); idx >= 0 {
			// Extract snippet around the match
			start := max(0, idx-50)
			end := min(len(pageText), idx+len(query)+50)
			snippet := pageText[start:end]
			results = append(results, SearchResult{
				Page:    i,
				Snippet: snippet,
			})
		}
	}

	return results, nil
}
