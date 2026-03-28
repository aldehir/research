package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

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

func handleDeletePaper(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		err := store.DeletePaper(db, id)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "paper not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete paper")
			return
		}
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
