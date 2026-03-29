package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aldehir/research/frontend"
	"github.com/aldehir/research/internal/anthropic"
	"github.com/aldehir/research/internal/api"
	"github.com/aldehir/research/internal/chat"
	luaeval "github.com/aldehir/research/internal/lua"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
	"github.com/spf13/cobra"
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "research-server",
		Short: "Research paper reader server",
		RunE:  runServe,
	}

	f := cmd.Flags()
	f.String("addr", envOrDefault("ADDR", ":8080"), "listen address")
	f.String("db-path", envOrDefault("DB_PATH", "research.db"), "SQLite database path")
	f.String("data-dir", envOrDefault("DATA_DIR", "./data"), "data directory for PDFs and attachments")
	f.String("pdf-dir", envOrDefault("PDF_DIR", ""), "PDF storage directory (default: {data-dir}/pdfs)")
	f.String("log-level", envOrDefault("LOG_LEVEL", "info"), "log level (debug, info, warn, error)")
	f.String("frontend-dir", envOrDefault("FRONTEND_DIR", ""), "serve frontend from this directory instead of embedded build")

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	addr, _ := cmd.Flags().GetString("addr")
	dbPath, _ := cmd.Flags().GetString("db-path")
	dataDir, _ := cmd.Flags().GetString("data-dir")
	pdfDir, _ := cmd.Flags().GetString("pdf-dir")
	logLevelStr, _ := cmd.Flags().GetString("log-level")
	frontendDir, _ := cmd.Flags().GetString("frontend-dir")

	if pdfDir == "" {
		pdfDir = filepath.Join(dataDir, "pdfs")
	}

	logLevel := new(slog.LevelVar)
	if err := logLevel.UnmarshalText([]byte(logLevelStr)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", logLevelStr, err)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(handler)

	db, err := store.Open(dbPath, logger)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	var provider chat.Provider
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		var opts []anthropic.Option
		if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
			opts = append(opts, anthropic.WithModel(model))
		}
		if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(baseURL))
		}
		client := anthropic.NewClient(apiKey, opts...)
		provider = anthropic.NewAdapter(client)
		logger.Info("Anthropic API client initialized", "model", client.Model, "base_url", client.BaseURL)
	} else {
		logger.Warn("ANTHROPIC_API_KEY not set, chat features will be unavailable")
	}
	storage := pdf.NewStorage(pdfDir)

	indexer := pdf.NewIndexer(db, logger)
	go runIndexer(indexer, storage, logger)

	luaEval := luaeval.NewEvaluator(5 * time.Second)
	mux := api.NewMux(db, storage, provider, luaEval, logger, api.WithDataDir(dataDir))

	frontendFS := resolveFrontendFS(frontendDir, logger)
	if frontendFS != nil {
		serveFrontend(mux, frontendFS)
	}

	logger.Info("server starting", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

func runIndexer(indexer *pdf.Indexer, storage *pdf.Storage, logger *slog.Logger) {
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

func resolveFrontendFS(dir string, logger *slog.Logger) fs.FS {
	if dir != "" {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			logger.Error("frontend-dir is not a valid directory", "dir", dir, "error", err)
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
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = path[1:]
		}

		if _, err := fs.Stat(frontendFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
