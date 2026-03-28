package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
)

func handleListPapers(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		papers, err := store.ListPapers(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list papers", logger)
			return
		}
		if papers == nil {
			papers = []store.Paper{}
		}
		writeJSON(w, http.StatusOK, papers)
	}
}

func handleGetPaper(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		paper, err := store.GetPaper(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper", logger)
			return
		}
		writeJSON(w, http.StatusOK, paper)
	}
}

func handleUploadPaper(db *sql.DB, storage *pdf.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing file field", logger)
			return
		}
		defer file.Close()

		// Read first 4 bytes to check for %PDF magic
		var magic [4]byte
		if _, err := io.ReadFull(file, magic[:]); err != nil {
			writeError(w, http.StatusBadRequest, "file too small to be a PDF", logger)
			return
		}
		if !bytes.Equal(magic[:], []byte("%PDF")) {
			writeError(w, http.StatusBadRequest, "file is not a PDF", logger)
			return
		}

		// Rewind by prepending magic bytes to remaining content
		content := io.MultiReader(bytes.NewReader(magic[:]), file)

		id, err := newUUID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate ID", logger)
			return
		}

		path, size, err := storage.Save(id, content)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save PDF", logger)
			return
		}

		// Derive title from filename without .pdf extension
		title := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))

		paper := store.Paper{
			ID:        id,
			Title:     title,
			FilePath:  path,
			FileSize:  size,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}

		if err := store.CreatePaper(db, paper); err != nil {
			storage.Delete(id)
			writeError(w, http.StatusInternalServerError, "failed to create paper record", logger)
			return
		}

		writeJSON(w, http.StatusCreated, paper)
	}
}

func handleServePDF(db *sql.DB, storage *pdf.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, err := store.GetPaper(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper", logger)
			return
		}

		path := storage.Path(id)
		f, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "file not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to open file", logger)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline")
		io.Copy(w, f)
	}
}

func handleDeletePaper(db *sql.DB, storage *pdf.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// Get paper before deleting to know the file ID
		_, err := store.GetPaper(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper", logger)
			return
		}

		err = store.DeletePaper(db, id)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete paper", logger)
			return
		}

		// Best-effort file deletion
		storage.Delete(id)

		w.WriteHeader(http.StatusNoContent)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, logger *slog.Logger) {
	if status >= 500 {
		logger.Error(msg, "status", status)
	} else {
		logger.Warn(msg, "status", status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
