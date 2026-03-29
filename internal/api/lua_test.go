package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	luaeval "github.com/aldehir/research/internal/lua"
)

func luaMux() *http.ServeMux {
	eval := luaeval.NewEvaluator(5 * time.Second)
	return NewMux(nil, nil, nil, eval, slog.Default())
}

func TestEvalLua_Success(t *testing.T) {
	mux := luaMux()
	body, _ := json.Marshal(map[string]string{"code": `print("hello")`})
	req := httptest.NewRequest(http.MethodPost, "/api/lua/eval", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "hello\n", resp["output"])
	assert.Empty(t, resp["error"])
}

func TestEvalLua_SyntaxError(t *testing.T) {
	mux := luaMux()
	body, _ := json.Marshal(map[string]string{"code": `if then end`})
	req := httptest.NewRequest(http.MethodPost, "/api/lua/eval", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Empty(t, resp["output"])
	assert.NotEmpty(t, resp["error"])
}

func TestEvalLua_Timeout(t *testing.T) {
	eval := luaeval.NewEvaluator(100 * time.Millisecond)
	mux := http.NewServeMux()
	mux.Handle("POST /api/lua/eval", handleEvalLua(eval, slog.Default()))

	body, _ := json.Marshal(map[string]string{"code": `while true do end`})
	req := httptest.NewRequest(http.MethodPost, "/api/lua/eval", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Contains(t, resp["error"], "timeout")
}

func TestEvalLua_EmptyCode(t *testing.T) {
	mux := luaMux()
	body, _ := json.Marshal(map[string]string{"code": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/lua/eval", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEvalLua_MissingBody(t *testing.T) {
	mux := luaMux()
	req := httptest.NewRequest(http.MethodPost, "/api/lua/eval", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
