package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMux(t *testing.T) (*http.ServeMux, *store.TestDB, *pdf.Storage) {
	t.Helper()
	tdb := store.NewTestDB(t)
	storage := pdf.NewStorage(filepath.Join(t.TempDir(), "pdfs"))
	mux := NewMux(tdb.DB, storage)
	return mux, tdb, storage
}

func TestListPapers(t *testing.T) {
	t.Run("empty database returns empty array", func(t *testing.T) {
		mux, _, _ := testMux(t)

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
		mux, tdb, _ := testMux(t)

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
		mux, tdb, _ := testMux(t)

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
		mux, _, _ := testMux(t)

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
		mux, tdb, _ := testMux(t)

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
		mux, _, _ := testMux(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/papers/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["error"], "not found")
	})

	t.Run("delete also removes file from disk", func(t *testing.T) {
		mux, _, storage := testMux(t)

		// Upload a PDF first
		pdfContent := []byte("%PDF-1.4 test")
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, err := w.CreateFormFile("file", "removeme.pdf")
		require.NoError(t, err)
		_, err = part.Write(pdfContent)
		require.NoError(t, err)
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/papers", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		var paper store.Paper
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&paper))

		// File should exist
		_, err = os.Stat(storage.Path(paper.ID))
		require.NoError(t, err)

		// Delete
		req = httptest.NewRequest(http.MethodDelete, "/api/papers/"+paper.ID, nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// File should be gone
		_, err = os.Stat(storage.Path(paper.ID))
		assert.True(t, os.IsNotExist(err))
	})
}

func TestServePDF(t *testing.T) {
	t.Run("serves PDF for existing paper", func(t *testing.T) {
		mux, tdb, storage := testMux(t)

		pdfContent := []byte("%PDF-1.4 test content for serving")

		p := store.Paper{ID: "p1", Title: "Serve Test", FilePath: storage.Path("p1"), FileSize: int64(len(pdfContent)), CreatedAt: "2026-03-28T00:00:00Z"}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		// Write the PDF file to storage
		require.NoError(t, os.MkdirAll(filepath.Dir(storage.Path("p1")), 0o755))
		require.NoError(t, os.WriteFile(storage.Path("p1"), pdfContent, 0o644))

		req := httptest.NewRequest(http.MethodGet, "/api/papers/p1/pdf", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
		assert.Equal(t, "inline", rec.Header().Get("Content-Disposition"))
		assert.Equal(t, pdfContent, rec.Body.Bytes())
	})

	t.Run("returns 404 for non-existent paper", func(t *testing.T) {
		mux, _, _ := testMux(t)

		req := httptest.NewRequest(http.MethodGet, "/api/papers/missing/pdf", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["error"], "not found")
	})

	t.Run("returns 404 when paper exists but file missing on disk", func(t *testing.T) {
		mux, tdb, storage := testMux(t)

		p := store.Paper{ID: "p2", Title: "No File", FilePath: storage.Path("p2"), FileSize: 100, CreatedAt: "2026-03-28T00:00:00Z"}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		// Do NOT create the file on disk

		req := httptest.NewRequest(http.MethodGet, "/api/papers/p2/pdf", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["error"], "not found")
	})
}

func TestUploadPaper(t *testing.T) {
	t.Run("valid PDF returns 201 with paper", func(t *testing.T) {
		mux, _, _ := testMux(t)

		pdfContent := []byte("%PDF-1.4 test content here")
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, err := w.CreateFormFile("file", "my-research.pdf")
		require.NoError(t, err)
		_, err = part.Write(pdfContent)
		require.NoError(t, err)
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/papers", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var paper store.Paper
		err = json.NewDecoder(rec.Body).Decode(&paper)
		require.NoError(t, err)
		assert.NotEmpty(t, paper.ID)
		assert.Equal(t, "my-research", paper.Title)
		assert.Equal(t, int64(len(pdfContent)), paper.FileSize)
		assert.NotEmpty(t, paper.FilePath)
		assert.NotEmpty(t, paper.CreatedAt)
	})

	t.Run("non-PDF file returns 400", func(t *testing.T) {
		mux, _, _ := testMux(t)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, err := w.CreateFormFile("file", "notapdf.txt")
		require.NoError(t, err)
		_, err = part.Write([]byte("this is not a PDF"))
		require.NoError(t, err)
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/papers", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var body map[string]string
		err = json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body["error"], "PDF")
	})
}
