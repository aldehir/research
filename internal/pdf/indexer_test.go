package pdf

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/aldehir/research/internal/store"
	"github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func seedPaper(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	p := store.Paper{
		ID:        id,
		Title:     "Test Paper",
		FilePath:  "/papers/test.pdf",
		FileSize:  12345,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, store.CreatePaper(db, p))
}

func createTestPDFWithPages(t *testing.T, path string, pages []string) {
	t.Helper()
	doc := fpdf.New("P", "mm", "Letter", "")
	doc.SetFont("Helvetica", "", 12)
	for _, text := range pages {
		doc.AddPage()
		doc.Text(10, 20, text)
	}
	require.NoError(t, doc.OutputFileAndClose(path))
}

func TestIndexer_IndexPaper(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db, "paper-1")

	pdfPath := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFWithPages(t, pdfPath, []string{"Page one content", "Page two content"})

	idx := NewIndexer(db)
	err := idx.IndexPaper("paper-1", pdfPath)
	require.NoError(t, err)

	// Should have stored text for both pages
	text1, err := store.GetPageText(db, "paper-1", 1)
	require.NoError(t, err)
	assert.Contains(t, text1, "Page one content")

	text2, err := store.GetPageText(db, "paper-1", 2)
	require.NoError(t, err)
	assert.Contains(t, text2, "Page two content")
}

func TestIndexer_IndexPaper_SetsTimestamp(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db, "paper-1")

	pdfPath := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFWithPages(t, pdfPath, []string{"Content"})

	idx := NewIndexer(db)
	require.NoError(t, idx.IndexPaper("paper-1", pdfPath))

	// text_indexed_at should be set
	var ts *string
	err := db.QueryRow("SELECT text_indexed_at FROM papers WHERE id = ?", "paper-1").Scan(&ts)
	require.NoError(t, err)
	assert.NotNil(t, ts)
}

func TestIndexer_IndexPaper_SkipsAlreadyIndexed(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db, "paper-1")

	pdfPath := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFWithPages(t, pdfPath, []string{"Content"})

	idx := NewIndexer(db)
	require.NoError(t, idx.IndexPaper("paper-1", pdfPath))

	// Mark as indexed
	require.NoError(t, store.SetTextIndexedAt(db, "paper-1", "2026-03-28T12:00:00Z"))

	// IndexUnindexed should skip paper-1
	err := idx.IndexUnindexed(func(paperID string) string {
		return pdfPath
	})
	require.NoError(t, err)
}

func TestIndexer_IndexUnindexed(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db, "paper-1")
	seedPaper(t, db, "paper-2")

	dir := t.TempDir()
	pdf1 := filepath.Join(dir, "p1.pdf")
	pdf2 := filepath.Join(dir, "p2.pdf")
	createTestPDFWithPages(t, pdf1, []string{"Alpha content"})
	createTestPDFWithPages(t, pdf2, []string{"Beta content"})

	// Mark paper-2 as already indexed
	require.NoError(t, store.SetTextIndexedAt(db, "paper-2", "2026-03-28T12:00:00Z"))

	idx := NewIndexer(db)
	err := idx.IndexUnindexed(func(paperID string) string {
		if paperID == "paper-1" {
			return pdf1
		}
		return pdf2
	})
	require.NoError(t, err)

	// paper-1 should be indexed
	text, err := store.GetPageText(db, "paper-1", 1)
	require.NoError(t, err)
	assert.Contains(t, text, "Alpha content")

	// paper-2 should NOT have new page text (was already indexed)
	_, err = store.GetPageText(db, "paper-2", 1)
	assert.ErrorIs(t, err, store.ErrNotFound)
}
