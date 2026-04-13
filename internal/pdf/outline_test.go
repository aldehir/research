package pdf

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func qpdfAvailable() bool {
	_, err := exec.LookPath("qpdf")
	return err == nil
}

func createTestPDFWithBookmarks(t *testing.T, path string) {
	t.Helper()
	doc := fpdf.New("P", "mm", "Letter", "")

	doc.AddPage()
	doc.Bookmark("Chapter 1", 0, -1)
	doc.SetFont("Helvetica", "", 12)
	doc.Text(10, 20, "Chapter 1 content")

	doc.AddPage()
	doc.Bookmark("Section 1.1", 1, -1)
	doc.Text(10, 20, "Section 1.1 content")

	doc.AddPage()
	doc.Bookmark("Section 1.2", 1, -1)
	doc.Text(10, 20, "Section 1.2 content")

	doc.AddPage()
	doc.Bookmark("Chapter 2", 0, -1)
	doc.Text(10, 20, "Chapter 2 content")

	require.NoError(t, doc.OutputFileAndClose(path))
}

func TestExtractOutline(t *testing.T) {
	if !qpdfAvailable() {
		t.Skip("qpdf not available")
	}

	path := filepath.Join(t.TempDir(), "outline.pdf")
	createTestPDFWithBookmarks(t, path)

	entries, err := ExtractOutline(path)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "Chapter 1", entries[0].Title)
	assert.Equal(t, 1, entries[0].Page)
	require.Len(t, entries[0].Children, 2)
	assert.Equal(t, "Section 1.1", entries[0].Children[0].Title)
	assert.Equal(t, 2, entries[0].Children[0].Page)
	assert.Equal(t, "Section 1.2", entries[0].Children[1].Title)
	assert.Equal(t, 3, entries[0].Children[1].Page)

	assert.Equal(t, "Chapter 2", entries[1].Title)
	assert.Equal(t, 4, entries[1].Page)
	assert.Empty(t, entries[1].Children)
}

func TestExtractOutline_NoOutline(t *testing.T) {
	if !qpdfAvailable() {
		t.Skip("qpdf not available")
	}

	path := filepath.Join(t.TempDir(), "no-outline.pdf")
	createTestPDFWithText(t, path, "no bookmarks here")

	entries, err := ExtractOutline(path)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestExtractOutline_FileNotFound(t *testing.T) {
	if !qpdfAvailable() {
		t.Skip("qpdf not available")
	}

	_, err := ExtractOutline("/nonexistent/path.pdf")
	require.Error(t, err)
}

func TestFormatOutline(t *testing.T) {
	entries := []OutlineEntry{
		{
			Title: "Chapter 1",
			Page:  1,
			Children: []OutlineEntry{
				{Title: "Section 1.1", Page: 3},
				{Title: "Section 1.2", Page: 7},
			},
		},
		{Title: "Chapter 2", Page: 15},
	}

	got := FormatOutline(entries)
	expected := "- Chapter 1 (p. 1)\n  - Section 1.1 (p. 3)\n  - Section 1.2 (p. 7)\n- Chapter 2 (p. 15)"
	assert.Equal(t, expected, got)
}

func TestFormatOutline_Empty(t *testing.T) {
	assert.Equal(t, "", FormatOutline(nil))
	assert.Equal(t, "", FormatOutline([]OutlineEntry{}))
}

func TestFormatOutline_NoPageNumbers(t *testing.T) {
	entries := []OutlineEntry{
		{Title: "Introduction"},
		{Title: "Conclusion"},
	}

	got := FormatOutline(entries)
	expected := "- Introduction\n- Conclusion"
	assert.Equal(t, expected, got)
}
