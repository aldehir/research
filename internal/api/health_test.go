package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	mux := NewMux(nil, nil, nil, nil, slog.Default())

	t.Run("returns 200 with status ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "ok", body["status"])
	})
}
