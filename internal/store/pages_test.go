package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertPageText(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	err := UpsertPageText(db, "paper-1", 1, "Hello world")
	require.NoError(t, err)

	text, err := GetPageText(db, "paper-1", 1)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", text)
}

func TestUpsertPageText_UpdateExisting(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	require.NoError(t, UpsertPageText(db, "paper-1", 1, "Original"))
	require.NoError(t, UpsertPageText(db, "paper-1", 1, "Updated"))

	text, err := GetPageText(db, "paper-1", 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated", text)
}

func TestGetPageText_NotFound(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	_, err := GetPageText(db, "paper-1", 99)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSearchPageText(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	require.NoError(t, UpsertPageText(db, "paper-1", 1, "Introduction to quantum computing"))
	require.NoError(t, UpsertPageText(db, "paper-1", 2, "Classical computing background"))
	require.NoError(t, UpsertPageText(db, "paper-1", 3, "Quantum entanglement experiments"))

	results, err := SearchPageText(db, "paper-1", "quantum")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Results ordered by page_num
	assert.Equal(t, 1, results[0].PageNum)
	assert.Equal(t, 3, results[1].PageNum)
}

func TestSearchPageText_NoMatch(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	require.NoError(t, UpsertPageText(db, "paper-1", 1, "Hello world"))

	results, err := SearchPageText(db, "paper-1", "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchPageText_ReturnsSnippet(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	require.NoError(t, UpsertPageText(db, "paper-1", 1, "The quantum computing revolution is changing everything"))

	results, err := SearchPageText(db, "paper-1", "quantum")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Snippet, "quantum")
}

func TestSearchPageText_ScopedToPaper(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	p2 := Paper{ID: "paper-2", Title: "Other", FilePath: "/b.pdf", FileSize: 100, CreatedAt: "2026-01-01T00:00:00Z"}
	require.NoError(t, CreatePaper(db, p2))

	require.NoError(t, UpsertPageText(db, "paper-1", 1, "quantum computing"))
	require.NoError(t, UpsertPageText(db, "paper-2", 1, "quantum physics"))

	results, err := SearchPageText(db, "paper-1", "quantum")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].PageNum)
}

func TestSetTextIndexedAt(t *testing.T) {
	db := testDB(t)
	seedPaper(t, db)

	err := SetTextIndexedAt(db, "paper-1", "2026-03-28T12:00:00Z")
	require.NoError(t, err)

	var ts *string
	err = db.QueryRow("SELECT text_indexed_at FROM papers WHERE id = ?", "paper-1").Scan(&ts)
	require.NoError(t, err)
	require.NotNil(t, ts)
	assert.Equal(t, "2026-03-28T12:00:00Z", *ts)
}

func TestListUnindexedPaperIDs(t *testing.T) {
	db := testDB(t)

	p1 := Paper{ID: "paper-1", Title: "Unindexed", FilePath: "/a.pdf", FileSize: 100, CreatedAt: "2026-01-01T00:00:00Z"}
	require.NoError(t, CreatePaper(db, p1))

	p2 := Paper{ID: "paper-2", Title: "Indexed", FilePath: "/b.pdf", FileSize: 100, CreatedAt: "2026-01-02T00:00:00Z"}
	require.NoError(t, CreatePaper(db, p2))
	require.NoError(t, SetTextIndexedAt(db, "paper-2", "2026-03-28T12:00:00Z"))

	ids, err := ListUnindexedPaperIDs(db)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, "paper-1", ids[0])
}
