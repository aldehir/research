package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestLoggingMiddleware(t *testing.T) {
	t.Run("logs method, path, status, and duration", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		wrapped := requestLogger(logger)(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/papers", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "method=GET")
		assert.Contains(t, logOutput, "path=/api/papers")
		assert.Contains(t, logOutput, "status=200")
		assert.Contains(t, logOutput, "duration")
	})

	t.Run("captures non-200 status codes", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		wrapped := requestLogger(logger)(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/papers/missing", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "status=404")
	})

	t.Run("defaults to 200 if WriteHeader not called", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("implicit 200"))
		})

		wrapped := requestLogger(logger)(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "status=200")
	})
}
