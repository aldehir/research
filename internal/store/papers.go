package store

import (
	"database/sql"
	"errors"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// Paper represents a research paper record.
type Paper struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	CreatedAt string `json:"created_at"`
}

// CreatePaper inserts a new paper record.
func CreatePaper(db *sql.DB, p Paper) error {
	_, err := db.Exec(
		`INSERT INTO papers (id, title, file_path, file_size, created_at) VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.Title, p.FilePath, p.FileSize, p.CreatedAt,
	)
	return err
}

// ListPapers returns all papers ordered by created_at descending.
func ListPapers(db *sql.DB) ([]Paper, error) {
	rows, err := db.Query(`SELECT id, title, file_path, file_size, created_at FROM papers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var papers []Paper
	for rows.Next() {
		var p Paper
		if err := rows.Scan(&p.ID, &p.Title, &p.FilePath, &p.FileSize, &p.CreatedAt); err != nil {
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
		`SELECT id, title, file_path, file_size, created_at FROM papers WHERE id = ?`, id,
	).Scan(&p.ID, &p.Title, &p.FilePath, &p.FileSize, &p.CreatedAt)
	return p, err
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
