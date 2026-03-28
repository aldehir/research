package store

import (
	"database/sql"
	"log/slog"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database at the given DSN, enables foreign keys,
// and runs schema migrations.
func Open(dsn string, logger ...*slog.Logger) (*sql.DB, error) {
	log := slog.Default()
	if len(logger) > 0 && logger[0] != nil {
		log = logger[0]
	}

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

	log.Info("database opened", "dsn", dsn)
	return db, nil
}

// Migrate runs schema migrations against the database.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	for _, m := range migrations {
		// Ignore errors from ALTER TABLE ADD COLUMN — column may already exist.
		db.Exec(m)
	}
	return nil
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

var migrations = []string{
	`ALTER TABLE papers ADD COLUMN author TEXT`,
	`ALTER TABLE papers ADD COLUMN subject TEXT`,
	`ALTER TABLE papers ADD COLUMN published_date TEXT`,
	`ALTER TABLE papers ADD COLUMN page_count INTEGER`,
}
