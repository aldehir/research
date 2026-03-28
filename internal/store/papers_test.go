package store

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndGetPaper(t *testing.T) {
	db := testDB(t)

	paper := Paper{
		ID:        "test-id-1",
		Title:     "Test Paper",
		FilePath:  "/papers/test.pdf",
		FileSize:  12345,
		CreatedAt: "2026-03-28T00:00:00Z",
	}

	err := CreatePaper(db, paper)
	require.NoError(t, err)

	got, err := GetPaper(db, "test-id-1")
	require.NoError(t, err)
	assert.Equal(t, paper, got)
}

func TestListPapers(t *testing.T) {
	db := testDB(t)

	t.Run("empty database returns empty slice", func(t *testing.T) {
		papers, err := ListPapers(db)
		require.NoError(t, err)
		assert.Empty(t, papers)
	})

	t.Run("returns papers ordered by created_at desc", func(t *testing.T) {
		p1 := Paper{ID: "id-1", Title: "Older", FilePath: "/a.pdf", FileSize: 100, CreatedAt: "2026-01-01T00:00:00Z"}
		p2 := Paper{ID: "id-2", Title: "Newer", FilePath: "/b.pdf", FileSize: 200, CreatedAt: "2026-03-01T00:00:00Z"}

		require.NoError(t, CreatePaper(db, p1))
		require.NoError(t, CreatePaper(db, p2))

		papers, err := ListPapers(db)
		require.NoError(t, err)
		require.Len(t, papers, 2)
		assert.Equal(t, "id-2", papers[0].ID)
		assert.Equal(t, "id-1", papers[1].ID)
	})
}

func TestGetPaperNotFound(t *testing.T) {
	db := testDB(t)

	_, err := GetPaper(db, "nonexistent")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeletePaper(t *testing.T) {
	db := testDB(t)

	paper := Paper{
		ID:        "del-id",
		Title:     "To Delete",
		FilePath:  "/del.pdf",
		FileSize:  999,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, CreatePaper(db, paper))

	err := DeletePaper(db, "del-id")
	require.NoError(t, err)

	_, err = GetPaper(db, "del-id")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeletePaperNotFound(t *testing.T) {
	db := testDB(t)

	err := DeletePaper(db, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestCreateAndGetPaper_WithMetadata(t *testing.T) {
	db := testDB(t)

	author := "John Doe"
	subject := "Physics"
	publishedDate := "2025-01-15"
	pageCount := 42

	paper := Paper{
		ID:            "meta-1",
		Title:         "Paper With Metadata",
		FilePath:      "/papers/meta.pdf",
		FileSize:      5000,
		Author:        &author,
		Subject:       &subject,
		PublishedDate: &publishedDate,
		PageCount:     &pageCount,
		CreatedAt:     "2026-03-28T00:00:00Z",
	}

	err := CreatePaper(db, paper)
	require.NoError(t, err)

	got, err := GetPaper(db, "meta-1")
	require.NoError(t, err)
	assert.Equal(t, paper, got)
}

func TestUpdatePaperMetadata(t *testing.T) {
	db := testDB(t)

	paper := Paper{
		ID:        "update-meta-1",
		Title:     "Original",
		FilePath:  "/papers/orig.pdf",
		FileSize:  1000,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, CreatePaper(db, paper))

	author := "Jane Smith"
	pageCount := 10
	err := UpdatePaperMetadata(db, "update-meta-1", PaperMetadata{
		Author:    &author,
		PageCount: &pageCount,
	})
	require.NoError(t, err)

	got, err := GetPaper(db, "update-meta-1")
	require.NoError(t, err)
	require.NotNil(t, got.Author)
	assert.Equal(t, "Jane Smith", *got.Author)
	require.NotNil(t, got.PageCount)
	assert.Equal(t, 10, *got.PageCount)
	assert.Nil(t, got.Subject)
	assert.Nil(t, got.PublishedDate)
}

func TestCreateAndGetPaper_MetadataDefaulsToNil(t *testing.T) {
	db := testDB(t)

	paper := Paper{
		ID:        "no-meta-1",
		Title:     "Paper Without Metadata",
		FilePath:  "/papers/nometa.pdf",
		FileSize:  3000,
		CreatedAt: "2026-03-28T00:00:00Z",
	}

	err := CreatePaper(db, paper)
	require.NoError(t, err)

	got, err := GetPaper(db, "no-meta-1")
	require.NoError(t, err)
	assert.Nil(t, got.Author)
	assert.Nil(t, got.Subject)
	assert.Nil(t, got.PublishedDate)
	assert.Nil(t, got.PageCount)
}
