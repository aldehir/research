package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/aldehir/research/internal/anthropic"
	"github.com/aldehir/research/internal/api"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	dbPath := "research.db"
	if v := os.Getenv("DB_PATH"); v != "" {
		dbPath = v
	}

	pdfDir := "./data/pdfs"
	if v := os.Getenv("PDF_DIR"); v != "" {
		pdfDir = v
	}

	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	var chat api.ChatStreamer
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		chat = anthropic.NewClient(apiKey)
		log.Println("Anthropic API client initialized")
	} else {
		log.Println("WARNING: ANTHROPIC_API_KEY not set, chat features will be unavailable")
	}
	storage := pdf.NewStorage(pdfDir)
	mux := api.NewMux(db, storage, chat)

	serveFrontend(mux)

	fmt.Printf("Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func serveFrontend(mux *http.ServeMux) {
	const buildDir = "frontend/build"

	info, err := os.Stat(buildDir)
	if err != nil || !info.IsDir() {
		log.Printf("Frontend build directory %q not found, skipping static file serving", buildDir)
		return
	}

	frontendFS := os.DirFS(buildDir)
	fileServer := http.FileServerFS(frontendFS)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the exact file first
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = path[1:] // strip leading slash
		}

		// Check if file exists
		if _, err := fs.Stat(frontendFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for any unmatched route
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
