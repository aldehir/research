package store

import (
	"database/sql"
	"testing"
)

// TestDB wraps a *sql.DB for use in tests.
type TestDB struct {
	DB *sql.DB
}

// NewTestDB opens an in-memory SQLite database for testing.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &TestDB{DB: db}
}
