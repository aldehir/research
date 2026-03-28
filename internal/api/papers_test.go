package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMux(t *testing.T) (*http.ServeMux, *store.TestDB) {
	t.Helper()
	tdb := store.NewTestDB(t)
	mux := NewMux(tdb.DB)
	return mux, tdb
}

func TestListPapers(t *testing.T) {
	t.Run("empty database returns empty array", func(t *testing.T) {
		mux, _ := testMux(t)

		req := httptest.NewRequest(http.MethodGet, "/api/papers", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var papers []store.Paper
		err := json.NewDecoder(rec.Body).Decode(&papers)
		require.NoError(t, err)
		assert.Empty(t, papers)
	})

	t.Run("returns papers", func(t *testing.T) {
		mux, tdb := testMux(t)

		p := store.Paper{ID: "p1", Title: "A Paper", FilePath: "/a.pdf", FileSize: 100, CreatedAt: "2026-03-28T00:00:00Z"}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		req := httptest.NewRequest(http.MethodGet, "/api/papers", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var papers []store.Paper
		err := json.NewDecoder(rec.Body).Decode(&papers)
		require.NoError(t, err)
		require.Len(t, papers, 1)
		assert.Equal(t, "p1", papers[0].ID)
	})
}

func TestGetPaper(t *testing.T) {
	t.Run("existing paper returns 200", func(t *testing.T) {
		mux, tdb := testMux(t)

		p := store.Paper{ID: "p1", Title: "A Paper", FilePath: "/a.pdf", FileSize: 100, CreatedAt: "2026-03-28T00:00:00Z"}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		req := httptest.NewRequest(http.MethodGet, "/api/papers/p1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got store.Paper
		err := json.NewDecoder(rec.Body).Decode(&got)
		require.NoError(t, err)
		assert.Equal(t, "p1", got.ID)
		assert.Equal(t, "A Paper", got.Title)
	})

	t.Run("non-existent paper returns 404", func(t *testing.T) {
		mux, _ := testMux(t)

		req := httptest.NewRequest(http.MethodGet, "/api/papers/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["error"], "not found")
	})
}

func TestDeletePaper(t *testing.T) {
	t.Run("delete existing paper returns 204", func(t *testing.T) {
		mux, tdb := testMux(t)

		p := store.Paper{ID: "p1", Title: "A Paper", FilePath: "/a.pdf", FileSize: 100, CreatedAt: "2026-03-28T00:00:00Z"}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		req := httptest.NewRequest(http.MethodDelete, "/api/papers/p1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Confirm it's gone
		req = httptest.NewRequest(http.MethodGet, "/api/papers/p1", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("delete non-existent paper returns 404", func(t *testing.T) {
		mux, _ := testMux(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/papers/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["error"], "not found")
	})
}
