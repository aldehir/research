package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/aldehir/research/internal/store"
)

func handleGetAttachmentImage(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attID := r.PathValue("id")

		att, err := store.GetAttachment(db, attID)
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "attachment not found", logger)
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to get attachment", logger)
			return
		}

		data, err := os.ReadFile(att.FilePath)
		if err != nil {
			logger.Error("failed to read attachment file", "id", attID, "path", att.FilePath, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read attachment image", logger)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Write(data)
	}
}
