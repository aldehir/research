package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	luaeval "github.com/aldehir/research/internal/lua"
	"github.com/aldehir/research/internal/pdf"
)

func NewMux(db *sql.DB, storage *pdf.Storage, chat ChatStreamer, luaEval *luaeval.Evaluator, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	wrap := requestLogger(logger)
	mux.Handle("GET /api/health", wrap(http.HandlerFunc(handleHealth)))
	mux.Handle("GET /api/papers", wrap(handleListPapers(db, logger)))
	mux.Handle("POST /api/papers", wrap(handleUploadPaper(db, storage, logger)))
	mux.Handle("GET /api/papers/{id}", wrap(handleGetPaper(db, logger)))
	mux.Handle("GET /api/papers/{id}/pdf", wrap(handleServePDF(db, storage, logger)))
	mux.Handle("DELETE /api/papers/{id}", wrap(handleDeletePaper(db, storage, logger)))
	mux.Handle("PATCH /api/papers/{id}/position", wrap(handleUpdateReadingPosition(db, logger)))
	mux.Handle("GET /api/papers/{id}/chats", wrap(handleListChatSessions(db, logger)))
	mux.Handle("POST /api/papers/{id}/chats", wrap(handleCreateChatSession(db, logger)))
	mux.Handle("GET /api/papers/{id}/chats/{chatId}", wrap(handleGetChatSession(db, logger)))
	mux.Handle("DELETE /api/papers/{id}/chats/{chatId}", wrap(handleDeleteChatSession(db, logger)))
	mux.Handle("POST /api/papers/{id}/chats/{chatId}/messages", wrap(handleSendMessage(db, storage, chat, logger)))
	if luaEval != nil {
		mux.Handle("POST /api/lua/eval", wrap(handleEvalLua(luaEval, logger)))
	}
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
