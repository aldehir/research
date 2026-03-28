package store_test

import (
	"database/sql"
	"testing"

	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_CreatesTablesOnMigration(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	tables := []string{"papers", "chat_sessions", "messages"}
	for _, table := range tables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		assert.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}
}

func TestOpen_MigrationIsIdempotent(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Run migration again by calling Migrate directly
	err = store.Migrate(db)
	assert.NoError(t, err)
}

func TestOpen_ForeignKeysEnabled(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled)
}

func TestCascadeDelete_PaperDeleteRemovesSessions(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Insert a paper
	_, err = db.Exec(
		"INSERT INTO papers (id, title, file_path, file_size, created_at) VALUES (?, ?, ?, ?, ?)",
		"paper-1", "Test Paper", "/path/to/file.pdf", 1024, "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	// Insert a chat session
	_, err = db.Exec(
		"INSERT INTO chat_sessions (id, paper_id, title, created_at) VALUES (?, ?, ?, ?)",
		"session-1", "paper-1", "Chat 1", "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	// Insert a message
	_, err = db.Exec(
		"INSERT INTO messages (id, chat_session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)",
		"msg-1", "session-1", "user", "Hello", "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	// Delete the paper
	_, err = db.Exec("DELETE FROM papers WHERE id = ?", "paper-1")
	require.NoError(t, err)

	// Session should be gone
	assertRowCount(t, db, "chat_sessions", 0)
	// Message should be gone too (cascaded through session)
	assertRowCount(t, db, "messages", 0)
}

func TestCascadeDelete_SessionDeleteRemovesMessages(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO papers (id, title, file_path, file_size, created_at) VALUES (?, ?, ?, ?, ?)",
		"paper-1", "Test Paper", "/path/to/file.pdf", 1024, "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO chat_sessions (id, paper_id, title, created_at) VALUES (?, ?, ?, ?)",
		"session-1", "paper-1", "Chat 1", "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO messages (id, chat_session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)",
		"msg-1", "session-1", "user", "Hello", "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	// Delete the session
	_, err = db.Exec("DELETE FROM chat_sessions WHERE id = ?", "session-1")
	require.NoError(t, err)

	// Message should be gone
	assertRowCount(t, db, "messages", 0)
	// Paper should still exist
	assertRowCount(t, db, "papers", 1)
}

func TestOpen_CreatesPaperPagesTable(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// paper_pages table should exist
	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='paper_pages'",
	).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "paper_pages", name)
}

func TestOpen_CreatesPaperPagesFTSTable(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// paper_pages_fts virtual table should exist
	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='paper_pages_fts'",
	).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "paper_pages_fts", name)
}

func TestOpen_PapersHasTextIndexedAtColumn(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// text_indexed_at column should exist on papers
	_, err = db.Exec("SELECT text_indexed_at FROM papers LIMIT 0")
	assert.NoError(t, err)
}

func TestCascadeDelete_PaperDeleteRemovesPages(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO papers (id, title, file_path, file_size, created_at) VALUES (?, ?, ?, ?, ?)",
		"paper-1", "Test Paper", "/path/to/file.pdf", 1024, "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO paper_pages (id, paper_id, page_num, text_content) VALUES (?, ?, ?, ?)",
		"page-1", "paper-1", 1, "Hello world",
	)
	require.NoError(t, err)

	// Delete the paper
	_, err = db.Exec("DELETE FROM papers WHERE id = ?", "paper-1")
	require.NoError(t, err)

	// paper_pages should be gone
	assertRowCount(t, db, "paper_pages", 0)
}

func TestCascadeDelete_PaperDeleteCleansFTS(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO papers (id, title, file_path, file_size, created_at) VALUES (?, ?, ?, ?, ?)",
		"paper-1", "Test Paper", "/path/to/file.pdf", 1024, "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	require.NoError(t, store.UpsertPageText(db, "paper-1", 1, "quantum computing is great"))

	// Verify FTS has data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM paper_pages_fts WHERE paper_pages_fts MATCH 'quantum'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Delete the paper
	_, err = db.Exec("DELETE FROM papers WHERE id = ?", "paper-1")
	require.NoError(t, err)

	// FTS should be empty
	err = db.QueryRow("SELECT COUNT(*) FROM paper_pages_fts WHERE paper_pages_fts MATCH 'quantum'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func assertRowCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, expected, count, "expected %d rows in %s", expected, table)
}
