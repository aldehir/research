package pdf

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// OutlineEntry represents a single entry in a PDF document outline (bookmarks/TOC).
type OutlineEntry struct {
	Title    string         `json:"title"`
	Page     int            `json:"page,omitempty"`
	Children []OutlineEntry `json:"children,omitempty"`
}

// ExtractOutline extracts the document outline (bookmarks) from a PDF file
// using qpdf. Returns nil with no error if the PDF has no outline.
func ExtractOutline(path string) ([]OutlineEntry, error) {
	out, err := exec.Command("qpdf", "--json=2", "--json-key=outlines", path).Output()
	if err != nil {
		return nil, fmt.Errorf("qpdf outline extraction: %w", err)
	}

	var result struct {
		Outlines []qpdfOutline `json:"outlines"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse qpdf output: %w", err)
	}

	entries := convertOutlines(result.Outlines)
	return entries, nil
}

type qpdfOutline struct {
	Title string         `json:"title"`
	Page  int            `json:"destpageposfrom1"`
	Kids  []qpdfOutline  `json:"kids"`
}

func convertOutlines(items []qpdfOutline) []OutlineEntry {
	if len(items) == 0 {
		return nil
	}
	entries := make([]OutlineEntry, len(items))
	for i, item := range items {
		entries[i] = OutlineEntry{
			Title:    item.Title,
			Page:     item.Page,
			Children: convertOutlines(item.Kids),
		}
	}
	return entries
}

// FormatOutline formats outline entries as an indented list for use in prompts.
// Returns an empty string if entries is empty.
func FormatOutline(entries []OutlineEntry) string {
	if len(entries) == 0 {
		return ""
	}
	var b strings.Builder
	writeOutlineEntries(&b, entries, 0)
	return strings.TrimRight(b.String(), "\n")
}

func writeOutlineEntries(b *strings.Builder, entries []OutlineEntry, depth int) {
	for _, e := range entries {
		for range depth {
			b.WriteString("  ")
		}
		b.WriteString("- ")
		b.WriteString(e.Title)
		if e.Page > 0 {
			fmt.Fprintf(b, " (p. %d)", e.Page)
		}
		b.WriteByte('\n')
		writeOutlineEntries(b, e.Children, depth+1)
	}
}
