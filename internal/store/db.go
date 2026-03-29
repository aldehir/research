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
	created_at TEXT NOT NULL
);
`

var migrations = []string{
	`ALTER TABLE papers ADD COLUMN author TEXT`,
	`ALTER TABLE papers ADD COLUMN subject TEXT`,
	`ALTER TABLE papers ADD COLUMN published_date TEXT`,
	`ALTER TABLE papers ADD COLUMN page_count INTEGER`,
	`ALTER TABLE papers ADD COLUMN text_indexed_at TEXT`,
	`CREATE TABLE IF NOT EXISTS paper_pages (
		id TEXT PRIMARY KEY,
		paper_id TEXT NOT NULL REFERENCES papers(id) ON DELETE CASCADE,
		page_num INTEGER NOT NULL,
		text_content TEXT NOT NULL,
		UNIQUE(paper_id, page_num)
	)`,
	`CREATE VIRTUAL TABLE IF NOT EXISTS paper_pages_fts USING fts5(
		text_content,
		content=paper_pages,
		content_rowid=rowid
	)`,
	`ALTER TABLE papers ADD COLUMN last_read_page INTEGER`,
	// Triggers to keep FTS5 index in sync with paper_pages
	`CREATE TRIGGER IF NOT EXISTS paper_pages_ai AFTER INSERT ON paper_pages BEGIN
		INSERT INTO paper_pages_fts(rowid, text_content) VALUES (new.rowid, new.text_content);
	END`,
	`CREATE TRIGGER IF NOT EXISTS paper_pages_ad AFTER DELETE ON paper_pages BEGIN
		INSERT INTO paper_pages_fts(paper_pages_fts, rowid, text_content) VALUES ('delete', old.rowid, old.text_content);
	END`,
	`CREATE TRIGGER IF NOT EXISTS paper_pages_au AFTER UPDATE ON paper_pages BEGIN
		INSERT INTO paper_pages_fts(paper_pages_fts, rowid, text_content) VALUES ('delete', old.rowid, old.text_content);
		INSERT INTO paper_pages_fts(rowid, text_content) VALUES (new.rowid, new.text_content);
	END`,
}
