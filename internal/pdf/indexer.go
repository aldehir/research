package pdf

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aldehir/research/internal/store"
)

// Indexer extracts text from PDFs and stores it in the paper_pages table.
type Indexer struct {
	db *sql.DB
}

// NewIndexer creates an Indexer that writes page text to the given database.
func NewIndexer(db *sql.DB) *Indexer {
	return &Indexer{db: db}
}

// IndexPaper extracts all pages from the PDF at path and writes them to paper_pages.
// It also sets the text_indexed_at timestamp on the paper.
func (ix *Indexer) IndexPaper(paperID, pdfPath string) error {
	out, err := exec.Command("pdftotext", "-layout", pdfPath, "-").Output()
	if err != nil {
		return fmt.Errorf("pdftotext: %w", err)
	}

	pages := strings.Split(string(out), "\f")

	for i, pageText := range pages {
		text := strings.TrimRight(pageText, "\n")
		if i == len(pages)-1 && text == "" {
			break // trailing empty split after last form feed
		}
		if err := store.UpsertPageText(ix.db, paperID, i+1, text); err != nil {
			return fmt.Errorf("upsert page %d: %w", i+1, err)
		}
	}

	return store.SetTextIndexedAt(ix.db, paperID, time.Now().UTC().Format(time.RFC3339))
}

// IndexUnindexed finds all papers without extracted text and indexes them.
// The pathFn callback returns the PDF file path for a given paper ID.
func (ix *Indexer) IndexUnindexed(pathFn func(paperID string) string) error {
	ids, err := store.ListUnindexedPaperIDs(ix.db)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := ix.IndexPaper(id, pathFn(id)); err != nil {
			return fmt.Errorf("index paper %s: %w", id, err)
		}
	}
	return nil
}
