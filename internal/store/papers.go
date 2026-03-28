package store

import (
	"database/sql"
	"errors"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// Paper represents a research paper record.
type Paper struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	FilePath      string  `json:"file_path"`
	FileSize      int64   `json:"file_size"`
	Author        *string `json:"author,omitempty"`
	Subject       *string `json:"subject,omitempty"`
	PublishedDate *string `json:"published_date,omitempty"`
	PageCount     *int    `json:"page_count,omitempty"`
	LastReadPage  *int    `json:"last_read_page,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// CreatePaper inserts a new paper record.
func CreatePaper(db *sql.DB, p Paper) error {
	_, err := db.Exec(
		`INSERT INTO papers (id, title, file_path, file_size, author, subject, published_date, page_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Title, p.FilePath, p.FileSize, p.Author, p.Subject, p.PublishedDate, p.PageCount, p.CreatedAt,
	)
	return err
}

// ListPapers returns all papers ordered by created_at descending.
func ListPapers(db *sql.DB) ([]Paper, error) {
	rows, err := db.Query(`SELECT id, title, file_path, file_size, author, subject, published_date, page_count, last_read_page, created_at FROM papers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var papers []Paper
	for rows.Next() {
		var p Paper
		if err := rows.Scan(&p.ID, &p.Title, &p.FilePath, &p.FileSize, &p.Author, &p.Subject, &p.PublishedDate, &p.PageCount, &p.LastReadPage, &p.CreatedAt); err != nil {
			return nil, err
		}
		papers = append(papers, p)
	}
	return papers, rows.Err()
}

// GetPaper returns a single paper by ID. Returns sql.ErrNoRows if not found.
func GetPaper(db *sql.DB, id string) (Paper, error) {
	var p Paper
	err := db.QueryRow(
		`SELECT id, title, file_path, file_size, author, subject, published_date, page_count, last_read_page, created_at FROM papers WHERE id = ?`, id,
	).Scan(&p.ID, &p.Title, &p.FilePath, &p.FileSize, &p.Author, &p.Subject, &p.PublishedDate, &p.PageCount, &p.LastReadPage, &p.CreatedAt)
	return p, err
}

// PaperMetadata holds optional metadata fields for updating a paper.
type PaperMetadata struct {
	Author        *string
	Subject       *string
	PublishedDate *string
	PageCount     *int
}

// UpdatePaperMetadata updates the metadata columns of a paper.
func UpdatePaperMetadata(db *sql.DB, id string, m PaperMetadata) error {
	_, err := db.Exec(
		`UPDATE papers SET author = ?, subject = ?, published_date = ?, page_count = ? WHERE id = ?`,
		m.Author, m.Subject, m.PublishedDate, m.PageCount, id,
	)
	return err
}

// UpdateReadingPosition sets the last_read_page for a paper.
// Returns ErrNotFound if the paper does not exist.
func UpdateReadingPosition(db *sql.DB, id string, page int) error {
	result, err := db.Exec(`UPDATE papers SET last_read_page = ? WHERE id = ?`, page, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeletePaper deletes a paper by ID. Returns ErrNotFound if the paper does not exist.
func DeletePaper(db *sql.DB, id string) error {
	result, err := db.Exec(`DELETE FROM papers WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
