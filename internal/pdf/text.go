package pdf

import (
	"fmt"
	"os/exec"
	"strconv"
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

// ExtractPageText extracts the text content of a specific page (1-based)
// using pdftotext with layout preservation.
func ExtractPageText(path string, pageNum int) (string, error) {
	count, err := PageCount(path)
	if err != nil {
		return "", err
	}
	if pageNum < 1 || pageNum > count {
		return "", fmt.Errorf("page %d out of range (1-%d)", pageNum, count)
	}

	pageStr := strconv.Itoa(pageNum)
	out, err := exec.Command("pdftotext", "-layout", "-f", pageStr, "-l", pageStr, path, "-").Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext: %w", err)
	}
	return strings.TrimRight(string(out), "\n\f"), nil
}

// ExtractRegionText extracts text from a rectangular region of a PDF page.
// Coordinates (x, y, w, h) are in PDF points (1/72 inch), origin at top-left.
func ExtractRegionText(path string, pageNum, x, y, w, h int) (string, error) {
	count, err := PageCount(path)
	if err != nil {
		return "", err
	}
	if pageNum < 1 || pageNum > count {
		return "", fmt.Errorf("page %d out of range (1-%d)", pageNum, count)
	}

	pageStr := strconv.Itoa(pageNum)
	out, err := exec.Command(
		"pdftotext", "-layout",
		"-f", pageStr, "-l", pageStr,
		"-x", strconv.Itoa(x),
		"-y", strconv.Itoa(y),
		"-W", strconv.Itoa(w),
		"-H", strconv.Itoa(h),
		path, "-",
	).Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext region: %w", err)
	}
	return strings.TrimRight(string(out), "\n\f"), nil
}

// SearchText searches all pages of a PDF for the given query string.
// Returns matching pages with text snippets.
func SearchText(path string, query string) ([]SearchResult, error) {
	count, err := PageCount(path)
	if err != nil {
		return nil, err
	}

	// Extract all text at once for search
	out, err := exec.Command("pdftotext", "-layout", path, "-").Output()
	if err != nil {
		return nil, fmt.Errorf("pdftotext: %w", err)
	}

	// pdftotext separates pages with form feed (\f)
	pages := strings.Split(string(out), "\f")
	queryLower := strings.ToLower(query)
	var results []SearchResult

	for i := 0; i < len(pages) && i < count; i++ {
		pageText := pages[i]
		if idx := strings.Index(strings.ToLower(pageText), queryLower); idx >= 0 {
			start := max(0, idx-50)
			end := min(len(pageText), idx+len(query)+50)
			snippet := pageText[start:end]
			results = append(results, SearchResult{
				Page:    i + 1,
				Snippet: strings.TrimSpace(snippet),
			})
		}
	}

	return results, nil
}
