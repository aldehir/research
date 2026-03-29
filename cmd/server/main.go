package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aldehir/research/frontend"
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

	logLevel := new(slog.LevelVar)
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if err := logLevel.UnmarshalText([]byte(v)); err != nil {
			slog.Error("invalid LOG_LEVEL", "value", v, "error", err)
			os.Exit(1)
		}
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(handler)

	db, err := store.Open(dbPath, logger)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	var chat api.ChatStreamer
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		var opts []anthropic.Option
		if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
			opts = append(opts, anthropic.WithModel(model))
		}
		client := anthropic.NewClient(apiKey, opts...)
		chat = client
		logger.Info("Anthropic API client initialized", "model", client.Model)
	} else {
		logger.Warn("ANTHROPIC_API_KEY not set, chat features will be unavailable")
	}
	storage := pdf.NewStorage(pdfDir)

	// Start background PDF text indexer
	indexer := pdf.NewIndexer(db, logger)
	go runIndexer(indexer, storage, logger)

	mux := api.NewMux(db, storage, chat, logger)

	frontendFS := resolveFrontendFS(logger)
	if frontendFS != nil {
		serveFrontend(mux, frontendFS)
	}

	logger.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func runIndexer(indexer *pdf.Indexer, storage *pdf.Storage, logger *slog.Logger) {
	// Run once immediately on startup
	if err := indexer.IndexUnindexed(storage.Path); err != nil {
		logger.Warn("indexer run failed", "error", err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := indexer.IndexUnindexed(storage.Path); err != nil {
			logger.Warn("indexer run failed", "error", err)
		}
	}
}

func resolveFrontendFS(logger *slog.Logger) fs.FS {
	if dir := os.Getenv("FRONTEND_DIR"); dir != "" {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			logger.Error("FRONTEND_DIR is not a valid directory", "dir", dir, "error", err)
			os.Exit(1)
		}
		logger.Info("serving frontend from directory", "dir", dir)
		return os.DirFS(dir)
	}

	sub, err := fs.Sub(frontend.BuildFS, "build")
	if err != nil {
		logger.Info("no embedded frontend build, skipping static file serving")
		return nil
	}

	// Verify the embedded FS has content
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		logger.Info("embedded frontend build has no index.html, skipping static file serving")
		return nil
	}

	logger.Info("serving frontend from embedded build")
	return sub
}

func serveFrontend(mux *http.ServeMux, frontendFS fs.FS) {
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
