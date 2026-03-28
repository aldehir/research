package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database at the given DSN, enables foreign keys,
// and runs schema migrations.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	if err := Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// Migrate runs schema migrations against the database.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS papers (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	file_path TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chat_sessions (
	id TEXT PRIMARY KEY,
	paper_id TEXT NOT NULL REFERENCES papers(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	chat_session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	selected_text TEXT,
	surrounding_text TEXT,
	created_at TEXT NOT NULL
);
`
