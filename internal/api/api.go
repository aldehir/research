package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/aldehir/research/internal/pdf"
)

func NewMux(db *sql.DB, storage *pdf.Storage) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/papers", handleListPapers(db))
	mux.HandleFunc("POST /api/papers", handleUploadPaper(db, storage))
	mux.HandleFunc("GET /api/papers/{id}", handleGetPaper(db))
	mux.HandleFunc("GET /api/papers/{id}/pdf", handleServePDF(db, storage))
	mux.HandleFunc("DELETE /api/papers/{id}", handleDeletePaper(db, storage))
	mux.HandleFunc("GET /api/papers/{id}/chats", handleListChatSessions(db))
	mux.HandleFunc("POST /api/papers/{id}/chats", handleCreateChatSession(db))
	mux.HandleFunc("GET /api/papers/{id}/chats/{chatId}", handleGetChatSession(db))
	mux.HandleFunc("DELETE /api/papers/{id}/chats/{chatId}", handleDeleteChatSession(db))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
