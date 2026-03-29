package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
