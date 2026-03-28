package pdf

import (
	"testing"

	"github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/require"
)

// createTestPDFWithText creates a valid PDF at path with the given text on page 1.
func createTestPDFWithText(t *testing.T, path string, text string) {
	t.Helper()
	doc := fpdf.New("P", "mm", "Letter", "")
	doc.AddPage()
	doc.SetFont("Helvetica", "", 12)
	doc.Text(10, 20, text)
	require.NoError(t, doc.OutputFileAndClose(path))
}

// createTestPDFMultiPage creates a valid PDF with the specified number of pages.
func createTestPDFMultiPage(t *testing.T, path string, pages int) {
	t.Helper()
	doc := fpdf.New("P", "mm", "Letter", "")
	for i := 1; i <= pages; i++ {
		doc.AddPage()
		doc.SetFont("Helvetica", "", 12)
		doc.Text(10, 20, "Page content")
	}
	require.NoError(t, doc.OutputFileAndClose(path))
}

// createTestPDFWithMetadata creates a PDF with document metadata set.
func createTestPDFWithMetadata(t *testing.T, path string) {
	t.Helper()
	doc := fpdf.New("P", "mm", "Letter", "")
	doc.SetTitle("Test Document Title", true)
	doc.SetAuthor("Test Author", true)
	doc.SetSubject("Test Subject", true)
	doc.AddPage()
	doc.SetFont("Helvetica", "", 12)
	doc.Text(10, 20, "content")
	require.NoError(t, doc.OutputFileAndClose(path))
}

// createTestPDFSeparateWords creates a PDF with each word as a separate text element,
// simulating how real PDFs typically encode text.
func createTestPDFSeparateWords(t *testing.T, path string) {
	t.Helper()
	doc := fpdf.New("P", "mm", "Letter", "")
	doc.AddPage()
	doc.SetFont("Helvetica", "", 12)
	// Place words at separate horizontal positions on the same line
	doc.Text(10, 20, "Hello")
	doc.Text(27, 20, "World")
	doc.Text(46, 20, "Test")
	// Second line
	doc.Text(10, 30, "Second")
	doc.Text(32, 30, "line")
	require.NoError(t, doc.OutputFileAndClose(path))
}
