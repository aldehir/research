package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/aldehir/research/internal/chat"
	luaeval "github.com/aldehir/research/internal/lua"
	"github.com/aldehir/research/internal/pdf"
)

// MuxOption configures the API mux.
type MuxOption func(*muxConfig)

type muxConfig struct {
	dataDir      string
	retentionTTL time.Duration
}

// WithDataDir sets the data directory for storing attachment images.
func WithDataDir(dir string) MuxOption {
	return func(c *muxConfig) { c.dataDir = dir }
}

// WithRetentionTTL sets how long completed streams are retained for reconnection.
func WithRetentionTTL(d time.Duration) MuxOption {
	return func(c *muxConfig) { c.retentionTTL = d }
}

func NewMux(db *sql.DB, storage *pdf.Storage, provider chat.Provider, luaEval *luaeval.Evaluator, logger *slog.Logger, opts ...MuxOption) *http.ServeMux {
	var cfg muxConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	registry := NewStreamRegistry(logger)
	if cfg.retentionTTL > 0 {
		registry.RetentionTTL = cfg.retentionTTL
	}

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
	mux.Handle("POST /api/papers/{id}/region", wrap(handleExtractRegion(db, storage, logger)))
	mux.Handle("POST /api/papers/{id}/chats/{chatId}/messages", wrap(handleSendMessage(db, storage, provider, cfg.dataDir, registry, logger)))
	mux.Handle("GET /api/papers/{id}/chats/{chatId}/stream", wrap(handleReconnectStream(registry, logger)))
	mux.Handle("GET /api/attachments/{id}/image", wrap(handleGetAttachmentImage(db, logger)))
	if luaEval != nil {
		mux.Handle("POST /api/lua/eval", wrap(handleEvalLua(luaEval, logger)))
	}
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
