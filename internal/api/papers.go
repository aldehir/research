package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
)

func handleListPapers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		papers, err := store.ListPapers(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list papers")
			return
		}
		if papers == nil {
			papers = []store.Paper{}
		}
		writeJSON(w, http.StatusOK, papers)
	}
}

func handleGetPaper(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		paper, err := store.GetPaper(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper")
			return
		}
		writeJSON(w, http.StatusOK, paper)
	}
}

func handleUploadPaper(db *sql.DB, storage *pdf.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing file field")
			return
		}
		defer file.Close()

		// Read first 4 bytes to check for %PDF magic
		var magic [4]byte
		if _, err := io.ReadFull(file, magic[:]); err != nil {
			writeError(w, http.StatusBadRequest, "file too small to be a PDF")
			return
		}
		if !bytes.Equal(magic[:], []byte("%PDF")) {
			writeError(w, http.StatusBadRequest, "file is not a PDF")
			return
		}

		// Rewind by prepending magic bytes to remaining content
		content := io.MultiReader(bytes.NewReader(magic[:]), file)

		id, err := newUUID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate ID")
			return
		}

		path, size, err := storage.Save(id, content)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save PDF")
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
			writeError(w, http.StatusInternalServerError, "failed to create paper record")
			return
		}

		writeJSON(w, http.StatusCreated, paper)
	}
}

func handleDeletePaper(db *sql.DB, storage *pdf.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// Get paper before deleting to know the file ID
		_, err := store.GetPaper(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper")
			return
		}

		err = store.DeletePaper(db, id)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "paper not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete paper")
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

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
