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
