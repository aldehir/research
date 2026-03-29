package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractRegion(t *testing.T) {
	t.Run("returns text and image for a valid region", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Hello World")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
			CreatedAt: "2026-03-28T00:00:00Z",
		}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		mux := NewMux(tdb.DB, storage, nil, nil, slog.Default())

		body := `{"page":1,"x":0,"y":0,"w":300,"h":100}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/region", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result struct {
			Text      string `json:"text"`
			ImageData string `json:"image_data"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
		assert.Contains(t, result.Text, "Hello")
		assert.NotEmpty(t, result.ImageData, "should return base64 image data")
	})

	t.Run("paper not found returns 404", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())
		mux := NewMux(tdb.DB, storage, nil, nil, slog.Default())

		body := `{"page":1,"x":0,"y":0,"w":100,"h":100}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/nonexistent/region", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())
		mux := NewMux(tdb.DB, storage, nil, nil, slog.Default())

		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/region", strings.NewReader("not json"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("region with no text still returns image", func(t *testing.T) {
		tdb := store.NewTestDB(t)
		storage := pdf.NewStorage(t.TempDir())

		pdfPath := storage.Path("paper-1")
		createTestPDFWithText(t, pdfPath, "Hello World")

		pageCount := 1
		p := store.Paper{
			ID:        "paper-1",
			Title:     "Test Paper",
			FilePath:  pdfPath,
			FileSize:  1000,
			PageCount: &pageCount,
			CreatedAt: "2026-03-28T00:00:00Z",
		}
		require.NoError(t, store.CreatePaper(tdb.DB, p))

		mux := NewMux(tdb.DB, storage, nil, nil, slog.Default())

		// Small region far from text
		body := `{"page":1,"x":500,"y":500,"w":10,"h":10}`
		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/region", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result struct {
			Text      string `json:"text"`
			ImageData string `json:"image_data"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
		assert.NotEmpty(t, result.ImageData, "should still return image even with no text")
	})
}
