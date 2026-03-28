package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteError_Logs5xxAsError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	rec := httptest.NewRecorder()
	writeError(rec, http.StatusInternalServerError, "something broke", logger)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))
	assert.Equal(t, "ERROR", logEntry["level"])
	assert.Equal(t, "something broke", logEntry["msg"])
	assert.Equal(t, float64(500), logEntry["status"])
}

func TestWriteError_Logs4xxAsWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	rec := httptest.NewRecorder()
	writeError(rec, http.StatusNotFound, "paper not found", logger)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))
	assert.Equal(t, "WARN", logEntry["level"])
	assert.Equal(t, "paper not found", logEntry["msg"])
	assert.Equal(t, float64(404), logEntry["status"])
}
