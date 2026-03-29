package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd_DefaultFlags(t *testing.T) {
	cmd := newRootCmd()

	assert.Equal(t, ":8080", cmd.Flag("addr").DefValue)
	assert.Equal(t, "research.db", cmd.Flag("db-path").DefValue)
	assert.Equal(t, "./data", cmd.Flag("data-dir").DefValue)
	assert.Equal(t, "", cmd.Flag("pdf-dir").DefValue)
	assert.Equal(t, "info", cmd.Flag("log-level").DefValue)
	assert.Equal(t, "", cmd.Flag("frontend-dir").DefValue)
}

func TestNewRootCmd_FlagOverride(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--addr", ":9090", "--db-path", "/tmp/test.db", "--log-level", "debug"})

	// Parse flags without executing the command
	err := cmd.ParseFlags([]string{"--addr", ":9090", "--db-path", "/tmp/test.db", "--log-level", "debug"})
	require.NoError(t, err)

	addr, _ := cmd.Flags().GetString("addr")
	dbPath, _ := cmd.Flags().GetString("db-path")
	logLevel, _ := cmd.Flags().GetString("log-level")

	assert.Equal(t, ":9090", addr)
	assert.Equal(t, "/tmp/test.db", dbPath)
	assert.Equal(t, "debug", logLevel)
}

func TestNewRootCmd_EnvVarFallback(t *testing.T) {
	t.Setenv("ADDR", ":3000")
	t.Setenv("DB_PATH", "/tmp/env.db")
	t.Setenv("DATA_DIR", "/tmp/envdata")
	t.Setenv("PDF_DIR", "/tmp/envpdfs")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("FRONTEND_DIR", "/tmp/frontend")

	cmd := newRootCmd()

	assert.Equal(t, ":3000", cmd.Flag("addr").DefValue)
	assert.Equal(t, "/tmp/env.db", cmd.Flag("db-path").DefValue)
	assert.Equal(t, "/tmp/envdata", cmd.Flag("data-dir").DefValue)
	assert.Equal(t, "/tmp/envpdfs", cmd.Flag("pdf-dir").DefValue)
	assert.Equal(t, "warn", cmd.Flag("log-level").DefValue)
	assert.Equal(t, "/tmp/frontend", cmd.Flag("frontend-dir").DefValue)
}

func TestNewRootCmd_FlagOverridesEnvVar(t *testing.T) {
	t.Setenv("ADDR", ":3000")

	cmd := newRootCmd()
	err := cmd.ParseFlags([]string{"--addr", ":4000"})
	require.NoError(t, err)

	addr, _ := cmd.Flags().GetString("addr")
	assert.Equal(t, ":4000", addr)
}

func TestNewRootCmd_NoAnthropicFlags(t *testing.T) {
	cmd := newRootCmd()

	assert.Nil(t, cmd.Flag("anthropic-api-key"), "ANTHROPIC_API_KEY should not be a flag")
	assert.Nil(t, cmd.Flag("anthropic-model"), "ANTHROPIC_MODEL should not be a flag")
}

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":                {Data: []byte("<html>hello</html>")},
		"_app/immutable/start.js":   {Data: []byte("console.log('start')")},
		"_app/immutable/style.css":  {Data: []byte("body{}")},
		"pdf.worker.min.mjs":        {Data: []byte("// worker")},
	}
}

func TestServeFrontend_ServesIndexHTML(t *testing.T) {
	mux := http.NewServeMux()
	serveFrontend(mux, testFS())

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>hello</html>")
}

func TestServeFrontend_ServesStaticAssets(t *testing.T) {
	mux := http.NewServeMux()
	serveFrontend(mux, testFS())

	tests := []struct {
		path        string
		wantContent string
	}{
		{"/_app/immutable/start.js", "console.log('start')"},
		{"/_app/immutable/style.css", "body{}"},
		{"/pdf.worker.min.mjs", "// worker"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantContent)
		})
	}
}

func TestServeFrontend_SPAFallback(t *testing.T) {
	mux := http.NewServeMux()
	serveFrontend(mux, testFS())

	// Non-existent path should fall back to index.html
	req := httptest.NewRequest("GET", "/papers/some-id", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>hello</html>")
}
