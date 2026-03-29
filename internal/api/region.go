package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
)

func handleExtractRegion(db *sql.DB, storage *pdf.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")
		start := time.Now()

		var body struct {
			Page int `json:"page"`
			X    int `json:"x"`
			Y    int `json:"y"`
			W    int `json:"w"`
			H    int `json:"h"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body", logger)
			return
		}

		paper, err := store.GetPaper(db, paperID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper", logger)
			return
		}

		pdfPath := storage.Path(paper.ID)
		logger.Info("extracting region",
			"paper_id", paperID,
			"page", body.Page,
			"x", body.X, "y", body.Y,
			"w", body.W, "h", body.H,
		)

		// Extract text with padding so characters near the selection edge
		// are included. pdftotext clips by character position, which is
		// stricter than pdftoppm's geometric crop.
		const textPad = 15 // points
		tx := max(0, body.X-textPad)
		ty := max(0, body.Y-textPad)
		tw := body.W + 2*textPad
		th := body.H + 2*textPad
		text, err := pdf.ExtractRegionText(pdfPath, body.Page, tx, ty, tw, th)
		if err != nil {
			logger.Warn("region text extraction failed", "error", err)
			text = ""
		}

		// Render image
		pngBytes, err := pdf.RenderRegion(pdfPath, body.Page, body.X, body.Y, body.W, body.H)
		if err != nil {
			logger.Error("region render failed", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to render region", logger)
			return
		}

		imageData := base64.StdEncoding.EncodeToString(pngBytes)

		logger.Info("region extracted",
			"paper_id", paperID,
			"text_length", len(text),
			"image_bytes", len(pngBytes),
			"duration", time.Since(start),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"text":       text,
			"image_data": imageData,
		})
	}
}
