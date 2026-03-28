package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func NewMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/papers", handleListPapers(db))
	mux.HandleFunc("GET /api/papers/{id}", handleGetPaper(db))
	mux.HandleFunc("DELETE /api/papers/{id}", handleDeletePaper(db))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
