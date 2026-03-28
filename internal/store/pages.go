package store

import (
	"crypto/rand"
	"database/sql"
	"fmt"
)

// PageSearchResult represents a search hit within a paper's indexed pages.
type PageSearchResult struct {
	PageNum int    `json:"page"`
	Snippet string `json:"snippet"`
}

// UpsertPageText inserts or updates the text content for a specific page of a paper.
func UpsertPageText(db *sql.DB, paperID string, pageNum int, text string) error {
	id, err := generateID()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO paper_pages (id, paper_id, page_num, text_content)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(paper_id, page_num) DO UPDATE SET text_content = excluded.text_content`,
		id, paperID, pageNum, text,
	)
	return err
}

// GetPageText returns the extracted text for a specific page. Returns ErrNotFound
// if the page has not been indexed.
func GetPageText(db *sql.DB, paperID string, pageNum int) (string, error) {
	var text string
	err := db.QueryRow(
		`SELECT text_content FROM paper_pages WHERE paper_id = ? AND page_num = ?`,
		paperID, pageNum,
	).Scan(&text)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return text, err
}

// SearchPageText performs a full-text search over a paper's indexed pages using FTS5.
// Results are ordered by page number.
func SearchPageText(db *sql.DB, paperID string, query string) ([]PageSearchResult, error) {
	rows, err := db.Query(
		`SELECT pp.page_num, snippet(paper_pages_fts, 0, '', '', '...', 32)
		 FROM paper_pages_fts
		 JOIN paper_pages pp ON pp.rowid = paper_pages_fts.rowid
		 WHERE paper_pages_fts MATCH ?
		   AND pp.paper_id = ?
		 ORDER BY pp.page_num`,
		query, paperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PageSearchResult
	for rows.Next() {
		var r PageSearchResult
		if err := rows.Scan(&r.PageNum, &r.Snippet); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SetTextIndexedAt marks a paper as having its text extracted at the given timestamp.
func SetTextIndexedAt(db *sql.DB, paperID string, timestamp string) error {
	_, err := db.Exec(
		`UPDATE papers SET text_indexed_at = ? WHERE id = ?`,
		timestamp, paperID,
	)
	return err
}

// ListUnindexedPaperIDs returns paper IDs that have not yet had their text extracted.
func ListUnindexedPaperIDs(db *sql.DB) ([]string, error) {
	rows, err := db.Query(
		`SELECT id FROM papers WHERE text_indexed_at IS NULL ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
